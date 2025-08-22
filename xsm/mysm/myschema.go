package mysm

import (
	"fmt"

	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/str"
)

var SysDBs = []string{"information_schema", "mysql", "performance_schema", "sys"}

func IsSysDB(name string) bool {
	for _, db := range SysDBs {
		if str.EqualFold(db, name) {
			return true
		}
	}
	return false
}

func SQLCreateSchema(name string) string {
	return "CREATE SCHEMA " + name
}

func SQLCommentSchema(name string, comment string) string {
	return fmt.Sprintf("ALTER SCHEMA %s COMMENT = '%s'", name, sqx.EscapeString(comment))
}

func SQLDeleteSchema(name string) string {
	return fmt.Sprintf("DROP SCHEMA " + name)
}
