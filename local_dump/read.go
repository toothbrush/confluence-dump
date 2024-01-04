package local_dump

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/toothbrush/confluence-dump/data"
	"gopkg.in/yaml.v2"
)

func ParseExistingMarkdown(storePath string, relativePath string) (data.LocalMarkdown, error) {
	fullPath, err := homedir.Expand(path.Join(storePath, relativePath))
	if err != nil {
		return data.LocalMarkdown{}, fmt.Errorf("local_dump: Couldn't expand homedir: %w", err)
	}

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
