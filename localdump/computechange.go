package localdump

import (
	"fmt"
)

// Returns the local item that matches the remote, or nil if our local copy is nonexistent or stale.
func (downloader *SpacesDownloader) LocalVersionIsRecent(pageID ContentID) (LocalMarkdown, bool, error) {
	remote, ok := downloader.remotePageMetadata[pageID]
	if !ok {
		// hmmm asking us about a thing we're not aware of!
		return LocalMarkdown{}, false, fmt.Errorf("localdump: remote cache queried about unknown item ID: %s", pageID)
	}

	ourItem, ok := downloader.localMarkdownCache[pageID]
	if !ok {
		// we don't have the remote item at all -- add it to the download list
		return LocalMarkdown{}, false, nil
	}

	remoteAncestry := remote.AncestorIDs
	localAncestry := ourItem.AncestorIDs

	// ok, we _are_ aware of it.  how about the version?
	if remote.Page.Version != nil &&
		remote.Page.Version.Number == ourItem.Version &&
		ancestryEqual(remoteAncestry, localAncestry) {
		// oh, we know about it, and it's the same version & ancestry! nothing to do here.
		return ourItem, true, nil
	} else {
		// something has changed.  redownload.
		return LocalMarkdown{}, false, nil
	}
}

func ancestryEqual(ancestry1 []ContentID, ancestry2 []ContentID) bool {
	if len(ancestry1) != len(ancestry2) {
		return false
	}

	// now we know their lengths are equal.
	for i := range ancestry1 {
		if ancestry1[i] != ancestry2[i] {
			return false
		}
	}

	return true
}
