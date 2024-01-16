package localdump

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/toothbrush/confluence-dump/confluence"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
)

type SpacesDownloader struct {
	StorePath       string
	Workers         int
	API             *confluence.API
	AlwaysDownload  bool
	WriteMarkdown   bool
	Prune           bool
	IncludeArchived bool

	Logger   *log.Logger
	loggerMu sync.Mutex

	// spaces metadata
	spacesMetadata map[string]confluence.Space

	// local markdown:
	localMarkdownCache map[ContentID]LocalMarkdown

	// place to store results
	remotePageMetadata map[ContentID]RemoteObjectMetadata
	remoteMetadataMu   sync.Mutex

	freshLocalFiles map[string]bool

	authorMetadata map[string]confluence.User
}

type JobType int8

const (
	PagesList JobType = iota
	PageFetch
	UserFetch
)

type Job struct {
	JobType JobType
	retries int

	// If fetching PagesList:
	// This makes us fetch a space by ID or just "blogposts"
	GetPagesQuery confluence.GetPagesQuery
	Org           string
	SpaceKey      string
	// SpaceID       string

	// Or, if PageFetch:
	PageID      string
	ContentType confluence.ContentType
	// ... need Query

	// Or, if UserFetch:
	GetUserQuery confluence.GetUserByIDQuery
}

func (downloader *SpacesDownloader) DownloadConfluenceSpaces(ctx context.Context, spaces []confluence.Space) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	downloader.spacesMetadata = make(map[string]confluence.Space)
	for _, s := range spaces {
		downloader.spacesMetadata[s.ID] = s
	}

	// first, load up local markdown database:
	downloader.Logger.Println("Loading local Markdown files, if any...")
	if err := downloader.LoadLocalMarkdown(); err != nil {
		return fmt.Errorf("localdump: failed to load local Markdown: %w", err)
	}
	downloader.Logger.Printf("...loaded %d Markdown files.\n", len(downloader.localMarkdownCache))

	// less first, determine entire list of pages in the spaces the user wants:
	downloader.Logger.Printf("Listing pages in %d spaces...\n", len(downloader.spacesMetadata))
	listPagesInSpacesJobs, err := downloader.generatePageListJobs(ctx)
	if err != nil {
		return fmt.Errorf("localdump: couldn't generate page-list jobs: %w", err)
	}

	if err := downloader.channelSoupRun(ctx, listPagesInSpacesJobs, downloader.Workers*100, "spaces"); err != nil {
		return fmt.Errorf("localdump: failed to channelsoup: %w", err)
	}
	downloader.Logger.Printf("...found %d total pages across %d spaces\n",
		len(downloader.remotePageMetadata),
		len(downloader.spacesMetadata))

	// set up ancestry cache for quick staleness check:
	if err := downloader.BuildCacheFromPagelist(); err != nil {
		return fmt.Errorf("localdump: failed to resolve all ancestry: %w", err)
	}

	// grab list of all users we've ever seen...
	downloader.Logger.Println("Fetching user metadata...")
	userJobs, err := downloader.generateUserFetchJobs(ctx)
	if err != nil {
		return fmt.Errorf("localdump: couldn't generate user-fetch jobs: %w", err)
	}
	if err := downloader.channelSoupRun(ctx, userJobs, len(userJobs), "users"); err != nil {
		return fmt.Errorf("localdump: failed to channelsoup: %w", err)
	}
	downloader.Logger.Printf("...refreshed %d total users.\n",
		len(downloader.authorMetadata))

	// This is a get-single-page type channelsoup:
	downloader.Logger.Println("Fetching pages...")
	pageJobs, err := downloader.generateSinglePageDownloadJobs(ctx)
	if err != nil {
		return fmt.Errorf("localdump: couldn't generate single-page jobs: %w", err)
	}
	// next, download pages that are stored in downloader.remotePageMetadata:
	if err := downloader.channelSoupRun(ctx, pageJobs, len(pageJobs), "pages"); err != nil {
		return fmt.Errorf("localdump: failed to channelsoup: %w", err)
	}
	downloader.Logger.Println("...done fetching pages.")

	if downloader.WriteMarkdown && downloader.Prune {
		// finally, prune local Markdown database:
		if err := downloader.pruneLocalDB(); err != nil {
			return fmt.Errorf("localdump: failed to prune: %w", err)
		}
		// TODO more detail
		downloader.Logger.Println("...done pruning pages.")
	}

	return nil
}

func (downloader *SpacesDownloader) generateUserFetchJobs(ctx context.Context) ([]Job, error) {
	jobs := make(map[string]Job) // to weed out dupes
	for _, s := range downloader.remotePageMetadata {
		id := s.Page.AuthorID

		if _, ok := jobs[id]; ok {
			// already exists
			continue
		}

		jobs[id] = Job{
			JobType: UserFetch,
			Org:     s.Page.Org,
			GetUserQuery: confluence.GetUserByIDQuery{
				ID: id,
			},
		}
	}
	return maps.Values(jobs), nil
}

func (downloader *SpacesDownloader) generatePageListJobs(ctx context.Context) ([]Job, error) {
	jobs := []Job{}
	for _, s := range downloader.spacesMetadata {
		var query confluence.GetPagesQuery
		query.Status = []string{"current"}

		if downloader.IncludeArchived {
			query.Status = append(query.Status, "archived")
		}

		if s.Key == "blogposts" {
			query.QueryType = confluence.BlogContent
		} else {
			// create initial PagesQuery, and pop it in the job queue.
			id, err := strconv.Atoi(s.ID)
			if err != nil {
				return nil, fmt.Errorf("localdump(%s): id was not an int: %w", s.Key, err)
			}
			query.QueryType = confluence.PageContent
			query.SpaceID = []int{id}
		}

		pagesListJob := Job{
			JobType:       PagesList,
			Org:           s.Org,
			SpaceKey:      s.Key,
			GetPagesQuery: query,
		}

		if s.Key == "blogposts" {
			pagesListJob.ContentType = confluence.BlogContent
		} else {
			pagesListJob.ContentType = confluence.PageContent
		}

		jobs = append(jobs, pagesListJob)
	}
	return jobs, nil
}

func (downloader *SpacesDownloader) generateSinglePageDownloadJobs(ctx context.Context) ([]Job, error) {
	jobs := []Job{}

	for _, p := range downloader.remotePageMetadata {
		// create initial PageQuery, and pop it in the job queue.
		// figure out space key this page belongs to:
		spaceID := p.Page.SpaceID
		if p.Page.ContentType == confluence.BlogContent {
			spaceID = "blogposts"
		}
		space, ok := downloader.spacesMetadata[spaceID]
		if !ok {
			return nil, fmt.Errorf("localdump: space id unknown: %s", p.Page.SpaceID)
		}

		pageDownloadJob := Job{
			JobType:     PageFetch,
			PageID:      p.Page.ID,
			Org:         space.Org,
			SpaceKey:    space.Key,
			ContentType: p.Page.ContentType,
		}

		jobs = append(jobs, pageDownloadJob)
	}

	return jobs, nil
}

func (downloader *SpacesDownloader) performJob(ctx context.Context, job Job) (JobResult, error) {
	switch job.JobType {
	case PagesList:
		listResult, err := downloader.performPageListJob(ctx, job)
		if err != nil {
			return JobResult{}, fmt.Errorf("downloader: Confluence download failed: %w", err)
		}
		return listResult, nil

	case PageFetch:
		pageResult, err := downloader.performPageDownloadJob(ctx, job)
		if err != nil {
			return JobResult{}, fmt.Errorf("downloader: Confluence download failed: %w", err)
		}
		// update freshLocalFiles
		downloader.remoteMetadataMu.Lock()
		defer downloader.remoteMetadataMu.Unlock()
		if downloader.freshLocalFiles == nil {
			downloader.freshLocalFiles = make(map[string]bool)
		}
		downloader.freshLocalFiles[string(pageResult.page.RelativePath)] = true
		return pageResult, nil

	case UserFetch:
		userResult, err := downloader.performUserDownloadJob(ctx, job)
		if err != nil {
			return JobResult{}, fmt.Errorf("downloader: Confluence download failed: %w", err)
		}
		return userResult, nil

	default:
		return JobResult{}, fmt.Errorf("downloader: unreachable case jobType = %d", job.JobType)
	}
}

func (downloader *SpacesDownloader) channelSoupRun(ctx context.Context, jobs []Job, chanBufferSize int, phaseName string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobQueue := make(chan Job, chanBufferSize)

	// enqueue the work:
	unitsOfWorkRemaining := int32(len(jobs))
	unitsOfWorkTotal := len(jobs)
	for _, j := range jobs {
		select {
		case jobQueue <- j:
		case <-ctx.Done():
			return context.Cause(ctx)
		}
	}
	// END enqueue the work

	results := make(chan JobResult, downloader.Workers*3)

	grp, gctx := errgroup.WithContext(ctx)

	workers := int32(downloader.Workers)
	for i := 0; i < downloader.Workers; i++ {
		grp.Go(func() error {
			for {
				select {
				case job, ok := <-jobQueue:
					if !ok {
						// input channel closed
						// Last one out closes the shop
						if atomic.AddInt32(&workers, -1) == 0 {
							close(results)
						}
						return nil
					}
					result, err := downloader.performJob(ctx, job)
					// at this point we would need to decide what kind of error we have
					// (instant-stop or transient)
					//
					// currently we're insta-stopping on any error.
					if err != nil {
						if job.retries > 3 {
							return fmt.Errorf("downloader.performJob: retries exceeded: %w", err)
						}
						job.retries++
						fmt.Printf("hit error %d on %v\n", job.retries, job)
						result = JobResult{
							JobType:     job.JobType,
							followUpJob: &job,
						}
					}
					if result.followUpJob != nil {
						// enqueue resulting job, if there is one!
						select {
						case jobQueue <- *result.followUpJob:
						case <-gctx.Done():
							return context.Cause(ctx)
						}
					} else {
						// this space is finished! (that is, the job didn't return a new job to run.)
						if atomic.AddInt32(&unitsOfWorkRemaining, -1) == 0 {
							// only do this when we're sure there's no more work:
							close(jobQueue)
						}
					}

					select {
					case results <- result:
					case <-gctx.Done():
						return context.Cause(ctx)
					}

				case <-gctx.Done():
					return context.Cause(ctx)
				}
			}
		})
	}
	p := mpb.New(mpb.WithWidth(64))

	bar := p.AddBar(int64(unitsOfWorkTotal),
		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(fmt.Sprintf("%s:", phaseName),
				decor.WC{C: decor.DindentRight | decor.DextraSpace}),
		),
		mpb.AppendDecorators(
			decor.CountersNoUnit("(%d/%d) "), // , wcc ...decor.WC)
			decor.NewPercentage("%d"),
			decor.Spinner([]string{" /", " -", " \\", " |"}),
		),
	)

	// print our results
	grp.Go(func() error {
		for {
			select {
			case result, ok := <-results:
				if !ok {
					// this is good news, we're done
					return nil
				}
				// ok means the channel isn't closed yet

				if result.finished {
					bar.Increment()
				}
				// downloader.printJobResults(result, unitsOfWorkRemaining, unitsOfWorkTotal)

			case <-gctx.Done():
				return context.Cause(ctx)
			}
		}
	})

	// Wait for all workers to return:
	if err := grp.Wait(); err != nil {
		return fmt.Errorf("localdump: failure: %w", err)
	}

	// wait for our bar to complete and flush
	p.Wait()

	return nil
}

func (downloader *SpacesDownloader) printJobResults(result JobResult, remaining int32, total int) {
	downloader.loggerMu.Lock()
	switch result.JobType {
	case PagesList:
		if result.finished {
			downloader.Logger.Printf("Listed space %s.\n", result.space)
		}
	case PageFetch:
		if result.pageDownloadOutcome == SkippedCached {
			downloader.Logger.Printf("(v%2d cached): %s\n", result.page.Version, result.page.RelativePath)
		} else {
			downloader.Logger.Printf("Fetched: %s\n", result.page.RelativePath)
		}
	case UserFetch:
		downloader.Logger.Printf("Fetched user: %s\n", result.user.Email)
	}
	downloader.loggerMu.Unlock()
}

func (downloader *SpacesDownloader) getPagesOrBlogs(ctx context.Context, job Job) (*confluence.MultiPageResponse, error) {
	if job.GetPagesQuery.QueryType == confluence.PageContent {
		result, err := downloader.API.GetPages(ctx, job.GetPagesQuery)
		return result, err
	} else {
		result, err := downloader.API.GetBlogPosts(ctx, job.GetPagesQuery)
		return result, err
	}
}

type JobResult struct {
	JobType JobType
	space   string

	// These fields are for listing-pages jobs:
	finished    bool
	itemsFound  int
	followUpJob *Job

	// These fields are for page-download jobs:
	pageDownloadOutcome DownloadAction
	page                *LocalMarkdown

	// Field for user-fetch job:
	user confluence.User
}

func (downloader *SpacesDownloader) performPageListJob(ctx context.Context, job Job) (JobResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	apiResult, err := downloader.getPagesOrBlogs(ctx, job)
	if err != nil {
		return JobResult{}, fmt.Errorf("localdump: failed getting partial page list: %w", err)
	}

	// eek what if we early return, probably defer unlock?
	downloader.remoteMetadataMu.Lock()
	defer downloader.remoteMetadataMu.Unlock()

	if downloader.remotePageMetadata == nil {
		downloader.remotePageMetadata = make(map[ContentID]RemoteObjectMetadata)
	}
	for _, p := range apiResult.Results {
		if _, ok := downloader.remotePageMetadata[ContentID(p.ID)]; ok {
			return JobResult{}, fmt.Errorf("localdump: received duplicate ID %s from API", p.ID)
		}
		p.ContentType = job.GetPagesQuery.QueryType
		downloader.remotePageMetadata[ContentID(p.ID)] = RemoteObjectMetadata{
			Page: p,
		}
	}

	result := JobResult{
		JobType:    job.JobType,
		space:      job.SpaceKey,
		finished:   apiResult.Links.Next == "",
		itemsFound: len(apiResult.Results),
	}

	// we're done with this space!  happy days!
	if apiResult.Links.Next == "" {
		return result, nil
	}

	q, err := url.Parse(apiResult.Links.Next)
	if err != nil {
		return JobResult{}, fmt.Errorf("confluence: couldn't parse _links.next: %w", err)
	}

	job.GetPagesQuery.Cursor = q.Query().Get("cursor")
	result.followUpJob = &job
	if result.followUpJob.GetPagesQuery.Cursor == "" {
		return JobResult{}, fmt.Errorf("confluence: expected parameter 'cursor' was empty")
	}
	return result, nil
}

func (downloader *SpacesDownloader) performUserDownloadJob(ctx context.Context, job Job) (JobResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	apiResult, err := downloader.API.GetUserByID(ctx, job.GetUserQuery)
	if err != nil {
		return JobResult{}, fmt.Errorf("localdump: failed getting user: %w", err)
	}

	downloader.remoteMetadataMu.Lock()
	defer downloader.remoteMetadataMu.Unlock()

	if downloader.authorMetadata == nil {
		downloader.authorMetadata = make(map[string]confluence.User)
	}
	if _, ok := downloader.authorMetadata[job.GetUserQuery.ID]; ok {
		return JobResult{}, fmt.Errorf("localdump: received duplicate user ID %s from API", job.GetUserQuery.ID)
	}
	downloader.authorMetadata[job.GetUserQuery.ID] = *apiResult

	result := JobResult{
		JobType:    job.JobType,
		finished:   true,
		itemsFound: 1,
		user:       *apiResult,
	}

	return result, nil
}

type DownloadAction int

const (
	SuccessfulDownload DownloadAction = iota
	SkippedCached
)

func (downloader *SpacesDownloader) getPageOrBlog(ctx context.Context, job Job) (*confluence.Page, error) {
	id, err := strconv.Atoi(job.PageID)
	if err != nil {
		return nil, fmt.Errorf("localdump: id was not an int: %w", err)
	}

	if job.ContentType == confluence.BlogContent {
		return downloader.API.GetBlogpostByID(ctx, confluence.GetPageByIDQuery{
			ID:         id,
			BodyFormat: "view",
		})
	} else {
		return downloader.API.GetPageByID(ctx, confluence.GetPageByIDQuery{
			ID:         id,
			BodyFormat: "view",
		})
	}
}

func (downloader *SpacesDownloader) performPageDownloadJob(ctx context.Context, job Job) (JobResult, error) {
	ourItem, ok, err := downloader.LocalVersionIsRecent(ContentID(job.PageID))
	if err != nil {
		return JobResult{}, fmt.Errorf("localdump: failed comparing cached versions: %w", err)
	}
	if ok && !downloader.AlwaysDownload {
		return JobResult{
			JobType: job.JobType,
			space:   job.SpaceKey,

			followUpJob: nil,
			finished:    true,
			itemsFound:  1,

			page:                &ourItem,
			pageDownloadOutcome: SkippedCached,
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := downloader.getPageOrBlog(ctx, job)
	if err != nil {
		return JobResult{}, fmt.Errorf("localdump: failed getting page: %w", err)
	}
	result.ContentType = job.ContentType
	result.SpaceKey = job.SpaceKey
	result.Org = job.Org

	markdown, err := downloader.ConvertToMarkdown(result)
	if err != nil {
		return JobResult{}, fmt.Errorf("localdump: convert to Markdown failed: %w", err)
	}

	if err = downloader.WriteMarkdownIntoLocal(markdown); err != nil {
		return JobResult{}, fmt.Errorf("localdump: failed writing file: %w", err)
	}

	return JobResult{
		JobType: job.JobType,
		space:   job.SpaceKey,

		followUpJob: nil,
		finished:    true,
		itemsFound:  1,

		page:                &markdown,
		pageDownloadOutcome: SuccessfulDownload,
	}, nil
}
