package deprecated

import (
	"regexp"
	"strings"

	pg_query "github.com/lfittl/pg_query_go"
)

// normalizeQuery converts "select * from users where id = 1" to "select * from users where id = ?"
func normalizeQuery(sql string) ([]byte, error) {
	mutex.Lock()
	tree, err := pg_query.Normalize(strings.ToLower(sql))
	mutex.Unlock()
	if err != nil {
		return nil, err
	}

	tree = truncateInLists(tree)
	return []byte(tree), err
}

func truncateInLists(str string) string {
	r := regexp.MustCompile(`(\?,\s*)+`)
	return r.ReplaceAllString(str, "")
}
