package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	md_plugin "github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/mitchellh/go-homedir"
	conf "github.com/virtomize/confluence-go-api"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	"github.com/toothbrush/confluence-dump/confluence_api"
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

	spaces, err := confluence_api.ListAllSpaces(*api)
	if err != nil {
		log.Fatal(err)
	}

	for _, space := range spaces {
		fmt.Printf("  - %s: %s\n", space.Key, space.Name)
	}

	space_to_export := "CORE"
	pages, err := GetAllPagesInSpace(*api, space_to_export)
	if err != nil {
		log.Fatal(err)
	}

	// build up id to title mapping, so that we can use it to determine the markdown output dir/filename.
	id_title_mapping, err := BuildIDTitleMapping(pages, space_to_export)
	if err != nil {
		log.Fatal(err)
	}

	for _, page := range pages {
		err = GetPageByIDThenStore(*api, page.ID, &id_title_mapping)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetPageByIDThenStore(api conf.API, id string, id_title_mapping *map[string]IdTitleSlug) error {
	c, err := confluence_api.RetrieveContentByID(api, id)
	if err != nil {
		return err
	}

	markdown, err := ConfluenceContentToMarkdown(c, id_title_mapping)
	if err != nil {
		return err
	}

	if err = WriteFileIntoRepo(markdown); err != nil {
		return fmt.Errorf("could not write to repo file: %w", err)
	}

	return nil
}

func ConfluenceContentToMarkdown(content *conf.Content, id_title_mapping *map[string]IdTitleSlug) (MarkdownOutput, error) {
	converter := md.NewConverter("", true, nil)
	// Github flavoured Markdown knows about tables 👍
	converter.Use(md_plugin.GitHubFlavored())
	markdown, err := converter.ConvertString(content.Body.View.Value)
	if err != nil {
		return MarkdownOutput{}, err
	}
	link := content.Links.Base + content.Links.WebUI

	// Are we able to set a base for all URLs?  Currently the Markdown has things like
	// '/wiki/spaces/DRE/pages/2946695376/Tools+and+Infrastructure' which are a bit un ergonomic.
	// we could (fancy mode) resolve to a link in the local dump or (grug mode) just add the
	// https://redbubble.atlassian.net base URL.
	ancestor_names := []string{}
	ancestor_ids := []string{}
	for _, ancestor := range content.Ancestors {
		ancestor_name, ok := (*id_title_mapping)[ancestor.ID]
		if ok {
			ancestor_names = append(
				ancestor_names,
				fmt.Sprintf("\"%s\"", ancestor_name.title),
			)
			ancestor_ids = append(
				ancestor_ids,
				ancestor.ID,
			)
		} else {
			// oh no, found an ID with no title mapped!!
			return MarkdownOutput{}, fmt.Errorf("oh no, found an ID we haven't seen before! %s", ancestor.ID)
		}
	}

	ancestor_ids_str := fmt.Sprintf("[%s]", strings.Join(ancestor_ids, ", "))

	body := fmt.Sprintf(`title: %s
date: %s
version: %d
object_id: %s
uri: %s
status: %s
type: %s
ancestor_names: %s
ancestor_ids: %s
---
%s
`,
		content.Title,
		content.Version.When,
		content.Version.Number,
		content.ID,
		link,
		content.Status,
		content.Type,
		strings.Join(ancestor_names, " > "),
		ancestor_ids_str,
		markdown)

	relativeOutputPath, err := PagePath(*content, id_title_mapping)
	if err != nil {
		return MarkdownOutput{}, fmt.Errorf("Hm, could not determine page path: %w", err)
	}

	return MarkdownOutput{
		id:         content.ID,
		content:    body,
		outputPath: relativeOutputPath,
	}, nil
}

// XXX Hmm this is a deprecated API?
func GetAllPagesInSpace(api conf.API, space string) ([]conf.Content, error) {
	//get content by space name
	there_is_more := true
	results := []conf.Content{}
	var position int

	position = 0
	for there_is_more {
		res, err := api.GetContent(conf.ContentQuery{
			SpaceKey: space,
			Start:    position,
		})
		if err != nil {
			return []conf.Content{}, err
		}
		position += res.Size
		there_is_more = res.Size > 0
		if there_is_more {
			results = append(results, res.Results...)
			fmt.Printf("Found %d items in %s\n", position, space)
		}
	}

	return results, nil
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

type IdTitleSlug struct {
	title     string
	slug      string
	space_key string
}

func BuildIDTitleMapping(pages []conf.Content, space_key string) (map[string]IdTitleSlug, error) {
	id_title_mapping := make(map[string]IdTitleSlug)

	for _, page := range pages {
		slug, err := canonicalise(page.Title)
		if err != nil {
			return nil, err
		}
		id_title_mapping[page.ID] = IdTitleSlug{
			title:     page.Title,
			slug:      slug,
			space_key: space_key,
		}
	}

	return id_title_mapping, nil
}

func WriteFileIntoRepo(contents MarkdownOutput) error {
	// Does REPO_BASE exist?
	expanded_repo_base, err := homedir.Expand(REPO_BASE)
	if err != nil {
		return fmt.Errorf("Couldn't expand homedir from %s: %w", REPO_BASE, err)
	}
	stat, err := os.Stat(expanded_repo_base)
	if err != nil {
		return fmt.Errorf("Error with stat'ing %s: %w", expanded_repo_base, err)
	}

	if !stat.IsDir() {
		// path is not a directory.  this is bad, we should bail
		return fmt.Errorf("REPO_BASE does not seem to be a directory.  Bailing.")
	}

	// construct destination path
	abs := path.Join(expanded_repo_base, contents.outputPath)
	directory := path.Dir(abs)

	fmt.Printf("Writing page %s to: %s...\n", contents.id, path.Join(REPO_BASE, contents.outputPath))
	// XXX there's probably a nicer way to express 0755 but meh
	if err = os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	f, err := os.Create(abs)
	if err != nil {
		return fmt.Errorf("could not create file %s: %w", abs, err)
	}

	defer f.Close()
	f.WriteString(contents.content)

	return nil
}

func PagePath(page conf.Content, id_to_slug *map[string]IdTitleSlug) (string, error) {
	path_parts := []string{}

	for _, ancestor := range page.Ancestors {
		if ancestor_name, ok := (*id_to_slug)[ancestor.ID]; ok {
			path_parts = append(path_parts, ancestor_name.slug)
		} else {
			// oh no, found an ID with no title mapped!!
			return "", fmt.Errorf("oh no, found an ID we haven't seen before! %s", ancestor.ID)
		}
	}

	if my_details, ok := (*id_to_slug)[page.ID]; ok {
		path_parts = append([]string{my_details.space_key}, path_parts...)
		path_parts = append(path_parts, fmt.Sprintf("%s.md", my_details.slug))
	} else {
		// oh no, our own ID isn't in the mapping?
		return "", fmt.Errorf("oh no, couldn't retrieve page ID %s from mapping!", page.ID)
	}

	return path.Join(path_parts...), nil
}

type MarkdownOutput struct {
	content    string
	id         string
	outputPath string
}
