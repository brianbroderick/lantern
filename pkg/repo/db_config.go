package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Default config values. These can be overwritten by the params passed in.
var (
	Env        = "dev"
	DBHost     = "localhost"
	DBPort     = "5432"
	DBUser     = "postgres"
	DBPassword = ""
	DBName     = "lantern"
)

func DisplayConfig() {
	fmt.Printf("\nEnvironment: %s\n", Env)
	fmt.Printf("DB Host: %s\n", DBHost)
	fmt.Printf("DB Port: %s\n", DBPort)
	fmt.Printf("DB User: %s\n", DBUser)
	fmt.Printf("DB Name: %s\n", DBName)
}

func OverrideConfig(strArgs map[string]*string, intArgs map[string]*int, boolArgs map[string]*bool) {
	if strArgs["env"] != nil {
		Env = *strArgs["env"]
	}
	if strArgs["dbHost"] != nil {
		DBHost = *strArgs["dbHost"]
	}
	if strArgs["dbPort"] != nil {
		DBPort = *strArgs["dbPort"]
	}
	if strArgs["dbUser"] != nil {
		DBUser = *strArgs["dbUser"]
	}
	if strArgs["dbPassword"] != nil {
		DBPassword = *strArgs["dbPassword"]
	}
	if strArgs["dbName"] != nil {
		DBName = *strArgs["dbName"]
	}
}

// Consider turning off autovaccum: ALTER TABLE table_name SET (autovacuum_enabled = false);
// At the end, make sure they're all turned back on:
// SELECT reloptions FROM pg_class WHERE relname = 'products'; -- Returns: {autovacuum_enabled=false}

func ExecuteQuery(db *sql.DB, sql string) {
	if sql == "" {
		return
	}

	// sql = fmt.Sprintf("BEGIN; SET session_replication_role = replica; %s COMMIT;", sql)
	ctx, _ := getCtx()
	_, err := db.ExecContext(ctx, sql)
	if err != nil {
		fmt.Printf("\nErr in executeQuery:\n\t%s\nSQL:\n\t%s\n\n", err.Error(), sql)
		return
	}
}

// Conn returns a connection to the database.
func Conn() *sql.DB {
	// Build connection string. Separate out each part because special characters in the password can cause issues.
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
		DBHost, DBPort, DBUser, DBName, DBPassword)

	db, err := sql.Open("postgres", psqlInfo)
	if HasErr("sql.Open", err) {
		return nil
	}

	return db
}

func Ping(db *sql.DB) {
	err := db.Ping()
	if HasErr("db.Ping", err) {
		return
	}
}

func getCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 86400*time.Second)
}

func HasErr(msg string, err error) bool {
	if err != nil {
		fmt.Printf("Message: %s\nHasErr: %s\n\n", msg, err.Error())
		return true
	}
	return false
}
