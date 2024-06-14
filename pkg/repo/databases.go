package repo

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Databases struct {
	Databases map[string]*Database `json:"databases,omitempty"` // the key is the UUIDv5 sha of the database
}

type Database struct {
	UID      uuid.UUID `json:"uid,omitempty"`      // unique UUID of the database
	Name     string    `json:"name,omitempty"`     // the name of the database
	Template string    `json:"template,omitempty"` // the template of the database
}

func (d *Databases) AddDatabase(database, template string) *Database {
	if _, ok := d.Databases[database]; !ok {
		d.Databases[database] = &Database{
			UID:      SetDatabaseUID(database),
			Name:     database,
			Template: template,
		}

	}

	return d.Databases[database]
}

func NewDatabases() *Databases {
	return &Databases{
		Databases: make(map[string]*Database),
	}
}

func (d *Databases) CountInDB() int {
	db := Conn()
	defer db.Close()

	var count int
	row := db.QueryRow("SELECT COUNT(1) FROM databases")
	row.Scan(&count)

	return count
}

func (d *Databases) Upsert(db *sql.DB) {
	if len(d.Databases) == 0 {
		return
	}

	rows := d.insValues()
	query := fmt.Sprintf(d.ins(), strings.Join(rows, ",\n"))

	ExecuteQuery(db, query)
}

func (d *Databases) ins() string {
	return `INSERT INTO databases (uid, name, template) 
	VALUES %s 
	ON CONFLICT (uid) DO NOTHING;`
}

func (d *Databases) insValues() []string {
	var rows []string

	for _, record := range d.Databases {
		rows = append(rows,
			fmt.Sprintf("('%s', '%s', '%s')",
				record.UID, record.Name, record.Template))
	}
	return rows
}

func SetDatabaseUID(name string) uuid.UUID {
	return UuidV5(name)
}
