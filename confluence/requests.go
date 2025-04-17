package confluence

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (api *API) GetUserByID(ctx context.Context, opts GetUserByIDQuery) (*User, error) {
	ep, err := api.getUserByIDEndpoint(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't get user endpoint: %w", err)
	}

	body, err := api.request(ctx, ep)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't perform request: %w", err)
	}

	var user User

	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("confluence: couldn't parse json response: %w", err)
	}

	return &user, nil
}

func (api *API) GetBlogpostByID(ctx context.Context, opts GetPageByIDQuery) (*Page, error) {
	ep, err := api.getBlogpostByIDEndpoint(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't get single blogpost endpoint: %w", err)
	}

	body, err := api.request(ctx, ep)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't perform request: %w", err)
	}

	var blogpost Page

	if err := json.Unmarshal(body, &blogpost); err != nil {
		return nil, fmt.Errorf("confluence: couldn't parse json response: %w", err)
	}

	return &blogpost, nil
}

func (api *API) GetPageByID(ctx context.Context, opts GetPageByIDQuery) (*Page, error) {
	ep, err := api.getPageByIDEndpoint(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't get single page endpoint: %w", err)
	}

	body, err := api.request(ctx, ep)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't perform request: %w", err)
	}

	var page Page

	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("confluence: couldn't parse json response: %w", err)
	}

	return &page, nil
}

func (api *API) GetFolderByID(ctx context.Context, opts GetFolderByIDQuery) (*Folder, error) {
	ep, err := api.getFolderByIDEndpoint(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't get folder endpoint: %w", err)
	}

	body, err := api.request(ctx, ep)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't perform request: %w", err)
	}

	var folder Folder
	if err := json.Unmarshal(body, &folder); err != nil {
		return nil, fmt.Errorf("confluence: couldn't parse json response: %w", err)
	}
	return &folder, nil
}

func (api *API) GetBlogPosts(ctx context.Context, opts GetPagesQuery) (*MultiPageResponse, error) {
	ep, err := api.getBlogPostsEndpoint(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't get blogposts endpoint: %w", err)
	}

	body, err := api.request(ctx, ep)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't perform request: %w", err)
	}

	var blogpostList MultiPageResponse

	if err := json.Unmarshal(body, &blogpostList); err != nil {
		return nil, fmt.Errorf("confluence: couldn't parse json response: %w", err)
	}

	return &blogpostList, nil
}

func (api *API) GetPages(ctx context.Context, opts GetPagesQuery) (*MultiPageResponse, error) {
	ep, err := api.getPagesEndpoint(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't get pages endpoint: %w", err)
	}

	body, err := api.request(ctx, ep)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't perform request: %w", err)
	}

	var pageList MultiPageResponse

	if err := json.Unmarshal(body, &pageList); err != nil {
		return nil, fmt.Errorf("confluence: couldn't parse json response: %w", err)
	}

	return &pageList, nil
}

func (api *API) getSpaces(ctx context.Context, opts SpacesQuery) (*AllSpaces, error) {
	ep, err := api.getSpaceEndpoint(opts)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't get spaces endpoint: %w", err)
	}

	body, err := api.request(ctx, ep)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't perform request: %w", err)
	}

	var allSpaces AllSpaces

	if err := json.Unmarshal(body, &allSpaces); err != nil {
		return nil, fmt.Errorf("confluence: couldn't parse json response: %w", err)
	}

	return &allSpaces, nil
}

// CurrentUser return current user information
func (api *API) CurrentUser(ctx context.Context) (*User, error) {
	ep, err := api.getCurrentUserEndpoint()
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't get current user endpoint: %w", err)
	}

	body, err := api.request(ctx, ep)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't perform http request: %w", err)
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("confluence: couldn't parse json response: %w", err)
	}

	return &user, nil
}

// Request implements the basic Request function
func (api *API) request(ctx context.Context, url *url.URL) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't instantiate http request: %w", err)
	}

	req.Header.Add("Accept", "application/json, */*")

	// if user & token are not set, do not add authorization header
	if api.username != "" && api.token != "" {
		req.SetBasicAuth(api.username, api.token)
	} else if api.token != "" {
		req.Header.Set("Authorization", "Bearer "+api.token)
	}

	response, err := api.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't perform http request: %w", err)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("confluence: couldn't read http response body: %w", err)
	}

	if err := response.Body.Close(); err != nil {
		return nil, fmt.Errorf("confluence: couldn't close response body: %w", err)
	}

	switch response.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusPartialContent, http.StatusNoContent, http.StatusResetContent:
		return body, nil
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("confluence: authentication failed")
	case http.StatusServiceUnavailable:
		return nil, fmt.Errorf("confluence: service is not available: %s", response.Status)
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("confluence: internal server error: %s", response.Status)
	case http.StatusConflict:
		return nil, fmt.Errorf("confluence: conflict: %s", response.Status)
	}

	return nil, fmt.Errorf("confluence: unknown HTTP response status: %s: %s", response.Status, url.String())
}
