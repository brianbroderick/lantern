package extractor

import "github.com/brianbroderick/lantern/pkg/sql/object"

func setTableAliases(env *object.Environment, aliases map[string]string) {
	env.Set("table_aliases", &object.StringHash{Value: aliases})
}
