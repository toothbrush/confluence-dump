package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	"github.com/mitchellh/go-homedir"
	"github.com/toothbrush/confluence-dump/confluence_api"
	"github.com/toothbrush/confluence-dump/data"
	"github.com/toothbrush/confluence-dump/local_dump"
)

const REPO_BASE = "~/confluence"

func main() {
	storePath, err := homedir.Expand(REPO_BASE)
	if err != nil {
		log.Fatal(err)
	}

	token_cmd_output, err := exec.Command("pass", "confluence-api-token/paul.david@redbubble.com").Output()
	if err != nil {
		log.Fatal(err)
	}

	token := strings.Split(string(token_cmd_output), "\n")[0]

	local_dump.LoadLocalMarkdown(storePath)
	return

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
	title_cache, err := data.BuildCacheFromPagelist(pages, space_to_export)
	if err != nil {
		log.Fatal(err)
	}

	for _, page := range pages {
		c, err := confluence_api.RetrieveContentByID(*api, page.ID)
		if err != nil {
			log.Fatal(err)
		}

		markdown, err := data.ConvertToMarkdown(c, title_cache)
		if err != nil {
			log.Fatal(err)
		}

		if err = local_dump.WriteMarkdownIntoLocal(storePath, markdown); err != nil {
			log.Fatal(err)
		}
	}
}
