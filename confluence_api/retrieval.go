package confluence_api

import (
	"fmt"
	"os"

	conf "github.com/virtomize/confluence-go-api"
)

// XXX(pd) 20240104: Hmm, this is a deprecated API? (seen in VCR recording)
func GetAllPagesInSpace(api conf.API, space string) ([]conf.Content, error) {
	// get content (just metadata) by space name
	more := true
	contents := []conf.Content{}
	position := 0

	for more {
		res, err := api.GetContent(conf.ContentQuery{
			SpaceKey: space,
			Start:    position,
			Expand:   []string{"version"},
		})
		if err != nil {
			return []conf.Content{}, fmt.Errorf("confluence_api: couldn't retrieve list of contents: %w", err)
		}

		position += res.Size
		more = res.Size > 0

		if more {
			contents = append(contents, res.Results...)
			fmt.Fprintf(os.Stderr, "Found %d items in %s...\n", position, space)
		}
	}

	return contents, nil
}

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
	position := 0
	spaces := []conf.Space{}

	for more {
		allspaces, err := api.GetAllSpaces(conf.AllSpacesQuery{
			Type:  "global",
			Start: position,
			Limit: 10,
		})

		if err != nil {
			return []conf.Space{}, fmt.Errorf("confluence_api: couldn't list spaces: %w", err)
		}

		position += allspaces.Size
		more = allspaces.Size > 0

		if more {
			for _, space := range allspaces.Results {
				spaces = append(spaces, space)
			}
			fmt.Fprintf(os.Stderr, "Found %d spaces...\n", position)
		}
	}

	return spaces, nil
}
