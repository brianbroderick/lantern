package repo

import (
	"fmt"

	"github.com/google/uuid"
)

type CreateStatement struct {
	UID              uuid.UUID `json:"uid,omitempty"`
	Scope            string    `json:"scope,omitempty"`
	IsUnique         bool      `json:"is_unique,omitempty"`
	UsedConcurrently bool      `json:"used_concurrently,omitempty"`
	IsTemp           bool      `json:"is_temp,omitempty"`
	IsUnlogged       bool      `json:"is_unlogged,omitempty"`
	ObjectType       string    `json:"object_type,omitempty"`
	IfNotExists      bool      `json:"if_not_exists,omitempty"`
	Name             string    `json:"name,omitempty"`
	OnCommit         string    `json:"on_commit,omitempty"`
	Operator         string    `json:"operator,omitempty"`
	Expression       string    `json:"expression,omitempty"`
	WhereClause      string    `json:"where_clause,omitempty"`
}

func (c *CreateStatement) SetUID() {
	c.UID = UuidV5(fmt.Sprintf("%s|%t|%t|%t|%s|%s|%s|%s",
		c.Scope, c.IsUnique, c.IsTemp, c.IsUnlogged, c.ObjectType, c.Name, c.Expression, c.WhereClause))
}
