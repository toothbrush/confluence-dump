package data

import (
	"fmt"
	"regexp"
	"strings"
)

func canonicalise(title string) (string, error) {
	str := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(title, " ")
	str = strings.ToLower(str)
	str = strings.Join(strings.Fields(str), "-")

	if len(str) > 101 {
		str = str[:100]
	}

	str = strings.Trim(str, "-")

	if len(str) < 2 {
		return "", fmt.Errorf("data: Slug too short: title was '%s'", title)
	}

	return str, nil
}

func BuildCacheFromPagelist(pages []ConfluenceContent) (RemoteContentCache, error) {
	id_title_mapping := make(RemoteContentCache)

	for _, page := range pages {
		slug, err := canonicalise(page.Content.Title)
		if err != nil {
			return nil, fmt.Errorf("data: Couldn't derive slug: %w", err)
		}
		if page.Content.Version == nil {
			return nil, fmt.Errorf("data: Found nil .Version field for Object ID %s", page.Content.ID)
		}
		id_title_mapping[page.Content.ID] = RemoteObjectMetadata{
			ID:       page.Content.ID,
			Title:    page.Content.Title,
			Slug:     slug,
			SpaceKey: page.Space.Space.Key,
			Version:  page.Content.Version.Number,
			Org:      page.Space.Org,
		}
	}

	return id_title_mapping, nil
}
