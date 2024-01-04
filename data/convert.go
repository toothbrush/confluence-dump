package data

import (
	"fmt"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	md_plugin "github.com/JohannesKaufmann/html-to-markdown/plugin"
	conf "github.com/virtomize/confluence-go-api"
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
	ancestor_ids := []string{}
	for _, ancestor := range content.Ancestors {
		ancestor_metadata, ok := metadata_cache[ancestor.ID]
		if ok {
			ancestor_names = append(
				ancestor_names,
				// might decide to ditch the quotes
				fmt.Sprintf("\"%s\"", ancestor_metadata.Title),
			)
			ancestor_ids = append(ancestor_ids, ancestor.ID)
		} else {
			// oh no, found an ID with no title mapped!!
			return LocalMarkdown{}, fmt.Errorf("data: Found an ID reference we haven't seen before! %s", ancestor.ID)
		}
	}

	ancestor_ids_str := fmt.Sprintf("[%s]", strings.Join(ancestor_ids, ", "))

	body := fmt.Sprintf(`title: %s
date: %s
version: %d
object_id: %s
uri: %s
status: %s
type: %s
ancestor_names: %s
ancestor_ids: %s
---
%s
`,
		content.Title,
		content.Version.When,
		content.Version.Number,
		content.ID,
		item_web_uri,
		content.Status,
		content.Type,
		strings.Join(ancestor_names, " > "),
		ancestor_ids_str,
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
