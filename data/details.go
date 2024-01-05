package data

import (
	"fmt"
	"path"

	conf "github.com/virtomize/confluence-go-api"
)

func PagePath(page conf.Content, remoteContentCache RemoteContentCache) (RelativePath, error) {
	pathParts := []string{}

	for _, ancestor := range page.Ancestors {
		if ancestorMetadata, ok := remoteContentCache[ContentID(ancestor.ID)]; ok {
			pathParts = append(pathParts, ancestorMetadata.Slug)
		} else {
			// oh no, found an ID with no title mapped!!
			// this .. should never happen.  We'll see.
			return "", fmt.Errorf("data: Couldn't retrieve page ID %s from cache", ancestor.ID)
		}
	}

	if pageMetadata, ok := remoteContentCache[ContentID(page.ID)]; ok {
		// if this is a blog post, let's also prepend the author's .. identifier.
		if page.Type == "blogpost" {
			userID, err := userID(page)
			if err != nil {
				return "", fmt.Errorf("data: failed to determine user's identity: %w", err)
			}

			pathParts = append([]string{userID}, pathParts...)
		}

		// prepend space code, e.g. CORE,
		pathParts = append([]string{pageMetadata.SpaceKey}, pathParts...)

		// prepend org slug, e.g. redbubble,
		pathParts = append([]string{pageMetadata.Org}, pathParts...)

		// append my filename, which is <id>-<slug>.md
		pathParts = append(pathParts, fmt.Sprintf("%s-%s.md", pageMetadata.ID, pageMetadata.Slug))
	} else {
		// oh no, our own ID isn't in the mapping?
		return "", fmt.Errorf("data: Couldn't retrieve page ID %s from cache", page.ID)
	}

	return RelativePath(path.Join(pathParts...)), nil
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
