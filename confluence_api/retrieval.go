package confluence_api

import (
	"fmt"
	"os"

	conf "github.com/virtomize/confluence-go-api"
)

func RetrieveContentByID(api conf.API, id string) (*conf.Content, error) {
	content, err := api.GetContentByID(id, conf.ContentQuery{
		Expand: []string{"ancestors", "body.view", "links", "version"},
	})
	if err != nil {
		return &conf.Content{}, fmt.Errorf("confluence_api: couldn't retrieve object id %s: %w", id, err)
	}

	return content, nil
}

func ListAllSpaces(api conf.API) ([]conf.Space, error) {
	more := true
	pointer := 0
	spaces := []conf.Space{}

	for more {
		allspaces, err := api.GetAllSpaces(conf.AllSpacesQuery{
			Type:  "global",
			Start: pointer,
			Limit: 10,
		})

		if err != nil {
			return []conf.Space{}, fmt.Errorf("confluence_api: couldn't list spaces: %w", err)
		}

		pointer += allspaces.Size
		more = allspaces.Size > 0

		if more {
			for _, space := range allspaces.Results {
				spaces = append(spaces, space)
			}
			fmt.Fprintf(os.Stderr, "Found %d spaces...\n", pointer)
		}
	}

	return spaces, nil
}
