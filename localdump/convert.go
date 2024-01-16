package localdump

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	mdplugin "github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/toothbrush/confluence-dump/confluence"
	"gopkg.in/yaml.v3"
)

func (downloader *SpacesDownloader) ConvertToMarkdown(content *confluence.Page) (LocalMarkdown, error) {
	converter := md.NewConverter("", true, nil)
	// Github flavoured Markdown knows about tables 👍
	converter.Use(mdplugin.GitHubFlavored())
	if content.Body.View == nil {
		return LocalMarkdown{}, fmt.Errorf("localdump: found nil .Body.View field for Object ID %s", content.ID)
	}

	markdown, err := converter.ConvertString(content.Body.View.Value)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("localdump: failed to convert to Markdown: %w", err)
	}
	itemWebURI := downloader.API.BaseURI.String() + content.Links.WebUI
	if _, err := url.Parse(itemWebURI); err != nil {
		return LocalMarkdown{}, fmt.Errorf("localdump: generated URL is bunk: %w", err)
	}

	// Are we able to set a base for all URLs?  Currently the Markdown has things like
	// '/wiki/spaces/DRE/pages/2946695376/Tools+and+Infrastructure' which are a bit un ergonomic.
	// we could (fancy mode) resolve to a link in the local dump or (grug mode) just add the
	// https://redbubble.atlassian.net base URL.
	ancestorNames := []string{}
	ancestorIDs := []int{}
	pageMetadata, ok := downloader.remotePageMetadata[ContentID(content.ID)]
	if !ok {
		return LocalMarkdown{}, fmt.Errorf("localdump: missing ancestry data: %w", err)
	}

	for _, ancestor := range pageMetadata.AncestorIDs {
		ancestorID, err := strconv.Atoi(string(ancestor))
		if err != nil {
			return LocalMarkdown{}, fmt.Errorf("localdump: object ID %s not an int: %w", ancestor, err)
		}
		ancestorIDs = append(ancestorIDs, ancestorID)

		ancestorMetadata, ok := downloader.remotePageMetadata[ancestor]
		if !ok {
			// oh no, found an ID with no title mapped!!
			return LocalMarkdown{}, fmt.Errorf("localdump: found an ID reference we haven't seen before! %s", ancestor)
		}
		ancestorNames = append(ancestorNames, ancestorMetadata.Page.Title)
	}
	id, err := strconv.Atoi(content.ID)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("localdump: object ID %s not an int: %w", content.ID, err)
	}
	if content.Version == nil {
		return LocalMarkdown{}, fmt.Errorf("localdump: found nil .Version field for object %s", content.ID)
	}

	timestamp, err := time.Parse(time.RFC3339, content.Version.CreatedAt)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("localdump: couldn't parse timestamp %s: %w", content.Version.CreatedAt, err)
	}

	header := MarkdownHeader{
		Title:         content.Title,
		Timestamp:     timestamp,
		Version:       content.Version.Number,
		ObjectID:      id,
		URI:           itemWebURI,
		Status:        content.Status,
		ObjectType:    content.ContentType.String(),
		AncestorNames: ancestorNames,
		AncestorIDs:   ancestorIDs,
	}

	yamlHeader, err := yaml.Marshal(header)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("localdump: Couldn't marshal header YAML: %w", err)
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

	relativeOutputPath, err := downloader.PagePath(*content)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("localdump: Couldn't determine page path: %w", err)
	}

	return LocalMarkdown{
		ID:           ContentID(content.ID),
		Content:      body,
		RelativePath: RelativePath(relativeOutputPath),
	}, nil
}
