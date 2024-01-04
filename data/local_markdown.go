package data

type LocalMarkdown struct {
	// contents of the file
	Content string

	// original Confluence ID of the item
	ID string

	// path relative to DUMP location (e.g., ~/confluence)
	RelativePath string
}

// i might want to rename this, because it's meh, but this guy is what we build up based on the info
// retrieved from Confluence.  we don't want to repeat requests, so once we've 'seen' an ID, we keep
// some information about it that other pages might need, like titles, for ancestry info.
type MetadataCache map[string]LocalMetadata

type LocalMetadata struct {
	ID       string
	Title    string
	Slug     string
	SpaceKey string
}