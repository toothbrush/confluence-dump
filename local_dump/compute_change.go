package local_dump

import (
	"fmt"
	"os"

	"github.com/toothbrush/confluence-dump/data"
)

func LocalPageIsStale(content data.ConfluenceContent, remote_cache data.RemoteContentCache, local_files data.LocalMarkdownCache) (bool, error) {
	id := data.ContentID(content.Content.ID)
	fmt.Printf("LocalPageIsStale for id = %s\n", id)

	if remote, ok := remote_cache[id]; ok {
		fmt.Printf("querying local_files for %s\n", id)
		if our_item, ok := local_files[id]; ok {
			fmt.Printf("our_item = %s, %d, %s\n", our_item.ID, our_item.Version, our_item.RelativePath)
			// ok, we _are_ aware of it.  how about the version?
			if our_item.Version == remote.Version {
				// oh, we know about it, and it's the same version! nothing to do here.
				return false, nil
			} else {
				// it's a different version.  redownload.
				fmt.Fprintf(os.Stderr, "present: staleness debugging: %d, %d, %s\n", our_item.Version, remote.Version, our_item.RelativePath)
				return true, nil
			}
		} else {
			// we don't have the remote item at all -- add it to the download list
			fmt.Fprintf(os.Stderr, "absent: staleness debugging: %d, %d, %s\n", our_item.Version, remote.Version, our_item.RelativePath)
			return true, nil
		}
	} else {
		// hmmm asking us about a thing we're not aware of!
		return false, fmt.Errorf("local_dump: Queried LocalPageIsStale for invalid remote ID %s", id)
	}
}
