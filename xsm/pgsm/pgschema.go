package pgsm

import (
	"fmt"

	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/str"
)

var SysSMs = []string{"information_schema", "pg_catalog", "pg_toast"}

func IsSysSM(name string) bool {
	for _, db := range SysSMs {
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
	return fmt.Sprintf("COMMENT ON SCHEMA %s IS '%s'", name, sqx.EscapeString(comment))
}

func SQLRenameSchema(old string, new string) string {
	return fmt.Sprintf("ALTER SCHEMA %s RENAME TO %s", old, new)
}

func SQLDeleteSchema(name string) string {
	return fmt.Sprintf("DROP SCHEMA %s CASCADE", name)
}
