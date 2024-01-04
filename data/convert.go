package data

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	md_plugin "github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/mitchellh/go-homedir"
	conf "github.com/virtomize/confluence-go-api"
	"gopkg.in/yaml.v3"
)

func ConvertToMarkdown(content *conf.Content, metadata_cache MetadataCache) (LocalMarkdown, error) {
	converter := md.NewConverter("", true, nil)
	// Github flavoured Markdown knows about tables 👍
	converter.Use(md_plugin.GitHubFlavored())
	markdown, err := converter.ConvertString(content.Body.View.Value)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Failed to convert to Markdown: %w", err)
	}
	item_web_uri := content.Links.Base + content.Links.WebUI

	// Are we able to set a base for all URLs?  Currently the Markdown has things like
	// '/wiki/spaces/DRE/pages/2946695376/Tools+and+Infrastructure' which are a bit un ergonomic.
	// we could (fancy mode) resolve to a link in the local dump or (grug mode) just add the
	// https://redbubble.atlassian.net base URL.
	ancestor_names := []string{}
	ancestor_ids := []int{}
	for _, ancestor := range content.Ancestors {
		ancestor_metadata, ok := metadata_cache[ancestor.ID]
		if ok {
			ancestor_names = append(ancestor_names, ancestor_metadata.Title)

			ancestor_id, err := strconv.Atoi(ancestor.ID)
			if err != nil {
				return LocalMarkdown{}, fmt.Errorf("data: Object ID %s not an int: %w", ancestor.ID, err)
			}
			ancestor_ids = append(ancestor_ids, ancestor_id)
		} else {
			// oh no, found an ID with no title mapped!!
			return LocalMarkdown{}, fmt.Errorf("data: Found an ID reference we haven't seen before! %s", ancestor.ID)
		}
	}
	id, err := strconv.Atoi(content.ID)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Object ID %s not an int: %w", content.ID, err)
	}

	timestamp, err := time.Parse(time.RFC3339, content.Version.When)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Couldn't parse timestamp %s: %w", content.Version.When, err)
	}

	header := MarkdownHeader{
		Title:         content.Title,
		Timestamp:     timestamp,
		Version:       content.Version.Number,
		ObjectId:      id,
		Uri:           item_web_uri,
		Status:        content.Status,
		ObjectType:    content.Type,
		AncestorNames: ancestor_names,
		AncestorIds:   ancestor_ids,
	}

	yaml_header, err := yaml.Marshal(header)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Couldn't marshal header YAML: %w", err)
	}

	body := fmt.Sprintf(`---
%s
---
%s
`,
		strings.TrimSpace(string(yaml_header)),
		markdown)

	relativeOutputPath, err := pagePath(*content, metadata_cache)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Couldn't determine page path: %w", err)
	}

	return LocalMarkdown{
		ID:           content.ID,
		Content:      body,
		RelativePath: relativeOutputPath,
	}, nil
}

func ParseExistingMarkdown(storePath string, relativePath string) (LocalMarkdown, error) {
	fullPath, err := homedir.Expand(path.Join(storePath, relativePath))
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Couldn't expand homedir: %w", err)
	}

	source, err := os.ReadFile(fullPath)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Couldn't read file %s: %w", fullPath, err)
	}

	d := yaml.NewDecoder(bytes.NewReader(source))
	header := new(MarkdownHeader)

	// we expect the first "document" to be our header YAML.
	if err := d.Decode(&header); err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Couldn't parse header of file %s: %w", fullPath, err)
	}
	// check it was parsed
	if header == nil {
		return LocalMarkdown{}, fmt.Errorf("data: Header seems empty in %s: %w", fullPath, err)
	}

	return LocalMarkdown{
		Content:      string(source),
		ID:           fmt.Sprintf("%d", header.ObjectId),
		RelativePath: relativePath,
		Version:      header.Version,
	}, nil
}

type MarkdownHeader struct {
	Title         string
	Timestamp     time.Time
	Version       int
	ObjectId      int `yaml:"object_id"`
	Uri           string
	Status        string
	ObjectType    string   `yaml:"object_type"`
	AncestorNames []string `yaml:"ancestor_names,flow"`
	AncestorIds   []int    `yaml:"ancestor_ids,flow"`
}
