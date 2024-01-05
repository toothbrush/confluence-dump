package localdump

import (
	"fmt"

	"github.com/toothbrush/confluence-dump/data"
)

func LocalPageIsStale(content data.ConfluenceContent, remoteCache data.RemoteContentCache, localFiles data.LocalMarkdownCache) (bool, error) {
	id := data.ContentID(content.Content.ID)

	if remote, ok := remoteCache[id]; ok {
		if ourItem, ok := localFiles[id]; ok {
			// ok, we _are_ aware of it.  how about the version?
			if ourItem.Version == remote.Version {
				// oh, we know about it, and it's the same version! nothing to do here.
				return false, nil
			} else {
				// it's a different version.  redownload.
				return true, nil
			}
		} else {
			// we don't have the remote item at all -- add it to the download list
			return true, nil
		}
	} else {
		// hmmm asking us about a thing we're not aware of!
		return false, fmt.Errorf("localdump: Queried LocalPageIsStale for invalid remote ID %s", id)
	}
}
