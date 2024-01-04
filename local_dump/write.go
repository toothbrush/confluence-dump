package local_dump

import (
	"fmt"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"

	"github.com/toothbrush/confluence-dump/data"
)

// TODO extract this into cobra global config
const REPO_BASE = "~/confluence"

func WriteMarkdownIntoLocal(contents data.LocalMarkdown) error {
	// Does REPO_BASE exist?
	expanded_repo_base, err := homedir.Expand(REPO_BASE)
	if err != nil {
		return fmt.Errorf("local_dump: Couldn't expand '%s': %w", REPO_BASE, err)
	}
	stat, err := os.Stat(expanded_repo_base)
	if err != nil {
		return fmt.Errorf("local_dump: Cannot stat '%s': %w", expanded_repo_base, err)
	}

	if !stat.IsDir() {
		// path is not a directory.  this is bad, we should bail
		return fmt.Errorf("local_dump: REPO_BASE not a directory: '%s'", expanded_repo_base)
	}

	// construct destination path
	abs := path.Join(expanded_repo_base, contents.RelativePath)
	directory := path.Dir(abs)

	fmt.Printf("Writing page %s to: %s...\n", contents.ID, path.Join(REPO_BASE, contents.RelativePath))
	// there's probably a nicer way to express 0755 but meh
	if err = os.MkdirAll(directory, 0755); err != nil {
		return fmt.Errorf("local_dump: Couldn't create directory %s: %w", directory, err)
	}

	f, err := os.Create(abs)
	if err != nil {
		return fmt.Errorf("local_dump: Couldn't create file %s: %w", abs, err)
	}

	defer f.Close()
	f.WriteString(contents.Content)

	return nil
}
