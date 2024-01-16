package confluence

// AllSpaces response type
type AllSpaces struct {
	Results []Space `json:"results"`

	Links struct {
		// Contains the relative URL for the next set of results, using a cursor query
		// parameter. This property will not be present if there is no additional data available.
		Next string `json:"next"`
	} `json:"_links"`
}

type MultiPageResponse struct {
	Results []Page `json:"results"`

	Links struct {
		// Contains the relative URL for the next set of results, using a cursor query
		// parameter. This property will not be present if there is no additional data available.
		Next string `json:"next"`
	} `json:"_links"`
}
