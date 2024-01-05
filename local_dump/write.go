package local_dump

import (
	"fmt"
	"os"
	"path"

	"github.com/toothbrush/confluence-dump/data"
)

func WriteMarkdownIntoLocal(storePath string, contents data.LocalMarkdown) error {
	// Does local repo exist?
	stat, err := os.Stat(storePath)
	if err != nil {
		return fmt.Errorf("local_dump: Cannot stat '%s': %w", storePath, err)
	}

	if !stat.IsDir() {
		// path is not a directory.  this is bad, we should bail
		return fmt.Errorf("local_dump: Local store path not a directory: '%s'", storePath)
	}

	// construct destination path
	abs := path.Join(storePath, string(contents.RelativePath))
	directory := path.Dir(abs)
	// there's probably a nicer way to express 0755 but meh
	if err = os.MkdirAll(directory, 0755); err != nil {
		return fmt.Errorf("local_dump: Couldn't create directory %s: %w", directory, err)
	}

	fmt.Printf("Writing page %s to: %s...\n", contents.ID, abs)

	f, err := os.Create(abs)
	if err != nil {
		return fmt.Errorf("local_dump: Couldn't create file %s: %w", abs, err)
	}

	defer f.Close()
	if _, err = f.WriteString(contents.Content); err != nil {
		return fmt.Errorf("local_dump: Couldn't write to file %s: %w", abs, err)
	}

	return nil
}
