/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/toothbrush/confluence-dump/confluence_api"
	"github.com/toothbrush/confluence-dump/data"
	"github.com/toothbrush/confluence-dump/local_dump"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Scrape Confluence space and download pages",
	Long:  `TODO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("  AlwaysDownload: %v\n", AlwaysDownload)
		return runDownload()
	},
}

var (
	AlwaysDownload bool
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	downloadCmd.Flags().BoolVarP(&AlwaysDownload, "always-download", "f", false, "Always download pages, skipping version check")
}

func runDownload() error {
	if LocalStore == "" {
		return fmt.Errorf("No location set for local store of Confluence data.  Use --store or set it in your config file.")
	}

	storePath, err := homedir.Expand(LocalStore)
	if err != nil {
		return fmt.Errorf("cmd: Confluence download failed: %w", err)
	}

	if _, err := os.Stat(storePath); err != nil {
		return fmt.Errorf("cmd: Confluence download failed: %w", err)
	}

	token_cmd_output, err := exec.Command(AuthTokenCmd[0], AuthTokenCmd[1:]...).Output()
	if err != nil {
		return fmt.Errorf("cmd: Confluence download failed: %w", err)
	}

	token := strings.Split(string(token_cmd_output), "\n")[0]

	local_markdown, err := local_dump.LoadLocalMarkdown(storePath)
	if err != nil {
		return fmt.Errorf("cmd: Confluence download failed: %w", err)
	}

	api, err := confluence_api.GetConfluenceAPI(
		ConfluenceInstance,
		AuthUsername,
		token)
	if err != nil {
		return fmt.Errorf("cmd: Confluence download failed: %w", err)
	}

	// set up VCR recordings.
	opts := &recorder.Options{
		CassetteName:       "fixtures/confluence-stuff",
		Mode:               recorder.ModeReplayWithNewEpisodes,
		SkipRequestLatency: true,
		RealTransport:      http.DefaultTransport,
	}
	r, err := recorder.NewWithOptions(opts)
	if err != nil {
		log.Fatal(fmt.Errorf("couldn't set up go-vcr recording: %w", err))
	}

	defer r.Stop() // Make sure recorder is stopped once done with it

	// Add a hook which removes Authorization headers from all requests
	hook := func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	}
	r.AddHook(hook, recorder.AfterCaptureHook)
	r.SetReplayableInteractions(true)

	vcr_client := r.GetDefaultClient()
	api.Client = vcr_client

	// get current user information
	currentUser, err := api.CurrentUser()
	if err != nil {
		return fmt.Errorf("cmd: Confluence download failed: %w", err)
	}

	fmt.Printf("Logged in to id.atlassian.com as '%s (%s)'...\n", currentUser.DisplayName, currentUser.AccountID)

	// list all spaces
	spaces, err := confluence_api.ListAllSpaces(*api)
	if err != nil {
		return fmt.Errorf("cmd: Confluence download failed: %w", err)
	}

	for _, space := range spaces {
		fmt.Printf("  - %s: %s\n", space.Key, space.Name)
	}

	// grab a list of pages from given space
	space_to_export := "CORE"
	pages, err := confluence_api.GetAllPagesInSpace(*api, space_to_export)
	if err != nil {
		return fmt.Errorf("cmd: Confluence download failed: %w", err)
	}

	// build up id to title mapping, so that we can use it to determine the markdown output dir/filename.
	remote_title_cache, err := data.BuildCacheFromPagelist(pages, space_to_export)
	if err != nil {
		return fmt.Errorf("cmd: Confluence download failed: %w", err)
	}
	fmt.Printf("Found %d pages on remote...\n", len(remote_title_cache))

	pages_to_download := local_dump.ChangedPages(remote_title_cache, local_markdown)
	fmt.Printf("Found %d updated pages to download...\n", len(pages_to_download))

	for _, page := range pages_to_download {
		c, err := confluence_api.RetrieveContentByID(*api, page)
		if err != nil {
			return fmt.Errorf("cmd: Confluence download failed: %w", err)
		}

		markdown, err := data.ConvertToMarkdown(c, remote_title_cache)
		if err != nil {
			return fmt.Errorf("cmd: Confluence download failed: %w", err)
		}

		if err = local_dump.WriteMarkdownIntoLocal(storePath, markdown); err != nil {
			return fmt.Errorf("cmd: Confluence download failed: %w", err)
		}
	}

	return nil
}
