package localdump

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/toothbrush/confluence-dump/confluence"
)

func (downloader *SpacesDownloader) pruneLocalDB() error {

	for _, s := range downloader.spacesMetadata {
		if err := downloader.pruneSpace(s); err != nil {
			return fmt.Errorf("localdump.pruneLocalDB: failed to prune space %s", s.Key)
		}
	}

	return nil
}

func (downloader *SpacesDownloader) pruneSpace(space confluence.Space) error {
	spaceDir := path.Join(downloader.StorePath, space.Org, space.Key)

	localFiles, err := ListAllMarkdownFiles(spaceDir)
	if err != nil {
		return fmt.Errorf("localdump.pruneSpace: failed to list *.md in: %s", localFiles)
	}

	for _, file := range localFiles {
		relative, err := filepath.Rel(downloader.StorePath, file)
		if err != nil {
			return fmt.Errorf("localdump.pruneSpace: failed to get relative path: %w", err)
		}

		if _, ok := downloader.freshLocalFiles[relative]; ok {
			// file is fresh, skip!
			continue
		}

		// if we're here, it's a stale/unknown file.
		downloader.Logger.Printf("Pruning: %s\n", relative)
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("localdump.pruneSpace: failed to delete: %w", err)
		}
	}

	return nil
}
