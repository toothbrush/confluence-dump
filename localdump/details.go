package localdump

import (
	"fmt"
	"path"

	"github.com/toothbrush/confluence-dump/confluence"
)

func (downloader *SpacesDownloader) PagePath(page confluence.Page) (RelativePath, error) {
	pathParts := []string{}
	pageMetadata, ok := downloader.remotePageMetadata[ContentID(page.ID)]
	if !ok {
		return "", fmt.Errorf("localdump: missing ancestry data: %s", page.ID)
	}

	for _, ancestorID := range pageMetadata.AncestorIDs {
		if ancestorMetadata, ok := downloader.remotePageMetadata[ancestorID]; ok {
			pathParts = append(pathParts, ancestorMetadata.Slug)
		} else {
			// oh no, found an ID with no title mapped!!
			// this .. should never happen.  We'll see.
			return "", fmt.Errorf("localdump: couldn't retrieve page ID %s from cache", ancestorID)
		}
	}

	// if this is a blog post, let's also prepend the author's .. identifier.
	if page.ContentType == confluence.BlogContent {
		userID, err := downloader.userID(page)
		if err != nil {
			return "", fmt.Errorf("localdump: failed to determine user's identity: %w", err)
		}

		pathParts = append([]string{userID}, pathParts...)
	}

	// prepend space code, e.g. CORE,
	if page.SpaceKey == "" {
		return "", fmt.Errorf("localdump: empty Space key for item: %s", page.ID)
	}
	pathParts = append([]string{page.SpaceKey}, pathParts...)

	// prepend org slug, e.g. redbubble,
	if page.Org == "" {
		return "", fmt.Errorf("localdump: empty Org key for item: %s", page.ID)
	}
	pathParts = append([]string{page.Org}, pathParts...)

	slug, err := canonicalise(page.Title)
	if err != nil {
		return "", fmt.Errorf("localdump: could not canonicalise title: %w", err)
	}

	// append my filename, which is <id>-<slug>.md
	pathParts = append(pathParts, fmt.Sprintf("%s-%s.md", page.ID, slug))

	return RelativePath(path.Join(pathParts...)), nil
}

func (downloader *SpacesDownloader) userID(page confluence.Page) (string, error) {
	authorID := page.AuthorID
	if authorID == "" {
		return "", fmt.Errorf("localdump: page .AuthorID blank for item %s", page.ID)
	}

	user, ok := downloader.authorMetadata[authorID]
	if !ok {
		return "", fmt.Errorf("localdump: failed to retrieve author info for %s", authorID)
	}

	if user.DisplayName != "" {
		slug, err := canonicalise(user.DisplayName)
		if err != nil {
			return "", fmt.Errorf("localdump: Failed to canonicalise username '%s': %w", user.DisplayName, err)
		}
		return slug, nil
	}
	if user.Username != "" {
		return user.Username, nil
	}
	if user.AccountID != "" {
		return user.AccountID, nil
	}

	return "", fmt.Errorf("localdump: User has no name or id for item %s", page.ID)
}
