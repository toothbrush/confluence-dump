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

	ancestryEqual, err := downloader.ancestryEqual(remoteAncestry, localAncestry)
	if err != nil {
		return LocalMarkdown{}, false, fmt.Errorf("localdump: error comparing ancestry: %w", err)
	}

	// ok, we _are_ aware of it.  how about the version?
	if remote.Page.Version != nil &&
		remote.Page.Version.Number == ourItem.Version &&
		ancestryEqual {
		// oh, we know about it, and it's the same version & ancestry! nothing to do here.
		return ourItem, true, nil
	} else {
		// something has changed.  redownload.
		return LocalMarkdown{}, false, nil
	}
}

func (downloader *SpacesDownloader) ancestryEqual(ancestry1 []ContentID, ancestry2 []ContentID) (bool, error) {
	if len(ancestry1) != len(ancestry2) {
		return false, nil
	}

	// now we know their lengths are equal, so we can safely iterate on the one index.
	for i := range ancestry1 {
		// first, we make sure that the ids of the ancestors are equal:
		ancestor1 := ancestry1[i]
		ancestor2 := ancestry2[i]
		if ancestor1 != ancestor2 {
			return false, nil
		}

		// however, we should refresh sub-pages if an ancestor was updated.  otherwise we'll miss
		// the case where a->b->c and b is renamed, in which case the folders on our local copy
		// should be renamed too.
		//
		// in time, we can probably make this more efficient by precomputing staleness in a DFS
		// way.... oh well.
		remoteAncestorMetadata, ok := downloader.remotePageMetadata[ancestor1]
		if !ok {
			return false, fmt.Errorf("localdump: could not look up remote ancestor metadata: %s", ancestor1)
		}

		localAncestorMetadata, ok := downloader.localMarkdownCache[ancestor1]
		if !ok {
			return false, fmt.Errorf("localdump: could not look up local ancestor cached metadata: %s", ancestor1)
		}

		if remoteAncestorMetadata.Page.Version == nil {
			return false, fmt.Errorf("localdump: unexpected nil .Version for ancestor: %+v", remoteAncestorMetadata)
		}

		if remoteAncestorMetadata.Page.Version.Number != localAncestorMetadata.Version {
			return false, nil
		}
	}

	return true, nil
}
