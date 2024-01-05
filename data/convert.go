package data

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	mdplugin "github.com/JohannesKaufmann/html-to-markdown/plugin"
	conf "github.com/virtomize/confluence-go-api"
	"gopkg.in/yaml.v3"
)

func ConvertToMarkdown(content *conf.Content, metadataCache RemoteContentCache) (LocalMarkdown, error) {
	converter := md.NewConverter("", true, nil)
	// Github flavoured Markdown knows about tables 👍
	converter.Use(mdplugin.GitHubFlavored())
	if content.Body.View == nil {
		return LocalMarkdown{}, fmt.Errorf("data: Found nil .Body.View field for Object ID %s", content.ID)
	}
	markdown, err := converter.ConvertString(content.Body.View.Value)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Failed to convert to Markdown: %w", err)
	}
	itemWebURI := content.Links.Base + content.Links.WebUI

	// Are we able to set a base for all URLs?  Currently the Markdown has things like
	// '/wiki/spaces/DRE/pages/2946695376/Tools+and+Infrastructure' which are a bit un ergonomic.
	// we could (fancy mode) resolve to a link in the local dump or (grug mode) just add the
	// https://redbubble.atlassian.net base URL.
	ancestorNames := []string{}
	ancestorIDs := []int{}
	for _, ancestor := range content.Ancestors {
		ancestorMetadata, ok := metadataCache[ContentID(ancestor.ID)]
		if ok {
			ancestorNames = append(ancestorNames, ancestorMetadata.Title)

			ancestorID, err := strconv.Atoi(ancestor.ID)
			if err != nil {
				return LocalMarkdown{}, fmt.Errorf("data: Object ID %s not an int: %w", ancestor.ID, err)
			}
			ancestorIDs = append(ancestorIDs, ancestorID)
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
		ObjectID:      id,
		URI:           itemWebURI,
		Status:        content.Status,
		ObjectType:    content.Type,
		AncestorNames: ancestorNames,
		AncestorIDs:   ancestorIDs,
	}

	yamlHeader, err := yaml.Marshal(header)
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
		strings.TrimSpace(string(yamlHeader)),
		markdown)

	relativeOutputPath, err := PagePath(*content, metadataCache)
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
	ObjectID      int `yaml:"object_id"`
	URI           string
	Status        string
	ObjectType    string   `yaml:"object_type"`
	AncestorNames []string `yaml:"ancestor_names,flow"`
	AncestorIDs   []int    `yaml:"ancestor_ids,flow"`
}
