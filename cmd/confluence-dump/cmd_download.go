/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/toothbrush/confluence-dump/confluence"
	"github.com/toothbrush/confluence-dump/localdump"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

var downloadUsage = strings.TrimSpace(`
Download Confluence posts
-------------------------

This command downloads the specified Confluence wiki spaces.  Probably, you'll want to configure
--spaces to only grab the spaces you want, but you may be greedy and choose --all-spaces.  If you
want inspiration, see 'confluence-dump list spaces'.

Typically, the download command will go through a few phases:

1. Ensure that the spaces you specified are valid and accessible.
2. Download a list of all posts and their version metadata, in the given spaces (this can take a
while).
3. Compare each page to whatever you have cached in your local Markdown store, and if that's stale,
download (of course, if this is the first time you're running confluence-dump, everything will
get downloaded).
4. Once the download is complete, all Markdown files in your local store that we _haven't_
downloaded or skipped will be assumed stale (e.g., they got moved, or are now deleted).  These files
will be deleted.  This only happens for spaces we scraped, so if your store has space A & B but this
time your ran with --spaces=A, we won't touch B's files at all.

Example invocation:

$ confluence-dump --spaces=CORE,DRE
$ confluence-dump --all-spaces # Disregards your configured list of spaces
`)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download pages from Confluence space(s)",
	Long:  downloadUsage,
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		return runDownload(ctx)
	},

	PostRunE: func(cmd *cobra.Command, args []string) error {
		if len(PostDownloadCmd) < 1 {
			// no post-download command specified, probably.
			return nil
		}

		postDownloadCmd := exec.Command(PostDownloadCmd[0], PostDownloadCmd[1:]...)
		expandedDir, err := homedir.Expand(ParsedConfig.StorePath)
		if err != nil {
			return fmt.Errorf("download: couldn't expand homedir: %w", err)
		}
		postDownloadCmd.Dir = expandedDir
		postDownloadCmdOutput, err := postDownloadCmd.CombinedOutput()

		if output := strings.TrimSpace(string(postDownloadCmdOutput)); output != "" {
			fmt.Printf("post-download-cmd output:\n%s\n", output)
		}
		if err != nil {
			return fmt.Errorf("download: failed to execute post-download-cmd '%v': %w", PostDownloadCmd, err)
		}

		return nil
	},
}

var (
	AlwaysDownload   bool
	WithVCR          bool
	IncludeBlogposts bool
	AllSpaces        bool
	WriteMarkdown    bool
	Prune            bool
	IncludeArchived  bool
	IncludePersonal  bool

	Spaces []string

	PostDownloadCmd []string
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	downloadCmd.Flags().BoolVarP(&AlwaysDownload, "always-download", "f", false, "always download pages, skipping version check")
	downloadCmd.Flags().BoolVar(&WithVCR, "with-vcr", false, "use go-vcr to cache responses")
	downloadCmd.Flags().BoolVar(&AllSpaces, "all-spaces", false, "download from all spaces")
	downloadCmd.Flags().BoolVar(&WriteMarkdown, "write-markdown", true, "write Markdown files to disk")
	downloadCmd.Flags().BoolVar(&Prune, "prune", true, "prune local Markdown files after download")
	downloadCmd.Flags().BoolVar(&IncludeArchived, "include-archived", false, "include archived content")
	downloadCmd.Flags().BoolVar(&IncludeBlogposts, "include-blogposts", false, "download blogposts as well as usual posts")
	downloadCmd.Flags().BoolVar(&IncludePersonal, "include-personal-spaces", false, "download pages from individuals' personal spaces")

	downloadCmd.PersistentFlags().StringSliceVar(&Spaces, "spaces", []string{}, "list of spaces to scrape")
	downloadCmd.PersistentFlags().StringSliceVar(&PostDownloadCmd, "post-download-cmd", []string{}, "command to execute after download")
}

func runDownload(ctx context.Context) error {
	start := time.Now()

	log := log.New(os.Stderr, "[confluence-dump] ", 0)

	if LocalStore == "" {
		return fmt.Errorf("download: no location for local store; use --store or set in config file")
	}

	storePath, err := homedir.Expand(LocalStore)
	if err != nil {
		return fmt.Errorf("download: couldn't expand homedir: %w", err)
	}

	storePathInfo, err := os.Stat(storePath)
	if os.IsNotExist(err) {
		log.Printf("Store '%s' doesn't exist, creating.\n", storePath)
		if err := os.Mkdir(storePath, os.FileMode(0755)); err != nil {
			return fmt.Errorf("download: failed to create directory %s: %w", storePath, err)
		}
	} else if err != nil {
		// this means there's an error, and it's not the one kind we know how to recover from
		return fmt.Errorf("download: failed to stat directory %s: %w", storePath, err)
	} else if !storePathInfo.IsDir() {
		return fmt.Errorf("download: storePath `%s` is not a directory", storePath)
	}

	storePathWithOrg := path.Join(storePath, ConfluenceInstance)
	if err := os.MkdirAll(storePathWithOrg, 0750); err != nil {
		return fmt.Errorf("localdump: couldn't create directory %s: %w", storePathWithOrg, err)
	}

	tokenCmdOutput, err := exec.Command(AuthTokenCmd[0], AuthTokenCmd[1:]...).Output()
	if err != nil {
		return fmt.Errorf("download: couldn't execute auth-token-cmd '%v': %w", AuthTokenCmd, err)
	}

	token := strings.Split(string(tokenCmdOutput), "\n")[0]

	api, err := confluence.NewAPI(
		ConfluenceInstance,
		AuthUsername,
		token)
	if err != nil {
		return fmt.Errorf("download: couldn't instantiate Confluence API: %w", err)
	}

	if WithVCR {
		// set up VCR recordings.
		opts := &recorder.Options{
			CassetteName:       "fixtures/confluence-stuff",
			Mode:               recorder.ModeReplayWithNewEpisodes,
			SkipRequestLatency: true,
			RealTransport:      http.DefaultTransport,
		}
		r, err := recorder.NewWithOptions(opts)
		if err != nil {
			return fmt.Errorf("download: couldn't set up go-vcr recording: %w", err)
		}

		defer r.Stop() // Make sure recorder is stopped once done with it

		// Add a hook which removes Authorization headers from all requests
		hook := func(i *cassette.Interaction) error {
			delete(i.Request.Headers, "Authorization")
			return nil
		}
		r.AddHook(hook, recorder.AfterCaptureHook)
		r.SetReplayableInteractions(true)

		vcrClient := r.GetDefaultClient()
		api.Client = vcrClient
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// get current user information
	currentUser, err := api.CurrentUser(ctx)
	if err != nil {
		return fmt.Errorf("download: couldn't query current user: %w", err)
	}

	log.Printf("Logged in to id.atlassian.com as '%s (%s)'...\n", currentUser.DisplayName, currentUser.Email)

	// list all spaces:
	log.Println("Listing Confluence spaces...")
	spacesRemote, err := api.ListAllSpaces(ctx, ConfluenceInstance, IncludePersonal)
	if err != nil {
		return fmt.Errorf("download: couldn't list Confluence spaces: %w", err)
	}
	log.Printf("Found %d spaces on '%s'.\n", len(spacesRemote), ConfluenceInstance)

	spacesToDownload := []confluence.Space{}
	if AllSpaces {
		for _, sp := range spacesRemote {
			spacesToDownload = append(spacesToDownload, sp)
		}
	} else {
		for _, requestedSpace := range Spaces {
			sp, ok := spacesRemote[requestedSpace]
			if !ok {
				return fmt.Errorf("download: requested space %s does not exist", requestedSpace)
			}
			spacesToDownload = append(spacesToDownload, sp)
		}
	}

	if IncludeBlogposts {
		// Add phantom "space" for storing blogposts:
		spacesToDownload = append(spacesToDownload,
			confluence.Space{
				ID:   "blogposts",
				Key:  "blogposts",
				Name: "Users' blogposts",
				Org:  ConfluenceInstance,
			},
		)
	}

	if len(spacesToDownload) == 0 {
		return fmt.Errorf("cmd_download: no spaces selected, nothing to do")
	}

	log.Println("Enqueuing for download:")
	for _, space := range spacesToDownload {
		log.Printf("  - %s: %s\n", space.Key, space.Name)
	}

	if AllSpaces && len(Spaces) > 0 {
		log.Println("🚨 WARNING: Both --all-spaces && --spaces set, ignoring --spaces.")
	}

	downloader := localdump.SpacesDownloader{
		StorePath:       storePath,
		Workers:         runtime.NumCPU(),
		Logger:          log,
		AlwaysDownload:  AlwaysDownload,
		API:             api,
		Debug:           Debug,
		WriteMarkdown:   WriteMarkdown,
		Prune:           Prune,
		IncludeArchived: IncludeArchived,
		IncludePersonal: IncludePersonal,
	}

	if err := downloader.DownloadConfluenceSpaces(ctx, spacesToDownload); err != nil {
		return fmt.Errorf("download: Couldn't download spaces: %w", err)
	}

	duration := time.Since(start)
	log.Printf("Finished in %s!\n", duration)

	return nil
}
