package confluence

// SpacesQuery defines the query parameters for:
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-space/#api-spaces-get
type SpacesQuery struct {
	// Filter the results to spaces based on...
	IDs    []int    `url:"ids,omitempty,comma"`    // their IDs.
	Keys   []string `url:"keys,omitempty,comma"`   // their keys.
	Type   string   `url:"type,omitempty"`         // their types. Valid values: "global" or "personal"
	Status string   `url:"status,omitempty"`       // their status: current, archived.
	Labels []string `url:"labels,omitempty,comma"` // their labels.

	Sort string `url:"sort,omitempty"` // Sort order: id, -id, key, -key, name, -name

	// 'Cursor' is used for pagination; this opaque cursor will be returned in the 'next' URL in the
	// 'Link' response header.  Use the relative URL in the 'Link' header to retrieve the next set
	// of results.
	Cursor string `url:"cursor,omitempty"`
	Limit  int    `url:"limit,omitempty"` // page limit; default 25, range 1-250
}

// GetPagesQuery defines the query parameters for:
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-page/#api-pages-get
//
// Incidentally, it's the same shape as the Get Blogs query:
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-blog-post/#api-blogposts-get
type GetPagesQuery struct {
	QueryType ContentType

	// Filter the results to pages based on...
	ID         []int    `url:"id,omitempty,comma"`       // ID of the pages
	SpaceID    []int    `url:"space-id,omitempty,comma"` // Limit to particular spaces (maximum 100 per query)
	Sort       string   `url:"sort,omitempty"`           // Sort order: id, -id, created-date, -created-date, modified-date, -modified-date, title, -title
	Status     []string `url:"status,omitempty,comma"`   // their status: current, archived, deleted, trashed
	Title      string   `url:"title,omitempty"`          // Filter by title
	BodyFormat string   `url:"body-format,omitempty"`    // The content format types to be returned in the body field of the response. If available, the representation will be available under a response field of the same name under the body field.

	// 'Cursor' is used for pagination; this opaque cursor will be returned in the 'next' URL in the
	// 'Link' response header.  Use the relative URL in the 'Link' header to retrieve the next set
	// of results.
	Cursor string `url:"cursor,omitempty"`
	Limit  int    `url:"limit,omitempty"` // page limit; default 25, range 1-250
}

// GetPageByIDQuery defines the query parameters for:
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-page/#api-pages-id-get
type GetPageByIDQuery struct {
	ID int `url:"-"` // ID of the page; required

	// Filter the results to pages based on...
	BodyFormat string `url:"body-format,omitempty"` // The content format types to be returned in the body field of the response. If available, the representation will be available under a response field of the same name under the body field. Valid values: storage, atlas_doc_format, view, export_view, anonymous_export_view
	GetDraft   bool   `url:"get-draft,omitempty"`
	Version    int    `url:"version,omitempty"` // Allows you to retrieve a previously published version. Specify the previous version's number to retrieve its details.
}

// GetUserByIDQuery defines the query parameters for v1 query:
// https://developer.atlassian.com/cloud/confluence/rest/v1/api-group-users/#api-wiki-rest-api-user-get
type GetUserByIDQuery struct {
	ID string `url:"accountId"` // ID of the user; required
}

// GetFolderByIDQuery defines the query parameters for:
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-folder/#api-folders-id-get
type GetFolderByIDQuery struct {
	ID                    int  `url:"-"` // ID of the folder; required
	IncludeCollaborators  bool `url:"include-collaborators,omitempty"`
	IncludeDirectChildren bool `url:"include-direct-children,omitempty"`
	IncludeOperations     bool `url:"include-operations,omitempty"`
	IncludeProperties     bool `url:"include-properties,omitempty"`
}
