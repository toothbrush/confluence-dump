package confluence

import (
	"fmt"
	"net/url"

	"github.com/google/go-querystring/query"
)

// getUserByIDEndpoint returns the (v1 but supported) API endpoint to fetch one user:
// https://developer.atlassian.com/cloud/confluence/rest/v1/api-group-users/#api-wiki-rest-api-user-get
func (a *API) getUserByIDEndpoint(opts GetUserByIDQuery) (*url.URL, error) {
	if opts.ID == "" {
		return nil, fmt.Errorf("confluence: please provide ID to get user")
	}

	ep, err := a.resolveEndpoint("/wiki/rest/api/user")
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't resolve endpoint: %w", err)
	}

	v, err := query.Values(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't encode query params: %w", err)
	}
	ep.RawQuery = v.Encode()

	return ep, nil
}

// getPageByIDEndpoint returns the (v2) API endpoint to download one page:
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-blog-post/#api-blogposts-id-get
func (a *API) getBlogpostByIDEndpoint(opts GetPageByIDQuery) (*url.URL, error) {
	if opts.ID < 1 {
		return nil, fmt.Errorf("confluence: please provide ID to get page by ID")
	}

	ep, err := a.resolveEndpoint(fmt.Sprintf("/wiki/api/v2/blogposts/%d", opts.ID))
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't resolve endpoint: %w", err)
	}

	v, err := query.Values(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't encode query params: %w", err)
	}
	ep.RawQuery = v.Encode()

	return ep, nil
}

// getPageByIDEndpoint returns the (v2) API endpoint to download one page:
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-page/#api-pages-id-get
func (a *API) getPageByIDEndpoint(opts GetPageByIDQuery) (*url.URL, error) {
	if opts.ID < 1 {
		return nil, fmt.Errorf("confluence: please provide ID to get page by ID")
	}

	ep, err := a.resolveEndpoint(fmt.Sprintf("/wiki/api/v2/pages/%d", opts.ID))
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't resolve endpoint: %w", err)
	}

	v, err := query.Values(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't encode query params: %w", err)
	}
	ep.RawQuery = v.Encode()

	return ep, nil
}

// getFolderByIDEndpoint returns the (v2) API endpoint to download a folder:
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-folder/#api-folders-id-get
func (a *API) getFolderByIDEndpoint(opts GetFolderByIDQuery) (*url.URL, error) {
	if opts.ID < 1 {
		return nil, fmt.Errorf("confluence: please provide folder ID")
	}

	ep, err := a.resolveEndpoint(fmt.Sprintf("/wiki/api/v2/folders/%d", opts.ID))
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't resolve folder endpoint: %w", err)
	}

	v, err := query.Values(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't encode query params: %w", err)
	}
	ep.RawQuery = v.Encode()
	return ep, nil
}

// getBlogPostsEndpoint returns the (v2) API endpoint to list pages
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-blog-post/#api-blogposts-get
func (a *API) getBlogPostsEndpoint(opts GetPagesQuery) (*url.URL, error) {
	ep, err := a.resolveEndpoint("/wiki/api/v2/blogposts")
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't resolve endpoint: %w", err)
	}

	v, err := query.Values(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't encode query params: %w", err)
	}
	ep.RawQuery = v.Encode()

	return ep, nil
}

// getPagesEndpoint returns the (v2) API endpoint to list pages
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-page/#api-pages-get
func (a *API) getPagesEndpoint(opts GetPagesQuery) (*url.URL, error) {
	ep, err := a.resolveEndpoint("/wiki/api/v2/pages")
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't resolve endpoint: %w", err)
	}

	v, err := query.Values(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't encode query params: %w", err)
	}
	ep.RawQuery = v.Encode()

	return ep, nil
}

// getSpaceEndpoint returns the (v2) API endpoint to list spaces
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-space/#api-spaces-get
func (a *API) getSpaceEndpoint(opts SpacesQuery) (*url.URL, error) {
	ep, err := a.resolveEndpoint("/wiki/api/v2/spaces")
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't resolve endpoint: %w", err)
	}

	v, err := query.Values(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't encode query params: %w", err)
	}
	ep.RawQuery = v.Encode()

	return ep, nil
}

// getCurrentUserEndpoint returns the (v1) API endpoint to query current user
// https://developer.atlassian.com/cloud/confluence/rest/v1/api-group-users/#api-wiki-rest-api-user-current-get
//
// This API is supported.
func (a *API) getCurrentUserEndpoint() (*url.URL, error) {
	return a.resolveEndpoint("/wiki/rest/api/user/current")
}

// Do a bit of error checking on endpoint format, and return it relative to the base URI.
func (a *API) resolveEndpoint(endpoint string) (*url.URL, error) {
	baseUri := a.BaseURI

	ref, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("confluence: failed to parse endpoint ref: %w", err)
	}

	return baseUri.ResolveReference(ref), nil
}
