package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	md_plugin "github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/mitchellh/go-homedir"
	conf "github.com/virtomize/confluence-go-api"
)

const REPO_BASE = "~/confluence"

func main() {
	api, err := GiveMeAnAPIInstance()
	if err != nil {
		log.Fatal(err)
	}

	// get current user information
	currentUser, err := api.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Logged in to id.atlassian.com as '%s (%s)'...\n", currentUser.DisplayName, currentUser.UserKey)

	space_to_export := "DRE"
	pages, err := GetAllPagesInSpace(*api, space_to_export)
	if err != nil {
		log.Fatal(err)
	}

	// build up id to title mapping, so that we can use it to determine the markdown output dir/filename.
	id_title_mapping, err := BuildIDTitleMapping(pages)
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
	c, err := GetOnePage(api, id)
	if err != nil {
		return err
	}

	markdown, err := ConfluenceContentToMarkdown(c, id_title_mapping)
	if err != nil {
		return err
	}

	path, err := PagePath(*c, id_title_mapping)
	if err != nil {
		return err
	}
	fmt.Printf("Writing page %s to: %s...\n", c.ID, path)
	WriteFileIntoRepo(path, markdown)

	return nil
}

func GetOnePage(api conf.API, id string) (*conf.Content, error) {
	c, err := api.GetContentByID(id, conf.ContentQuery{
		Expand: []string{"ancestors", "body.view", "links", "version"},
	})
	if err != nil {
		return &conf.Content{}, err
	}
	return c, nil
}

func ConfluenceContentToMarkdown(content *conf.Content, id_title_mapping *map[string]IdTitleSlug) (string, error) {
	converter := md.NewConverter("", true, nil)
	// Github flavoured Markdown knows about tables 👍
	converter.Use(md_plugin.GitHubFlavored())
	markdown, err := converter.ConvertString(content.Body.View.Value)
	if err != nil {
		return "", err
	}
	link := content.Links.Base + content.Links.WebUI

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
			return "", fmt.Errorf("oh no, found an ID we haven't seen before! %s", ancestor.ID)
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

	return body, nil
}

func PrintAllSpaces(api conf.API) error {
	fmt.Printf("Listing Confluence spaces:\n\n")
	spaces, err := api.GetAllSpaces(conf.AllSpacesQuery{
		Type:  "global",
		Start: 0,
		Limit: 1000,
	})
	if err != nil {
		return err
	}

	for _, space := range spaces.Results {
		fmt.Printf("  - %s: %s\n", space.Key, space.Name)
	}

	return nil
}

func GiveMeAnAPIInstance() (*conf.API, error) {
	token_cmd_output, err := exec.Command("pass", "confluence-api-token/paul.david@redbubble.com").Output()
	if err != nil {
		return &conf.API{}, err
	}

	token := strings.Split(string(token_cmd_output), "\n")[0]

	// initialize a new api instance
	api, err := conf.NewAPI("https://redbubble.atlassian.net/wiki/rest/api", "paul.david@redbubble.com", token)
	if err != nil {
		return &conf.API{}, err
	}

	return api, nil
}

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
	title string
	slug  string
}

func BuildIDTitleMapping(pages []conf.Content) (map[string]IdTitleSlug, error) {
	id_title_mapping := make(map[string]IdTitleSlug)

	for _, page := range pages {
		slug, err := canonicalise(page.Title)
		if err != nil {
			return nil, err
		}
		id_title_mapping[page.ID] = IdTitleSlug{
			title: page.Title,
			slug:  slug,
		}
	}

	return id_title_mapping, nil
}

func WriteFileIntoRepo(relativeFilename string, contents string) error {
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
	abs := path.Join(REPO_BASE, relativeFilename)
	directory := path.Dir(abs)

	// XXX there's probably a nicer way to express 0755 but meh
	if err = os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	f, err := os.Create(abs)
	if err != nil {
		return fmt.Errorf("could not create file %s: %w", abs, err)
	}

	defer f.Close()
	f.WriteString(contents)

	return nil
}

func PagePath(page conf.Content, id_to_slug *map[string]IdTitleSlug) (string, error) {
	path_parts := []string{}
	for _, ancestor := range page.Ancestors {
		ancestor_name, ok := (*id_to_slug)[ancestor.ID]
		if ok {
			path_parts = append(path_parts, ancestor_name.slug)
		} else {
			// oh no, found an ID with no title mapped!!
			return "", fmt.Errorf("oh no, found an ID we haven't seen before! %s", ancestor.ID)
		}
	}

	my_canonical_slug, err := canonicalise(page.Title)
	if err != nil {
		return "", err
	}

	path_parts = append(path_parts, fmt.Sprintf("%s.md", my_canonical_slug))

	return path.Join(path_parts...), nil
}
