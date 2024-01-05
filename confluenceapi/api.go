package confluenceapi

import (
	"fmt"

	conf "github.com/virtomize/confluence-go-api"
)

func GetConfluenceAPI(confluenceInstanceName string,
	username string,
	token string) (*conf.API, error) {

	if confluenceInstanceName == "" {
		return &conf.API{}, fmt.Errorf("configure your Confluence instance name --confluence-instance")
	}
	if username == "" {
		return &conf.API{}, fmt.Errorf("configure your Confluence username with --auth-username")
	}

	api, err := conf.NewAPI(
		fmt.Sprintf("https://%s.atlassian.net/wiki/rest/api", confluenceInstanceName),
		username,
		token,
	)
	if err != nil {
		return &conf.API{}, fmt.Errorf("confluenceapi: couldn't create API instance: %w", err)
	}

	return api, nil
}
