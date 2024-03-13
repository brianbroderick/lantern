package repo

import "github.com/google/uuid"

type Source struct {
	UID  uuid.UUID `json:"uid,omitempty"`  // unique UUID of the source
	Name string    `json:"name,omitempty"` // the name of the source
	URL  string    `json:"url,omitempty"`  // the url of the source
}

func NewSource(name, url string) *Source {
	return &Source{
		UID:  UuidV5(name),
		Name: name,
		URL:  url,
	}
}
