package localdump

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/toothbrush/confluence-dump/confluence"
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
		return "", fmt.Errorf("localdump: slug too short: title was '%s'", title)
	}

	return str, nil
}

func (downloader *SpacesDownloader) BuildCacheFromPagelist() error {
	for id, item := range downloader.remotePageMetadata {
		ancestors, err := downloader.determineAncestors(item.Page)
		if err != nil {
			return fmt.Errorf("localdump: couldn't determine ancestry for %s: %w", item.Page.ID, err)
		}

		if entry, ok := downloader.remotePageMetadata[id]; ok {
			slug, err := canonicalise(item.Page.Title)
			if err != nil {
				return fmt.Errorf("localdump: couldn't derive slug: %w", err)
			}
			entry.AncestorIDs = ancestors
			entry.Slug = slug
			downloader.remotePageMetadata[id] = entry
		} else {
			return fmt.Errorf("localdump: expected key %s missing from remotePageMetadata", id)
		}
	}

	return nil
}

func (downloader *SpacesDownloader) determineAncestors(page confluence.Page) ([]ContentID, error) {
	maxDepth := 20
	ancestors := []ContentID{}

	// only current pages have sensible ancestry.
	if page.Status != "current" {
		return ancestors, nil
	}

	currentPage := page

	for i := 0; i < maxDepth; i++ {
		if currentPage.ParentID == "" {
			// done
			return ancestors, nil
		} else {
			// we have another ancestor.  prepend it to the list, then figure out its parents.
			ancestors = append([]ContentID{ContentID(currentPage.ParentID)}, ancestors...)

			// what are the ancestor's parents?
			// ensure it's a valid ID, first!
			ancestor, ok := downloader.remotePageMetadata[ContentID(currentPage.ParentID)]
			if !ok {
				// damn, broken reference!
				return nil, fmt.Errorf("localdump: ancestor %s of page %s doesn't exist", currentPage.ParentID, currentPage.ID)
			}
			// cool it's legit
			currentPage = ancestor.Page
		}
	}

	return nil, fmt.Errorf("localdump: exceeded ancestry maximum depth for %s", page.ID)
}
