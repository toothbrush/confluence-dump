package data

type LocalMarkdown struct {
	// contents of the file
	Content string

	// original Confluence ID of the item
	ID string

	// path relative to DUMP location (e.g., ~/confluence)
	RelativePath string
}
