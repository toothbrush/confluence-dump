package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

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
}
