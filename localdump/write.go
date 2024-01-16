package localdump

import (
	"fmt"
	"os"
	"path"
)

func (downloader *SpacesDownloader) WriteMarkdownIntoLocal(contents LocalMarkdown) error {
	// Does local repo exist?
	stat, err := os.Stat(downloader.StorePath)
	if err != nil {
		return fmt.Errorf("localdump: cannot stat '%s': %w", downloader.StorePath, err)
	}

	if !stat.IsDir() {
		// path is not a directory.  this is bad, we should bail
		return fmt.Errorf("localdump: local store path not a directory: '%s'", downloader.StorePath)
	}

	// construct destination path
	abs := path.Join(downloader.StorePath, string(contents.RelativePath))
	directory := path.Dir(abs)

	if !downloader.WriteMarkdown {
		// exit early to dry run
		return nil
	}

	// there's probably a nicer way to express 0750 but meh
	if err = os.MkdirAll(directory, 0750); err != nil {
		return fmt.Errorf("localdump: couldn't create directory %s: %w", directory, err)
	}

	f, err := os.Create(abs)
	if err != nil {
		return fmt.Errorf("localdump: couldn't create file %s: %w", abs, err)
	}

	defer f.Close()
	if _, err = f.WriteString(contents.Content); err != nil {
		return fmt.Errorf("localdump: couldn't write to file %s: %w", abs, err)
	}

	return nil
}
