package data

import (
	"fmt"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	md_plugin "github.com/JohannesKaufmann/html-to-markdown/plugin"
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

	header := MarkdownHeader{
		Title:         content.Title,
		Date:          content.Version.When,
		Version:       content.Version.Number,
		ObjectId:      content.ID,
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
	markdown := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.Meta,
		),
	)

	fullPath, err := homedir.Expand(path.Join(storePath, relativePath))
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Couldn't expand homedir: %w", err)
	}

	source, err := os.ReadFile(fullPath)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Couldn't read file %s: %w", fullPath, err)
	}

	context := parser.NewContext()
	var buf bytes.Buffer
	if err := markdown.Convert(source, &buf, parser.WithContext(context)); err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Couldn't parse Markdown from %s: %w", relativePath, err)
	}
	header := MarkdownHeader{}
	metaData := meta.Get(context)

	id, err := safeParseFieldToString("object_id", metaData)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Invalid 'object_id' field: %w", err)
	}
	header.object_id = id

	title, err := safeParseFieldToString("title", metaData)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Invalid 'title' field: %w", err)
	}
	header.title = title

	version, err := safeParseFieldToString("version", metaData)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Invalid 'version' field: %w", err)
	}
	v, err := strconv.Atoi(version)
	if err != nil {
		return LocalMarkdown{}, fmt.Errorf("data: Invalid 'version' field: %w", err)
	}
	header.version = v

	return LocalMarkdown{
		Content:      string(source),
		ID:           header.object_id,
		RelativePath: relativePath,
	}, nil
}

func safeParseFieldToString(fieldname string, metadata map[string]interface{}) (string, error) {
	if raw, ok := metadata[fieldname]; ok {
		if val, ok := raw.(string); ok {
			return val, nil
		} else {
			return "", fmt.Errorf("data: Field '%s' is not 'string'", fieldname)
		}
	} else {
		return "", fmt.Errorf("data: Field '%s' not found in Markdown metadata", fieldname)
	}
}

type MarkdownHeader struct {
	Title         string
	Date          string // eventually: time.Time
	Version       int
	ObjectId      string // eventually: int
	Uri           string
	Status        string
	ObjectType    string
	AncestorNames []string
	AncestorIds   []string // eventually: []int
}
