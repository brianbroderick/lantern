package repo

import (
	"database/sql"
	"fmt"
	"strings"
)

type Owners struct {
	Owners map[string]*Owner `json:"owners,omitempty"`
}

func NewOwners() *Owners {
	return &Owners{
		Owners: make(map[string]*Owner),
	}
}

func (o *Owners) Add(name string) *Owner {
	if _, ok := o.Owners[name]; !ok {
		o.Owners[name] = &Owner{Name: name, UID: UuidV5(name)}
	}

	return o.Owners[name]
}

func (o *Owners) CountInDB(db *sql.DB) int {
	var count int
	row := db.QueryRow("SELECT COUNT(1) FROM owners")
	row.Scan(&count)

	return count
}

func (o *Owners) Upsert(db *sql.DB) {
	if len(o.Owners) == 0 {
		return
	}

	rows := o.insValues()
	query := fmt.Sprintf(o.ins(), strings.Join(rows, ",\n"))

	ExecuteQuery(db, query)
}

func (o *Owners) ins() string {
	return `INSERT INTO owners (uid, name)
	VALUES %s 
	ON CONFLICT (uid) DO UPDATE 
	SET uid = EXCLUDED.uid, 
	    name = EXCLUDED.name;`
}

func (o *Owners) insValues() []string {
	var rows []string

	for _, owner := range o.Owners {
		name := strings.ReplaceAll(owner.Name, "'", "''")

		rows = append(rows, fmt.Sprintf("('%s', '%s')", owner.UID, name))
	}

	return rows
}
