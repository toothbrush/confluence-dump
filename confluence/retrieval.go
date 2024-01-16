package confluence

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

func (api API) ListAllSpaces(ctx context.Context, orgName string) (map[string]Space, error) {
	spaces := map[string]Space{}

	query := SpacesQuery{
		// TODO: feature: do something sensible with "personal spaces"
		Type:  "global", // can be "personal", too... Or nil for both!
		Limit: 10,
	}

	for {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		allspaces, err := api.getSpaces(ctx, query)

		if err != nil {
			return nil, fmt.Errorf("confluence: couldn't list spaces: %w", err)
		}

		for _, space := range allspaces.Results {
			spaces[space.Key] = Space{
				ID:     space.ID,
				Key:    space.Key,
				Name:   space.Name,
				Type:   space.Type,
				Status: space.Status,
				Org:    orgName,
			}
		}

		if allspaces.Links.Next == "" {
			break
		} else {
			q, err := url.Parse(allspaces.Links.Next)
			if err != nil {
				return nil, fmt.Errorf("confluence: couldn't parse _links.next: %w", err)
			}
			query.Cursor = q.Query().Get("cursor")
			if query.Cursor == "" {
				return nil, fmt.Errorf("confluence: expected parameter 'cursor' was empty")
			}
		}
	}

	return spaces, nil
}
