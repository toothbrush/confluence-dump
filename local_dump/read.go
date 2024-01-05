package local_dump

import (
	"bytes"
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
		return data.LocalMarkdown{}, fmt.Errorf("local_dump: Couldn't read file %s: %w", fullPath, err)
	}

	d := yaml.NewDecoder(bytes.NewReader(source))
	header := new(data.MarkdownHeader)

	// we expect the first "document" to be our header YAML.
	if err := d.Decode(&header); err != nil {
		return data.LocalMarkdown{}, fmt.Errorf("local_dump: Couldn't parse header of file %s: %w", fullPath, err)
	}
	// check it was parsed
	if header == nil && header.ObjectId > 0 && header.Version > 0 {
		return data.LocalMarkdown{}, fmt.Errorf("local_dump: Header seems broken in %s", fullPath)
	}

	return data.LocalMarkdown{
		Content:      string(source),
		ID:           data.ContentID(fmt.Sprintf("%d", header.ObjectId)),
		RelativePath: data.RelativePath(relativePath),
		Version:      header.Version,
	}, nil
}

func LoadLocalMarkdown(storePath string) (data.LocalMarkdownCache, error) {
	// find files
	filenames, err := ListAllMarkdownFiles(storePath)
	if err != nil {
		return data.LocalMarkdownCache{}, fmt.Errorf("local_dump: Error loading Markdown files: %w", err)
	}

	local_markdown := data.LocalMarkdownCache{}
	// parse each file
	for _, file := range filenames {
		rel, err := filepath.Rel(storePath, file)
		if err != nil {
			return data.LocalMarkdownCache{}, fmt.Errorf("local_dump: Couldn't compute relative path of %s: %w", file, err)
		}

		md, err := ParseExistingMarkdown(storePath, rel)
		if err != nil {
			return data.LocalMarkdownCache{}, fmt.Errorf("local_dump: Couldn't load local Markdown file %s: %w", file, err)
		}

		if _, ok := local_markdown[md.ID]; ok {
			// oh damn, we have two or more files in the local repo that present the same ID!  warn the user.
			fmt.Fprintf(os.Stderr, "🚨 WARNING: Found duplicate id %s in file %s!  Undefined behaviour will result.\n", md.ID, md.RelativePath)
		}
		local_markdown[md.ID] = md
	}

	return local_markdown, nil
}

func ListAllMarkdownFiles(storePath string) ([]string, error) {
	filenames := []string{}
	err := filepath.Walk(storePath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("local_dump: Error walking file tree: %w", err)
			}
			if !info.IsDir() && strings.HasSuffix(path, ".md") {
				filenames = append(filenames, path)
			}
			return nil
		})
	if err != nil {
		return []string{}, fmt.Errorf("local_dump: Error walking file tree: %w", err)
	}

	return filenames, nil
}
