package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	conf "github.com/virtomize/confluence-go-api"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	"github.com/toothbrush/confluence-dump/confluence_api"
	"github.com/toothbrush/confluence-dump/data"
	"github.com/toothbrush/confluence-dump/local_dump"
)

const REPO_BASE = "~/confluence"

func main() {
	token_cmd_output, err := exec.Command("pass", "confluence-api-token/paul.david@redbubble.com").Output()
	if err != nil {
		log.Fatal(err)
	}

	token := strings.Split(string(token_cmd_output), "\n")[0]

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
	id_title_mapping, err := BuildIDTitleMapping(pages, space_to_export)
	if err != nil {
		log.Fatal(err)
	}

	for _, page := range pages {
		err = GetPageByIDThenStore(*api, page.ID, id_title_mapping)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetPageByIDThenStore(api conf.API, id string, id_title_mapping data.MetadataCache) error {
	c, err := confluence_api.RetrieveContentByID(api, id)
	if err != nil {
		return err
	}

	markdown, err := data.ConvertToMarkdown(c, id_title_mapping)
	if err != nil {
		return err
	}

	if err = local_dump.WriteMarkdownIntoLocal(markdown); err != nil {
		return fmt.Errorf("could not write to repo file: %w", err)
	}

	return nil
}

func canonicalise(title string) (string, error) {
	str := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(title, " ")
	str = strings.ToLower(str)
	str = strings.Join(strings.Fields(str), "-")

	if len(str) > 101 {
		str = str[:100]
	}

	str = strings.Trim(str, "-")

	if len(str) < 2 {
		return "", fmt.Errorf("Hm, slug ends up too short: '%s'", title)
	}

	return str, nil
}

func BuildIDTitleMapping(pages []conf.Content, space_key string) (data.MetadataCache, error) {
	id_title_mapping := make(data.MetadataCache)

	for _, page := range pages {
		slug, err := canonicalise(page.Title)
		if err != nil {
			return nil, err
		}
		id_title_mapping[page.ID] = data.LocalMetadata{
			Title:    page.Title,
			Slug:     slug,
			SpaceKey: space_key,
		}
	}

	return id_title_mapping, nil
}
