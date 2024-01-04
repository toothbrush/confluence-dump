package local_dump

import (
	"github.com/toothbrush/confluence-dump/data"
)

func ChangedPages(remote_cache data.MetadataCache, local_files data.LocalMarkdownLookup) []string {
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
