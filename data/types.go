package data

import (
	conf "github.com/virtomize/confluence-go-api"
)

// I want a little more detail than the built-in one provides.
type ConfluenceSpace struct {
	Space conf.Space
	Org   string
}

type ConfluenceContent struct {
	Content conf.Content
	Space   ConfluenceSpace
}
