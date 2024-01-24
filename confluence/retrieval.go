package confluence

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

func (api API) ListAllSpaces(ctx context.Context, orgName string, includePersonal bool) (map[string]Space, error) {
	spaces := map[string]Space{}

	query := SpacesQuery{
		Limit: 10,
	}

	if !includePersonal {
		// Logic here is a bit confusing.  The `type` parameter may be "global", "personal", or
		// nothing at all for both.  "global" will return spaces like DRE, CORE, etc., while
		// "personal" returns each user's space.  Leaving it empty gives us everything, so we only
		// set this if we _do not_ intend to include personal spaces in our query.
		query.Type = "global"
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
