package confluence_api

import (
	"fmt"

	conf "github.com/virtomize/confluence-go-api"
)

func GetConfluenceAPI(confluence_instance_name string,
	username string,
	token string) (*conf.API, error) {

	api, err := conf.NewAPI(
		fmt.Sprintf("https://%s.atlassian.net/wiki/rest/api", confluence_instance_name),
		username,
		token,
	)
	if err != nil {
		return &conf.API{}, fmt.Errorf("confluence_api: couldn't create API instance: %w", err)
	}

	return api, nil
}
