package repo

import "github.com/google/uuid"

type OwnersTables struct {
	UID      uuid.UUID `json:"uid,omitempty"`       // unique UUID of the row
	OwnerUID uuid.UUID `json:"owner_uid,omitempty"` // the UUID of the owner
	TableUID uuid.UUID `json:"table_uid,omitempty"` // the UUID of the table
}

func (o *OwnersTables) SetUID() {
	o.UID = UuidV5(o.OwnerUID.String() + o.TableUID.String())
}
