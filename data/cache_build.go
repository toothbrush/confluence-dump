package data

import (
	"fmt"
	"regexp"
	"strings"

	conf "github.com/virtomize/confluence-go-api"
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

func BuildCacheFromPagelist(pages []conf.Content, space_key string) (MetadataCache, error) {
	id_title_mapping := make(MetadataCache)

	for _, page := range pages {
		slug, err := canonicalise(page.Title)
		if err != nil {
			return nil, fmt.Errorf("data: Couldn't derive slug: %w", err)
		}
		if page.Version == nil {
			return nil, fmt.Errorf("data: Found nil .Version field for Object ID %s", page.ID)
		}
		id_title_mapping[page.ID] = RemoteObjectMetadata{
			ID:       page.ID,
			Title:    page.Title,
			Slug:     slug,
			SpaceKey: space_key,
			Version:  page.Version.Number,
		}
	}

	return id_title_mapping, nil
}
