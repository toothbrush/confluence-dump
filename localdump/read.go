package localdump

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/toothbrush/confluence-dump/data"
	"gopkg.in/yaml.v2"
)

func ParseExistingMarkdown(storePath string, relativePath string) (data.LocalMarkdown, error) {
	fullPath := path.Join(storePath, relativePath)
	source, err := os.ReadFile(fullPath)
	if err != nil {
		return data.LocalMarkdown{}, fmt.Errorf("localdump: Couldn't read file %s: %w", fullPath, err)
	}

	d := yaml.NewDecoder(bytes.NewReader(source))
	header := new(data.MarkdownHeader)

	// we expect the first "document" to be our header YAML.
	if err := d.Decode(&header); err != nil {
		return data.LocalMarkdown{}, fmt.Errorf("localdump: Couldn't parse header of file %s: %w", fullPath, err)
	}
	// check it was parsed
	if header == nil && header.ObjectId > 0 && header.Version > 0 {
		return data.LocalMarkdown{}, fmt.Errorf("localdump: Header seems broken in %s", fullPath)
	}

	return data.LocalMarkdown{
		Content:      string(source),
		ID:           data.ContentID(fmt.Sprintf("%d", header.ObjectId)),
		RelativePath: data.RelativePath(relativePath),
		Version:      header.Version,
	}, nil
}

// let's scope the markdown-database-loader to a particular space, so that pruning .. makes more
// sense.
func LoadLocalMarkdown(storePath string, space data.ConfluenceSpace) (data.LocalMarkdownCache, error) {
	pathForSpace := path.Join(storePath, space.Org, space.Space.Key)
	// find files
	filenames, err := ListAllMarkdownFiles(pathForSpace)
	if err != nil {
		return data.LocalMarkdownCache{}, fmt.Errorf("localdump: Error loading Markdown files: %w", err)
	}

	local_markdown := data.LocalMarkdownCache{}
	// parse each file
	for _, file := range filenames {
		rel, err := filepath.Rel(storePath, file)
		if err != nil {
			return data.LocalMarkdownCache{}, fmt.Errorf("localdump: Couldn't compute relative path of %s: %w", file, err)
		}

		md, err := ParseExistingMarkdown(storePath, rel)
		if err != nil {
			return data.LocalMarkdownCache{}, fmt.Errorf("localdump: Couldn't load local Markdown file %s: %w", file, err)
		}

		if _, ok := local_markdown[md.ID]; ok {
			// oh damn, we have two or more files in the local repo that present the same ID!  warn the user.
			fmt.Fprintf(os.Stderr, "🚨 WARNING: Found duplicate id %s in file %s!  Undefined behaviour will result.\n", md.ID, md.RelativePath)
		}
		local_markdown[md.ID] = md
	}

	return local_markdown, nil
}

// returns absolute pathnames
func ListAllMarkdownFiles(inFolder string) ([]string, error) {
	if _, err := os.Stat(inFolder); err == nil {
		// path/to/whatever exists
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist; this might mean this is the first time running.
		return []string{}, nil
	} else {
		// some other error
		return []string{}, fmt.Errorf("localdump: Error opening %s for file tree walk: %w", inFolder, err)
	}

	filenames := []string{}

	err := filepath.Walk(inFolder,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("localdump: Error during file tree walk: %w", err)
			}
			if !info.IsDir() && strings.HasSuffix(path, ".md") {
				filenames = append(filenames, path)
			}
			return nil
		})
	if err != nil {
		return []string{}, fmt.Errorf("localdump: Error initialising file tree walk: %w", err)
	}

	return filenames, nil
}
