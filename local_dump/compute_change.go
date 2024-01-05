package local_dump

import (
	"fmt"

	"github.com/toothbrush/confluence-dump/data"
)

func ChangedPages(remote_cache data.RemoteContentCache, local_files data.LocalMarkdownCache) []string {
	changed_pages := []string{}

	for _, remote := range remote_cache {
		if our_item, ok := local_files[remote.ID]; ok {
			// ok, we _are_ aware of it.  how about the version?
			our_version := our_item.Version
			if our_version != remote.Version {
				// it's a different version though.  redownload.
				changed_pages = append(changed_pages, remote.ID)
			}
		} else {
			// we don't have the remote item at all -- add it to the download list
			changed_pages = append(changed_pages, remote.ID)
		}
	}

	return changed_pages
}

func LocalPageIsStale(id string, remote_cache data.RemoteContentCache, local_files data.LocalMarkdownCache) (bool, error) {
	if remote, ok := remote_cache[id]; ok {
		if our_item, ok := local_files[id]; ok {
			// ok, we _are_ aware of it.  how about the version?
			our_version := our_item.Version
			if our_version != remote.Version {
				// it's a different version though.  redownload.
				return true, nil
			} else {
				// oh, we know about it, and it's the same version! nothing to do here.
				return false, nil
			}
		} else {
			// we don't have the remote item at all -- add it to the download list
			return true, nil
		}
	} else {
		// hmmm asking us about a thing we're not aware of!
		return false, fmt.Errorf("local_dump: Queried LocalPageIsStale for invalid remote ID %s", id)
	}
}
