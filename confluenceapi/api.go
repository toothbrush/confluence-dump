package confluenceapi

import (
	"fmt"

	conf "github.com/virtomize/confluence-go-api"
)

func GetConfluenceAPI(confluence_instance_name string,
	username string,
	token string) (*conf.API, error) {

	if confluence_instance_name == "" {
		return &conf.API{}, fmt.Errorf("Please configure your Confluence instance name --confluence-instance")
	}
	if username == "" {
		return &conf.API{}, fmt.Errorf("Please configure your Confluence username with --auth-username")
	}

	api, err := conf.NewAPI(
		fmt.Sprintf("https://%s.atlassian.net/wiki/rest/api", confluence_instance_name),
		username,
		token,
	)
	if err != nil {
		return &conf.API{}, fmt.Errorf("confluenceapi: couldn't create API instance: %w", err)
	}

	return api, nil
}
