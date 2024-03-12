package repo

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Databases struct {
	Databases map[string]uuid.UUID `json:"databases,omitempty"` // the key is the sha of the database
}

func (d *Databases) addDatabase(database string) uuid.UUID {
	if _, ok := d.Databases[database]; !ok {
		d.Databases[database] = UuidV5(database)
	}

	return d.Databases[database]
}

func NewDatabases() *Databases {
	return &Databases{
		Databases: make(map[string]uuid.UUID),
	}
}

func (d *Databases) Upsert() {
	if len(d.Databases) == 0 {
		return
	}

	rows := d.insValues()
	query := fmt.Sprintf(d.ins(), strings.Join(rows, ",\n"))

	db := Conn()
	defer db.Close()
	ExecuteQuery(db, query)
}

func (d *Databases) ins() string {
	return `INSERT INTO databases (uid, name) 
	VALUES %s 
	ON CONFLICT (uid) DO NOTHING;`
}

func (d *Databases) insValues() []string {
	var rows []string

	for name, uid := range d.Databases {
		rows = append(rows,
			fmt.Sprintf("('%s', '%s')",
				uid, name))
	}
	return rows
}
