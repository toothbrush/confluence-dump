package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	md_plugin "github.com/JohannesKaufmann/html-to-markdown/plugin"
	goconfluence "github.com/virtomize/confluence-go-api"
)

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

	fmt.Printf("Logged in to id.atlassian.com as '%s <%s>'...\n", currentUser.DisplayName, "..")

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

	id := "128385319"
	c, err := GetOnePage(*api, id)
	if err != nil {
		log.Fatal(err)
	}

	markdown, err := ConfluenceContentToMarkdown(c, &id_title_mapping)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(markdown)
}

func GetOnePage(api goconfluence.API, id string) (*goconfluence.Content, error) {
	c, err := api.GetContentByID(id, goconfluence.ContentQuery{
		Expand: []string{"ancestors", "body.view", "links", "version"},
	})
	if err != nil {
		return &goconfluence.Content{}, err
	}
	return c, nil
}

func ConfluenceContentToMarkdown(content *goconfluence.Content, id_title_mapping *map[string]IdTitleSlug) (string, error) {
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

func PrintAllSpaces(api goconfluence.API) error {
	fmt.Printf("Listing Confluence spaces:\n\n")
	spaces, err := api.GetAllSpaces(goconfluence.AllSpacesQuery{
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

func GiveMeAnAPIInstance() (*goconfluence.API, error) {
	token, err := exec.Command("pass", "confluence-api-token/paul.david@redbubble.com").Output()
	if err != nil {
		return &goconfluence.API{}, err
	}

	token_lines := strings.Split(strings.TrimSuffix(string(token), "\n"), "\n")

	// initialize a new api instance
	api, err := goconfluence.NewAPI("https://redbubble.atlassian.net/wiki/rest/api", "paul.david@redbubble.com", token_lines[0])
	if err != nil {
		return &goconfluence.API{}, err
	}

	return api, nil
}

func GetAllPagesInSpace(api goconfluence.API, space string) ([]goconfluence.Content, error) {
	//get content by space name
	there_is_more := true
	results := []goconfluence.Content{}
	var position int

	position = 0
	for there_is_more {
		res, err := api.GetContent(goconfluence.ContentQuery{
			SpaceKey: space,
			Start:    position,
		})
		if err != nil {
			return []goconfluence.Content{}, err
		}
		position += res.Size
		fmt.Printf("Found %d items in %s\n", position, space)
		results = append(results, res.Results...)
		there_is_more = res.Size > 0
	}

	return results, nil
}

func canonicalise(title string) (string, error) {
	str := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(title, " ")
	str = strings.Trim(str, "-")
	str = strings.ToLower(str)
	str = strings.Join(strings.Fields(str), "-")

	if len(str) < 2 {
		return "", fmt.Errorf("Hm, slug ends up too short: '%s'", title)
	}

	return str, nil
}

type IdTitleSlug struct {
	title string
	slug  string
}

func BuildIDTitleMapping(pages []goconfluence.Content) (map[string]IdTitleSlug, error) {
	id_title_mapping := make(map[string]IdTitleSlug)

	for _, page := range pages {
		fmt.Printf(" - %s: %s\n", page.ID, page.Title)
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
