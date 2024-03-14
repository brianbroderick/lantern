package extractor

import (
	"github.com/brianbroderick/lantern/pkg/sql/object"
	"github.com/google/uuid"
)

func setTableAliases(env *object.Environment, aliases map[string]string) {
	env.Set("table_aliases", &object.StringHash{Value: aliases})
}

func setDatabaseUID(env *object.Environment, uid uuid.UUID) {
	env.Set("database_uid", &object.UID{Value: uid})
}
