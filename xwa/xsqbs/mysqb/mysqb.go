package mysqb

import (
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

func ResetSequenceSQL(table string, starts ...int64) string {
	return ""
}

func JSONStringsContainsAll(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where("json_contains("+sqb.Quote(col)+", ?)", sqx.JSONStringArray(vals))
	}
}

func JSONIntsContainsAll(sqb *sqlx.Builder, col string, vals ...int) {
	if len(vals) > 0 {
		sqb.Where("json_contains("+sqb.Quote(col)+", ?)", sqx.JSONIntArray(vals))
	}
}

func JSONInt64sContainsAll(sqb *sqlx.Builder, col string, vals ...int64) {
	if len(vals) > 0 {
		sqb.Where("json_contains("+sqb.Quote(col)+", ?)", sqx.JSONInt64Array(vals))
	}
}

func JSONStringsContainsAny(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where("json_overlaps("+sqb.Quote(col)+", ?)", sqx.JSONStringArray(vals))
	}
}

func JSONIntsContainsAny(sqb *sqlx.Builder, col string, vals ...int) {
	if len(vals) > 0 {
		sqb.Where("json_overlaps("+sqb.Quote(col)+", ?)", sqx.JSONIntArray(vals))
	}
}

func JSONInt64sContainsAny(sqb *sqlx.Builder, col string, vals ...int64) {
	if len(vals) > 0 {
		sqb.Where("json_overlaps("+sqb.Quote(col)+", ?)", sqx.JSONInt64Array(vals))
	}
}

func JSONFlagsContainsAll(sqb *sqlx.Builder, col string, props ...string) {
	jsonFlagsContains(sqb, col, props, false)
}

func JSONFlagsContainsAny(sqb *sqlx.Builder, col string, props ...string) {
	jsonFlagsContains(sqb, col, props, false)
}

func jsonFlagsContains(sqb *sqlx.Builder, col string, props []string, all bool) {
	if len(props) == 0 {
		return
	}

	var sb str.Builder

	sb.WriteByte('(')
	for i, p := range props {
		if i > 0 {
			sb.WriteString(str.If(all, " AND ", " OR "))
		}
		sb.WriteString("json_extract(")
		sb.WriteString(sqb.Quote(col))
		sb.WriteString(", '$.")
		sb.WriteString(p)
		sb.WriteString("') = 1")
	}
	sb.WriteByte(')')

	sqb.Where(sb.String())
}
