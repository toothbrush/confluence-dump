package data

import (
	"fmt"
	"path"

	conf "github.com/virtomize/confluence-go-api"
)

func pagePath(page conf.Content, id_to_slug RemoteContentCache) (string, error) {
	path_parts := []string{}

	for _, ancestor := range page.Ancestors {
		if ancestor_metadata, ok := id_to_slug[ancestor.ID]; ok {
			path_parts = append(path_parts, ancestor_metadata.Slug)
		} else {
			// oh no, found an ID with no title mapped!!
			// this .. should never happen.  We'll see.
			return "", fmt.Errorf("data: Couldn't retrieve page ID %s from cache", ancestor.ID)
		}
	}

	if page_metadata, ok := id_to_slug[page.ID]; ok {
		// prepend space code, e.g. CORE,
		path_parts = append([]string{page_metadata.SpaceKey}, path_parts...)
		// append my filename, which is <slug>.md
		path_parts = append(path_parts, fmt.Sprintf("%s.md", page_metadata.Slug))
	} else {
		// oh no, our own ID isn't in the mapping?
		return "", fmt.Errorf("data: Couldn't retrieve page ID %s from cache", page.ID)
	}

	return path.Join(path_parts...), nil
}
