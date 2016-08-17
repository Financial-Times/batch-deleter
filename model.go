package main

type instructions struct {
	Hosts []string `json:"hosts"`
	Path  string   `json:"path"`
	Uuids []string `json:"uuids"`
}
