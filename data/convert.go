package data

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	md_plugin "github.com/JohannesKaufmann/html-to-markdown/plugin"
	conf "github.com/virtomize/confluence-go-api"
	"gopkg.in/yaml.v3"
)

func ConvertToMarkdown(content *conf.Content, metadata_cache RemoteContentCache) (LocalMarkdown, error) {
	converter := md.NewConverter("", true, nil)
	// Github flavoured Markdown knows about tables 👍
	converter.Use(md_plugin.GitHubFlavored())
	if content.Body.View == nil {
		return LocalMarkdown{}, fmt.Errorf("data: Found nil .Body.View field for Object ID %s", content.ID)
	}
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
		ancestor_metadata, ok := metadata_cache[ContentID(ancestor.ID)]
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
	if content.Version == nil {
		return LocalMarkdown{}, fmt.Errorf("data: Found nil .Version field for Object ID %s", content.ID)
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

	// maybe one day consider storing entire header in LocalMarkdown object, and putting the "write
	// to file" logic elsewhere.
	body := fmt.Sprintf(`---
%s
---
%s
`,
		strings.TrimSpace(string(yaml_header)),
		markdown)

	relativeOutputPath, err := PagePath(*content, metadata_cache)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Couldn't determine page path: %w", err)
	}

	return LocalMarkdown{
		ID:           ContentID(content.ID),
		Content:      body,
		RelativePath: RelativePath(relativeOutputPath),
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
