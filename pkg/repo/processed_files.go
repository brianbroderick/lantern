package repo

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ProcessedFile struct {
	UID         uuid.UUID `json:"uid,omitempty"`
	FileName    string    `json:"file_name,omitempty"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
}

func NewProcessedFile(fileName string) *ProcessedFile {
	sanitized := strings.ReplaceAll(fileName, "'", "''")

	return &ProcessedFile{
		UID:         UuidV5(fileName), // UUIDv5 sha of the file name, don't need the sanitized version
		FileName:    sanitized,
		ProcessedAt: time.Now()}
}

func (p *ProcessedFile) HasBeenProcessed(db *sql.DB) bool {
	var count int64
	row := db.QueryRow(
		fmt.Sprintf(`SELECT COUNT(1) FROM processed_files WHERE uid = '%s'`,
			p.UID))
	row.Scan(&count)

	return count > 0
}

func (p *ProcessedFile) Processed(db *sql.DB) {
	if p.FileName == "" {
		fmt.Println("ProcessedFile.FileName is required")
		return
	}

	query := fmt.Sprintf(`INSERT INTO processed_files (uid, file_name, processed_at) 
	VALUES ('%s', '%s', '%s') 
	ON CONFLICT (file_name) DO UPDATE SET processed_at = EXCLUDED.processed_at`,
		p.UID, p.FileName, p.ProcessedAt.Format("2006-01-02 15:04:05"))

	ExecuteQuery(db, query)
}
