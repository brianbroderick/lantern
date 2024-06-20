package repo

import (
	"database/sql"
	"fmt"
	"strings"
)

type OwnersTablesCollection struct {
	OwnersTables map[string]*OwnersTables `json:"owners_tables,omitempty"`
}

// NewOwnersTablesCollection creates a new OwnersTablesCollection struct
func NewOwnersTablesCollection() *OwnersTablesCollection {
	return &OwnersTablesCollection{
		OwnersTables: make(map[string]*OwnersTables),
	}
}

// Add adds a new OwnersTables to the collection
func (o *OwnersTablesCollection) Add(own *OwnersTables) *OwnersTables {
	if _, ok := o.OwnersTables[own.UID.String()]; !ok {
		o.OwnersTables[own.UID.String()] = own
	}

	return o.OwnersTables[own.UID.String()]
}

// CountInDB returns the number of rows in the owners_tables table
func (o *OwnersTablesCollection) CountInDB(db *sql.DB) int64 {
	var count int64
	row := db.QueryRow("SELECT COUNT(1) FROM owners_tables")
	row.Scan(&count)

	return count
}

// Upsert inserts or updates the rows in the owners_tables table
func (o *OwnersTablesCollection) Upsert(db *sql.DB) {
	if len(o.OwnersTables) == 0 {
		return
	}

	rows := o.insValues()
	query := fmt.Sprintf(o.ins(), strings.Join(rows, ",\n"))

	ExecuteQuery(db, query)
}

func (o *OwnersTablesCollection) ins() string {
	return `INSERT INTO owners_tables (uid, owner_uid, table_uid)
	VALUES %s
	ON CONFLICT (uid) DO UPDATE 
	SET uid = EXCLUDED.uid, 
	    owner_uid = EXCLUDED.owner_uid, 
		table_uid = EXCLUDED.table_uid;`
}

func (o *OwnersTablesCollection) insValues() []string {
	var rows []string

	for _, ownerTable := range o.OwnersTables {
		rows = append(rows, o.insValue(ownerTable))
	}
	return rows
}

func (o *OwnersTablesCollection) insValue(ownerTable *OwnersTables) string {
	return fmt.Sprintf("('%s', '%s', '%s')", ownerTable.UID, ownerTable.OwnerUID, ownerTable.TableUID)
}
