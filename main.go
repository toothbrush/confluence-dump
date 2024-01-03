package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	md_plugin "github.com/JohannesKaufmann/html-to-markdown/plugin"
	goconfluence "github.com/virtomize/confluence-go-api"
)

func main() {
	token, err := exec.Command("pass", "confluence-api-token/paul.david@redbubble.com").Output()
	if err != nil {
		log.Fatal(err)
	}

	token_lines := strings.Split(strings.TrimSuffix(string(token), "\n"), "\n")

	// initialize a new api instance
	api, err := goconfluence.NewAPI("https://redbubble.atlassian.net/wiki/rest/api", "paul.david@redbubble.com", token_lines[0])
	if err != nil {
		log.Fatal(err)
	}

	// get current user information
	currentUser, err := api.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Logged in to id.atlassian.com as '%s <%s>'...\n", currentUser.DisplayName, "..")

	fmt.Printf("Listing Confluence spaces:\n\n")
	spaces, err := api.GetAllSpaces(goconfluence.AllSpacesQuery{
		Type:  "global",
		Start: 0,
		Limit: 1000,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, space := range spaces.Results {
		fmt.Printf("  - %s: %s\n", space.Key, space.Name)
	}

	space_to_export := "DRE"

	//get content by query
	there_is_more := true
	results := []goconfluence.Content{}
	var position int

	position = 0
	for there_is_more {
		res, err := api.GetContent(goconfluence.ContentQuery{
			SpaceKey: space_to_export,
			Start:    position,
		})
		if err != nil {
			log.Fatal(err)
		}
		position += res.Size
		fmt.Printf("Found %d items in %s\n", position, space_to_export)
		results = append(results, res.Results...)
		there_is_more = res.Size > 0
	}
	id := "128385319"
	c, err := api.GetContentByID(id, goconfluence.ContentQuery{
		Expand: []string{"body.view", "links", "version"},
	})
	if err != nil {
		log.Fatal(err)
	}

	markdown, err := ConfluenceContentToMarkdown(*c)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(markdown)
}

func ConfluenceContentToMarkdown(content goconfluence.Content) (string, error) {

	converter := md.NewConverter("", true, nil)
	// Github flavoured Markdown knows about tables 👍
	converter.Use(md_plugin.GitHubFlavored())
	markdown, err := converter.ConvertString(content.Body.View.Value)
	if err != nil {
		return "", err
	}
	link := content.Links.Base + content.Links.WebUI

	body := fmt.Sprintf(`title: %s
date: %s
version: %d
object_id: %s
uri: %s
status: %s
type: %s
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
		markdown)

	return body, nil
}
