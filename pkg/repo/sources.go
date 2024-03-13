package repo

import (
	"fmt"
	"strings"
)

type Sources struct {
	Sources map[string]*Source `json:"sources,omitempty"` // the key is the sha of the database
}

func NewSources() *Sources {
	return &Sources{
		Sources: make(map[string]*Source),
	}
}

func (d *Sources) Add(name, url string) *Source {
	if _, ok := d.Sources[name]; !ok {
		d.Sources[name] = NewSource(name, url)
	}

	return d.Sources[name]
}

func (d *Sources) Upsert() {
	if len(d.Sources) == 0 {
		return
	}

	rows := d.insValues()
	query := fmt.Sprintf(d.ins(), strings.Join(rows, ",\n"))

	db := Conn()
	defer db.Close()
	ExecuteQuery(db, query)
}

func (d *Sources) ins() string {
	return `INSERT INTO sources (uid, name, url) 
	VALUES %s 
	ON CONFLICT (uid) DO NOTHING;`
}

func (d *Sources) insValues() []string {
	var rows []string

	for name, source := range d.Sources {
		rows = append(rows,
			fmt.Sprintf("('%s', '%s', '%s')",
				source.UID, name, source.URL))
	}
	return rows
}
