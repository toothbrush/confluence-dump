package data

import (
	"fmt"
	"path"

	conf "github.com/virtomize/confluence-go-api"
)

func PagePath(page conf.Content, remote_content_cache RemoteContentCache) (RelativePath, error) {
	path_parts := []string{}

	for _, ancestor := range page.Ancestors {
		if ancestor_metadata, ok := remote_content_cache[ContentID(ancestor.ID)]; ok {
			path_parts = append(path_parts, ancestor_metadata.Slug)
		} else {
			// oh no, found an ID with no title mapped!!
			// this .. should never happen.  We'll see.
			return "", fmt.Errorf("data: Couldn't retrieve page ID %s from cache", ancestor.ID)
		}
	}

	if page_metadata, ok := remote_content_cache[ContentID(page.ID)]; ok {
		// if this is a blog post, let's also prepend the author's .. identifier.
		if page.Type == "blogpost" {
			userID, err := userID(page)
			if err != nil {
				return "", fmt.Errorf("data: failed to determine user's identity: %w", err)
			}

			path_parts = append([]string{userID}, path_parts...)
		}

		// prepend space code, e.g. CORE,
		path_parts = append([]string{page_metadata.SpaceKey}, path_parts...)

		// prepend org slug, e.g. redbubble,
		path_parts = append([]string{page_metadata.Org}, path_parts...)

		// append my filename, which is <id>-<slug>.md
		path_parts = append(path_parts, fmt.Sprintf("%s-%s.md", page_metadata.ID, page_metadata.Slug))
	} else {
		// oh no, our own ID isn't in the mapping?
		return "", fmt.Errorf("data: Couldn't retrieve page ID %s from cache", page.ID)
	}

	return RelativePath(path.Join(path_parts...)), nil
}

func userID(page conf.Content) (string, error) {
	version := page.Version
	if version == nil {
		return "", fmt.Errorf("data: Page .Version nil for item %s", page.ID)
	}
	user := version.By
	if user == nil {
		return "", fmt.Errorf("data: Page .Version.By nil for item %s", page.ID)
	}
	if user.DisplayName != "" {
		slug, err := canonicalise(user.DisplayName)
		if err != nil {
			return "", fmt.Errorf("data: Failed to canonicalise username '%s': %w", user.DisplayName, err)
		}
		return slug, nil
	}
	if user.Username != "" {
		return user.Username, nil
	}
	if user.AccountID != "" {
		return user.AccountID, nil
	}

	return "", fmt.Errorf("data: User has no name or id for item %s", page.ID)
}
