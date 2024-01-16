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

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/toothbrush/confluence-dump/confluence"
	"github.com/toothbrush/confluence-dump/localdump"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Scrape Confluence space and download pages",
	Long:  `TODO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		return runDownload(ctx)
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

	Spaces []string
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	downloadCmd.Flags().BoolVarP(&AlwaysDownload, "always-download", "f", false, "always download pages, skipping version check")
	downloadCmd.Flags().BoolVar(&WithVCR, "with-vcr", false, "use go-vcr to cache responses")
	downloadCmd.Flags().BoolVar(&IncludeBlogposts, "include-blogposts", false, "download blogposts as well as usual posts")
	downloadCmd.Flags().BoolVar(&AllSpaces, "all-spaces", false, "download from all spaces")
	downloadCmd.Flags().BoolVar(&WriteMarkdown, "write-markdown", true, "write Markdown files to disk")
	downloadCmd.Flags().BoolVar(&Prune, "prune", true, "prune local Markdown files after download")
	downloadCmd.Flags().BoolVar(&IncludeArchived, "include-archived", false, "include archived content")

	downloadCmd.PersistentFlags().StringSliceVar(&Spaces, "spaces", []string{}, "list of spaces to scrape")
}

func runDownload(ctx context.Context) error {
	log := log.New(os.Stderr, "[confluence-dump] ", 0)

	if LocalStore == "" {
		return fmt.Errorf("download: no location for local store.  Use --store or set it in your config file")
	}

	storePath, err := homedir.Expand(LocalStore)
	if err != nil {
		return fmt.Errorf("download: couldn't expand homedir: %w", err)
	}

	if _, err := os.Stat(storePath); err != nil {
		log.Printf("Couldn't read %s, have you created the directory?\n", LocalStore)
		return fmt.Errorf("download: couldn't stat storePath %s: %w", storePath, err)
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
	spacesRemote, err := api.ListAllSpaces(ctx, ConfluenceInstance)
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
		WriteMarkdown:   WriteMarkdown,
		Prune:           Prune,
		IncludeArchived: IncludeArchived,
	}

	if err := downloader.DownloadConfluenceSpaces(ctx, spacesToDownload); err != nil {
		return fmt.Errorf("download: Couldn't download spaces: %w", err)
	}

	log.Println("Finished!")

	return nil
}