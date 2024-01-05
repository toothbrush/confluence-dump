package confluenceapi

import (
	"fmt"
	"os"

	"github.com/toothbrush/confluence-dump/data"
	"github.com/toothbrush/confluence-dump/localdump"
	conf "github.com/virtomize/confluence-go-api"
)

// XXX(pd) 20240104: Hmm, this is a deprecated API? (seen in VCR recording)
func GetAllPagesByQuery(api conf.API, query conf.ContentQuery) ([]conf.Content, error) {
	contents := []conf.Content{}
	position := 0

	for {
		query.Start = position
		res, err := api.GetContent(query)
		if err != nil {
			return nil, fmt.Errorf("confluenceapi: couldn't retrieve list of contents: %w", err)
		}

		if res.Size == 0 {
			break
		}

		contents = append(contents, res.Results...)
		position += res.Size
		fmt.Fprintf(os.Stderr, "Fetched %d items...\n", position)
	}

	return contents, nil
}

func GetAllPagesInSpace(api conf.API, space data.ConfluenceSpace) ([]data.ConfluenceContent, error) {
	// get content (just metadata) by space name
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
		return nil, fmt.Errorf("confluenceapi: Failed to retrieve space '%s' contents: %w", space.Space.Key, err)
	}

	contents := []data.ConfluenceContent{}
	for _, res := range results {
		contents = append(contents, data.ConfluenceContent{
			Content: res,
			Space:   space,
		})
	}

	return contents, nil
}

func DownloadIfChanged(alwaysDownload bool, api conf.API, content data.ConfluenceContent, remoteContentCache data.RemoteContentCache, localMarkdownCache data.LocalMarkdownCache, storePath string) error {
	stale, err := localdump.LocalPageIsStale(content, remoteContentCache, localMarkdownCache)
	if err != nil {
		return fmt.Errorf("confluenceapi: Staleness check failed: %w", err)
	}

	if !stale {
		if alwaysDownload {
			fmt.Fprintf(os.Stderr, "Page %s is up-to-date, redownloading anyway because always-download=true...\n", content.Content.ID)
		} else {
			if ourItem, ok := localMarkdownCache[data.ContentID(content.Content.ID)]; ok {
				fmt.Fprintf(os.Stderr, "Page %s is up-to-date in '%s', skipping...\n", content.Content.ID, ourItem.RelativePath)
				// early return :/
				return nil
			}
		}
	}

	c, err := RetrieveContentByID(api, content.Space, content.Content.ID)
	if err != nil {
		return fmt.Errorf("confluenceapi: Confluence download failed: %w", err)
	}

	markdown, err := data.ConvertToMarkdown(&c.Content, remoteContentCache)
	if err != nil {
		return fmt.Errorf("confluenceapi: Convert to Markdown failed: %w", err)
	}

	if err = localdump.WriteMarkdownIntoLocal(storePath, markdown); err != nil {
		return fmt.Errorf("confluenceapi: Write to file failed: %w", err)
	}

	return nil
}

func RetrieveContentByID(api conf.API, space data.ConfluenceSpace, id string) (data.ConfluenceContent, error) {
	content, err := api.GetContentByID(id, conf.ContentQuery{
		Expand: []string{"ancestors", "body.view", "links", "version"},
	})
	if err != nil {
		return data.ConfluenceContent{}, fmt.Errorf("confluenceapi: couldn't retrieve object id %s: %w", id, err)
	}

	return data.ConfluenceContent{
		Content: *content,
		Space:   space,
	}, nil
}

func ListAllSpaces(api conf.API, orgName string) (map[string]data.ConfluenceSpace, error) {
	position := 0
	spaces := map[string]data.ConfluenceSpace{}

	for {
		allspaces, err := api.GetAllSpaces(conf.AllSpacesQuery{
			Type:  "global",
			Start: position,
			Limit: 10,
		})

		if err != nil {
			return nil, fmt.Errorf("confluenceapi: couldn't list spaces: %w", err)
		}

		if allspaces.Size == 0 {
			break
		}

		for _, space := range allspaces.Results {
			spaces[space.Key] = data.ConfluenceSpace{
				Space: space,
				Org:   orgName,
			}
		}
		position += allspaces.Size
		fmt.Fprintf(os.Stderr, "Found %d spaces...\n", position)
	}

	return spaces, nil
}
