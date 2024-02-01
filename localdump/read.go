package localdump

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

func ParseExistingMarkdown(storePath string, relativePath string) (LocalMarkdown, error) {
	fullPath := path.Join(storePath, relativePath)
	source, err := os.ReadFile(fullPath)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("localdump: couldn't read file %s: %w", fullPath, err)
	}

	d := yaml.NewDecoder(bytes.NewReader(source))
	var header MarkdownHeader

	// we expect the first "document" to be our header YAML.
	if err := d.Decode(&header); err != nil {
		return LocalMarkdown{}, fmt.Errorf("localdump: couldn't parse header of file %s: %w", fullPath, err)
	}
	// check it was parsed
	if header.ObjectID < 1 ||
		header.Version < 1 {
		return LocalMarkdown{}, fmt.Errorf("localdump: header seems broken in %s", fullPath)
	}

	ancestorIDs := []ContentID{}
	for _, id := range header.AncestorIDs {
		ancestorIDs = append(ancestorIDs, ContentID(fmt.Sprintf("%d", id)))
	}

	return LocalMarkdown{
		Content:      string(source),
		ID:           ContentID(fmt.Sprintf("%d", header.ObjectID)),
		RelativePath: RelativePath(relativePath),
		Version:      header.Version,
		AncestorIDs:  ancestorIDs,
	}, nil
}

func (downloader *SpacesDownloader) LoadLocalMarkdown() error {
	// find files
	filenames, err := ListAllMarkdownFiles(downloader.StorePath)
	if err != nil {
		return fmt.Errorf("localdump: error loading Markdown files: %w", err)
	}

	downloader.localMarkdownCache = make(map[ContentID]LocalMarkdown)
	// parse each file
	for _, file := range filenames {
		rel, err := filepath.Rel(downloader.StorePath, file)
		if err != nil {
			return fmt.Errorf("localdump: couldn't compute relative path of %s: %w", file, err)
		}

		md, err := ParseExistingMarkdown(downloader.StorePath, rel)
		if err != nil {
			return fmt.Errorf("localdump: couldn't load local Markdown file %s: %w", file, err)
		}

		if _, ok := downloader.localMarkdownCache[md.ID]; ok {
			// oh damn, we have two or more files in the local repo that present the same ID!  warn the user.
			return fmt.Errorf("localdump: found duplicate id %s in file %s, please clean up", md.ID, md.RelativePath)
		}
		downloader.localMarkdownCache[md.ID] = md
	}

	return nil
}

// returns absolute pathnames
func ListAllMarkdownFiles(inFolder string) ([]string, error) {
	if _, err := os.Stat(inFolder); err == nil {
		// path/to/whatever exists
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist; this might mean this is the first time running.
		return nil, nil
	} else {
		// some other error
		return nil, fmt.Errorf("localdump: error opening %s for file tree walk: %w", inFolder, err)
	}

	filenames := []string{}

	err := filepath.Walk(inFolder,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("localdump: error during file tree walk: %w", err)
			}
			if !info.IsDir() && strings.HasSuffix(path, ".md") {
				filenames = append(filenames, path)
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("localdump: error initialising file tree walk: %w", err)
	}

	return filenames, nil
}

type MarkdownHeader struct {
	Title         string
	Timestamp     time.Time
	Version       int
	Author        string
	ObjectID      int `yaml:"object_id"`
	URI           string
	Status        string
	ObjectType    string   `yaml:"object_type"`
	AncestorNames []string `yaml:"ancestor_names,flow"`
	AncestorIDs   []int    `yaml:"ancestor_ids,flow"`
}
