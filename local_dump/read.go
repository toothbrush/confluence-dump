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
	if header == nil {
		return data.LocalMarkdown{}, fmt.Errorf("local_dump: Header seems empty in %s: %w", fullPath, err)
	}

	return data.LocalMarkdown{
		Content:      string(source),
		ID:           fmt.Sprintf("%d", header.ObjectId),
		RelativePath: relativePath,
		Version:      header.Version,
	}, nil
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