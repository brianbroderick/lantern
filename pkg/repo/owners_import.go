package repo

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type OwnersImports struct {
	Data     []OwnersImport `json:"data,omitempty"`
	Database string         `json:"database,omitempty"`
}

type OwnersImport struct {
	Table  string   `json:"table,omitempty"`
	Owners []string `json:"owners,omitempty"`
}

func NewOwnersImports() *OwnersImports {
	return &OwnersImports{}
}

func (o *OwnersImports) Import(data []byte) error {
	if err := o.Unmarshal(data); err != nil {
		return err
	}
	fmt.Printf("Records to import: %d\n", len(o.Data))

	databaseUID := SetDatabaseUID(o.Database)
	now := time.Now()

	db := Conn()
	defer db.Close()

	tables := NewTables()
	tables.SetAll(db)

	owners := NewOwners()
	collection := NewOwnersTablesCollection()

	for _, oi := range o.Data {
		spl := strings.Split(oi.Table, ".")

		schema := "public"
		tableName := ""

		switch len(spl) {
		case 1:
			tableName = spl[0]
		case 2:
			schema = spl[0]
			tableName = spl[1]
		}

		if _, ok := tables.Tables[fmt.Sprintf("%s.%s", schema, tableName)]; !ok {
			tab := &Table{Schema: schema, Name: tableName, TableType: "owners_import", DatabaseUID: databaseUID, UpdatedAt: now, CreatedAt: now, EstimatedRowCount: -1, DataSizeBytes: -1}
			tab.SetUID()
			tables.Add(tab)
		}

		for _, owner := range oi.Owners {
			owners.Add(owner)
			ot := &OwnersTables{
				OwnerUID: owners.Owners[owner].UID,
				TableUID: tables.Tables[fmt.Sprintf("%s.%s", schema, tableName)].UID,
			}
			ot.SetUID()
			collection.Add(ot)
		}
	}

	tables.Upsert(db)
	owners.Upsert(db)
	collection.Upsert(db)
	return nil
}

func (o *OwnersImports) Unmarshal(data []byte) error {
	return json.Unmarshal(data, o)
}
