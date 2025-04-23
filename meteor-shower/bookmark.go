package main

type Bookmark struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Parent string `json:"parent"`
}
