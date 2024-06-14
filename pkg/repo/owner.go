package repo

import "github.com/google/uuid"

type Owner struct {
	UID  uuid.UUID `json:"uid,omitempty"`  // unique UUID of the owner
	Name string    `json:"name,omitempty"` // the name of the owner
}

func (o *Owner) SetUID() {
	o.UID = UuidV5(o.Name)
}
