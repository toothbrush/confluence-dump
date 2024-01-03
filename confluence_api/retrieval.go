package confluence_api

import (
	"fmt"

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
