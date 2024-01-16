package confluence

import (
	"fmt"
	"net/http"
	"net/url"
)

func NewAPI(instance string, username string, token string) (*API, error) {

	if instance == "" {
		return &API{}, fmt.Errorf("confluence: configure your Confluence instance name --confluence-instance")
	}
	if username == "" {
		return &API{}, fmt.Errorf("confluence: configure your Confluence username with --auth-username")
	}
	if token == "" {
		return &API{}, fmt.Errorf("confluence: auth token is empty, please check auth-token-cmd")
	}

	u, err := url.ParseRequestURI(
		fmt.Sprintf("https://%s.atlassian.net/wiki",
			instance,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't parse REST API URL: %w", err)
	}

	a := &API{
		BaseURI:  u,
		token:    token,
		username: username,
	}
	a.Client = &http.Client{}

	return a, nil
}

type API struct {
	// The name of the Confluence instance, e.g. https://INSTANCE.atlassian.net
	BaseURI *url.URL

	// An HTTP client - you can substitute VCR or whatnot.
	Client *http.Client

	// Auth info
	username, token string
}
