/*
Copyright © 2024 paul <paul@denknerd.org>
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"
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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("download called")
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

const REPO_BASE = "~/confluence"

func runDownload() {
	storePath, err := homedir.Expand(REPO_BASE)
	if err != nil {
		log.Fatal(err)
	}

	token_cmd_output, err := exec.Command("pass", "confluence-api-token/paul.david@redbubble.com").Output()
	if err != nil {
		log.Fatal(err)
	}

	token := strings.Split(string(token_cmd_output), "\n")[0]

	local_markdown, err := local_dump.LoadLocalMarkdown(storePath)
	if err != nil {
		log.Fatal(err)
	}

	api, err := confluence_api.GetConfluenceAPI(
		"redbubble",
		"paul.david@redbubble.com",
		token)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	fmt.Printf("Logged in to id.atlassian.com as '%s (%s)'...\n", currentUser.DisplayName, currentUser.AccountID)

	// list all spaces
	spaces, err := confluence_api.ListAllSpaces(*api)
	if err != nil {
		log.Fatal(err)
	}

	for _, space := range spaces {
		fmt.Printf("  - %s: %s\n", space.Key, space.Name)
	}

	// grab a list of pages from given space
	space_to_export := "CORE"
	pages, err := confluence_api.GetAllPagesInSpace(*api, space_to_export)
	if err != nil {
		log.Fatal(err)
	}

	// build up id to title mapping, so that we can use it to determine the markdown output dir/filename.
	remote_title_cache, err := data.BuildCacheFromPagelist(pages, space_to_export)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d pages on remote...\n", len(remote_title_cache))

	pages_to_download := local_dump.ChangedPages(remote_title_cache, local_markdown)
	fmt.Printf("Found %d updated pages to download...\n", len(pages_to_download))

	for _, page := range pages_to_download {
		c, err := confluence_api.RetrieveContentByID(*api, page)
		if err != nil {
			log.Fatal(err)
		}

		markdown, err := data.ConvertToMarkdown(c, remote_title_cache)
		if err != nil {
			log.Fatal(err)
		}

		if err = local_dump.WriteMarkdownIntoLocal(storePath, markdown); err != nil {
			log.Fatal(err)
		}
	}
}
