package confluence_api

import (
	"fmt"
	"os"

	"github.com/toothbrush/confluence-dump/data"
	"github.com/toothbrush/confluence-dump/local_dump"
	conf "github.com/virtomize/confluence-go-api"
)

// XXX(pd) 20240104: Hmm, this is a deprecated API? (seen in VCR recording)
func GetAllPagesByQuery(api conf.API, query conf.ContentQuery) ([]conf.Content, error) {
	more := true
	contents := []conf.Content{}
	position := 0
	query_dup := query

	for more {
		query_dup.Start = position
		res, err := api.GetContent(query_dup)
		if err != nil {
			return []conf.Content{}, fmt.Errorf("confluence_api: couldn't retrieve list of contents: %w", err)
		}

		position += res.Size
		more = res.Size > 0

		if more {
			for _, res := range res.Results {
				contents = append(contents, res)
			}
			fmt.Fprintf(os.Stderr, "Fetched %d items...\n", position)
		}
	}

	return contents, nil
}

func GetAllPagesInSpace(api conf.API, space data.ConfluenceSpace) ([]data.ConfluenceContent, error) {
	// get content (just metadata) by space name
	contents := []data.ConfluenceContent{}

	query := conf.ContentQuery{Expand: []string{"version"}}
	if space.Space.Key == "blogposts" {
		// whoops, blogposts are special, they're not in a "space"
		query.Type = "blogpost"
	} else {
		// just a boring Confluence space
		query.SpaceKey = space.Space.Key
	}

	results, err := GetAllPagesByQuery(api, query)
	if err != nil {
		return []data.ConfluenceContent{}, fmt.Errorf("confluence_api: Failed to retrieve space '%s' contents: %w", space.Space.Key, err)
	}

	for _, res := range results {
		contents = append(contents, data.ConfluenceContent{
			Content: res,
			Space:   space,
		})
	}

	return contents, nil
}

func DownloadIfChanged(always_download bool, api conf.API, content data.ConfluenceContent, remote_title_cache data.RemoteContentCache, local_cache data.LocalMarkdownCache, storePath string) error {
	stale, err := local_dump.LocalPageIsStale(content, remote_title_cache, local_cache)
	if err != nil {
		return fmt.Errorf("confluence_api: Staleness check failed: %w", err)
	}

	if !stale {
		if always_download {
			fmt.Fprintf(os.Stderr, "Page %s is up-to-date, redownloading anyway because always-download=true...\n", content.Content.ID)
		} else {
			if our_item, ok := local_cache[data.ContentID(content.Content.ID)]; ok {
				fmt.Fprintf(os.Stderr, "Page %s is up-to-date in '%s', skipping...\n", content.Content.ID, our_item.RelativePath)
				// early return :/
				return nil
			}
		}
	}

	c, err := RetrieveContentByID(api, content.Space, content.Content.ID)
	if err != nil {
		return fmt.Errorf("confluence_api: Confluence download failed: %w", err)
	}

	markdown, err := data.ConvertToMarkdown(&c.Content, remote_title_cache)
	if err != nil {
		return fmt.Errorf("confluence_api: Convert to Markdown failed: %w", err)
	}

	if err = local_dump.WriteMarkdownIntoLocal(storePath, markdown); err != nil {
		return fmt.Errorf("confluence_api: Write to file failed: %w", err)
	}

	return nil
}

func RetrieveContentByID(api conf.API, space data.ConfluenceSpace, id string) (*data.ConfluenceContent, error) {
	content, err := api.GetContentByID(id, conf.ContentQuery{
		Expand: []string{"ancestors", "body.view", "links", "version"},
	})
	if err != nil {
		return &data.ConfluenceContent{}, fmt.Errorf("confluence_api: couldn't retrieve object id %s: %w", id, err)
	}

	return &data.ConfluenceContent{
		Content: *content,
		Space:   space,
	}, nil
}

func ListAllSpaces(api conf.API) (map[string]data.ConfluenceSpace, error) {
	more := true
	position := 0
	spaces := map[string]data.ConfluenceSpace{}

	for more {
		allspaces, err := api.GetAllSpaces(conf.AllSpacesQuery{
			Type:  "global",
			Start: position,
			Limit: 10,
		})

		if err != nil {
			return map[string]data.ConfluenceSpace{}, fmt.Errorf("confluence_api: couldn't list spaces: %w", err)
		}

		position += allspaces.Size
		more = allspaces.Size > 0

		if more {
			for _, space := range allspaces.Results {
				spaces[space.Key] = data.ConfluenceSpace{
					Space: space,
					Org:   "foo",
				}
			}
			fmt.Fprintf(os.Stderr, "Found %d spaces...\n", position)
		}
	}

	return spaces, nil
}
