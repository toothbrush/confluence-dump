package data

type LocalMarkdown struct {
	// contents of the file
	Content string

	// original Confluence ID of the item
	ID ContentID

	Version int

	// path relative to DUMP location (e.g., ~/confluence)
	RelativePath RelativePath
}

type LocalMarkdownCache map[ContentID]LocalMarkdown

type RemoteObjectMetadata struct {
	ID       ContentID
	Title    string
	Slug     string
	SpaceKey string
	Version  int
	Org      string
}

// i might want to rename this, because it's meh, but this guy is what we build up based on the info
// retrieved from Confluence.  we don't want to repeat requests, so once we've 'seen' an ID, we keep
// some information about it that other pages might need, like titles, for ancestry info.
type RemoteContentCache map[ContentID]RemoteObjectMetadata

type ContentID string
type RelativePath string
