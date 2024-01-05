package local_dump

import (
	"fmt"

	"github.com/toothbrush/confluence-dump/data"
)

func LocalPageIsStale(content data.ConfluenceContent, remote_cache data.RemoteContentCache, local_files data.LocalMarkdownCache) (bool, error) {
	id := content.Content.ID

	if remote, ok := remote_cache[id]; ok {
		if our_item, ok := local_files[id]; ok {

			remote_item_target_path, err := data.PagePath(content.Content, remote_cache)
			if err != nil {
				return false, fmt.Errorf("local_dump: Couldn't determine target path for object ID %s: %w", id, err)
			}

			// ok, we _are_ aware of it.  how about the version?
			if (our_item.Version == remote.Version) && (our_item.RelativePath == remote_item_target_path) {
				// oh, we know about it, and it's the same version! nothing to do here.
				return false, nil
			} else {
				// it's a different version or wrong file path.  redownload.
				return true, nil
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
