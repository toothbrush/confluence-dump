package confluence

import "encoding/json"

// See https://developer.atlassian.com/cloud/confluence/rest/v1/api-group-users/#api-wiki-rest-api-user-get
type User struct {
	Type        string `json:"type"`
	Username    string `json:"username"`
	UserKey     string `json:"userKey"`
	AccountID   string `json:"accountId"`
	AccountType string `json:"accountType"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

// See https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-space/#api-spaces-get. I'm
// embellishing that with the Org/"Confluence instance name" field for convenience.
type Space struct {
	ID     string `json:"id,omitempty"`
	Key    string `json:"key,omitempty"`
	Name   string `json:"name,omitempty"`
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
	Org    string
}

// See https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-page/#api-pages-get. I'm
// embellishing that with the Org/"Confluence instance name" field for convenience.
//
// I'm cheating a bit because i plan to use this for blogposts, too, even though that's a different
// endpoint&type on Confluence's end:
// https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-blog-post/#api-blogposts-id-get.
//
// I think it'll be fine.  Just means Parent.* will be empty.
type Page struct {
	ID          string `json:"id,omitempty"`
	Status      string `json:"status,omitempty"` // current, archived, deleted, trashed
	Title       string `json:"title,omitempty"`
	SpaceID     string `json:"spaceId,omitempty"`
	ParentID    string `json:"parentId,omitempty"`
	ParentType  string `json:"parentType,omitempty"`
	Position    int    `json:"position,omitempty"`
	AuthorID    string `json:"authorId,omitempty"`
	OwnerID     string `json:"ownerId,omitempty"`
	LastOwnerID string `json:"lastOwnerId,omitempty"`

	CreatedAt string   `json:"createdAt"`
	Version   *Version `json:"version,omitempty"`

	Body Body `json:"body"`

	Links struct {
		WebUI  string `json:"webui"`
		EditUI string `json:"editui"`
		TinyUI string `json:"tinyui"`
	} `json:"_links"`

	SpaceKey string
	Org      string

	ContentType ContentType
}

// Folder represents a Confluence folder (just organises other pages)
// Very similar to a page, but non-existent fields have been commented out
type Folder struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"` // current, archived, deleted, trashed
	Title  string `json:"title,omitempty"`
	//SpaceID     string `json:"spaceId,omitempty"`
	ParentID   string `json:"parentId,omitempty"`
	ParentType string `json:"parentType,omitempty"`
	Position   int    `json:"position,omitempty"`
	AuthorID   string `json:"authorId,omitempty"`
	OwnerID    string `json:"ownerId,omitempty"`
	//LastOwnerID string `json:"lastOwnerId,omitempty"`

	// The API docs claim this is a string in "YYYY-MM-DDTHH:mm:ss.sssZ" format, but this is a lie
	CreatedAt json.Number `json:"createdAt"`
	Version   *Version    `json:"version,omitempty"`

	//Body Body `json:"body"`

	Links struct {
		WebUI string `json:"webui"`
		//EditUI string `json:"editui"`
		//TinyUI string `json:"tinyui"`
	} `json:"_links"`

	SpaceKey string
	Org      string

	ContentType ContentType
}

// Version defines the content version number
// the version number is used for updating content
type Version struct {
	CreatedAt string `json:"createdAt"`
	Message   string `json:"message,omitempty"`
	Number    int    `json:"number"`
	MinorEdit bool   `json:"minorEdit"`
	AuthorID  string `json:"authorId,omitempty"`
}

// Body holds the storage information
type Body struct {
	Storage        Storage  `json:"storage"`
	AtlasDocFormat *Storage `json:"atlas_doc_format,omitempty"`
	View           *Storage `json:"view,omitempty"`
}

// Storage defines the storage information
type Storage struct {
	Representation string `json:"representation"`
	Value          string `json:"value"`
}

type ContentType int

const (
	PageContent ContentType = iota
	BlogContent
	FolderContent
)

func (c ContentType) String() string {
	switch c {
	case BlogContent:
		return "blogpost"
	case FolderContent:
		return "folder"
	default:
		return "page"
	}
}
