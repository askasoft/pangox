package pgsqb

import (
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/pqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
)

func StringArrayContainsAll(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" @> ?", pqx.StringArray(vals))
	}
}

func IntArrayContainsAll(sqb *sqlx.Builder, col string, vals ...int) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" @> ?", pqx.IntArray(vals))
	}
}

func Int64ArrayContainsAll(sqb *sqlx.Builder, col string, vals ...int64) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" @> ?", pqx.Int64Array(vals))
	}
}

func StringArrayNotContainsAll(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where("NOT "+sqb.Quote(col)+" @> ?", pqx.StringArray(vals))
	}
}

func IntArrayNotContainsAll(sqb *sqlx.Builder, col string, vals ...int) {
	if len(vals) > 0 {
		sqb.Where("NOT "+sqb.Quote(col)+" @> ?", pqx.IntArray(vals))
	}
}

func Int64ArrayNotContainsAll(sqb *sqlx.Builder, col string, vals ...int64) {
	if len(vals) > 0 {
		sqb.Where("NOT "+sqb.Quote(col)+" @> ?", pqx.Int64Array(vals))
	}
}

func StringArrayContainsAny(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" && ?", pqx.StringArray(vals))
	}
}

func IntArrayContainsAny(sqb *sqlx.Builder, col string, vals ...int) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" && ?", pqx.IntArray(vals))
	}
}

func Int64ArrayContainsAny(sqb *sqlx.Builder, col string, vals ...int64) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" && ?", pqx.Int64Array(vals))
	}
}

func StringArrayNotContainsAny(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where("NOT "+sqb.Quote(col)+" && ?", pqx.StringArray(vals))
	}
}

func IntArrayNotContainsAny(sqb *sqlx.Builder, col string, vals ...int) {
	if len(vals) > 0 {
		sqb.Where("NOT "+sqb.Quote(col)+" && ?", pqx.IntArray(vals))
	}
}

func Int64ArrayNotContainsAny(sqb *sqlx.Builder, col string, vals ...int64) {
	if len(vals) > 0 {
		sqb.Where("NOT "+sqb.Quote(col)+" && ?", pqx.Int64Array(vals))
	}
}

func JSONStringsContainsAll(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" @> ?", sqx.JSONStringArray(vals))
	}
}

func JSONIntsContainsAll(sqb *sqlx.Builder, col string, vals ...int) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" @> ?", sqx.JSONIntArray(vals))
	}
}

func JSONInt64sContainsAll(sqb *sqlx.Builder, col string, vals ...int64) {
	if len(vals) > 0 {
		sqb.Where(sqb.Quote(col)+" @> ?", sqx.JSONInt64Array(vals))
	}
}

func JSONStringsContainsAny(sqb *sqlx.Builder, col string, vals ...string) {
	if len(vals) == 0 {
		return
	}

	var sb str.Builder
	args := make([]any, 0, len(vals))

	sb.WriteByte('(')
	for i, v := range vals {
		if i > 0 {
			sb.WriteString(" OR ")
		}
		sb.WriteString(sqb.Quote(col))
		sb.WriteString(" @> ?")
		args = append(args, sqx.JSONStringArray{v})
	}
	sb.WriteByte(')')

	sqb.Where(sb.String(), args...)
}

func JSONIntsContainsAny(sqb *sqlx.Builder, col string, vals ...int) {
	if len(vals) > 0 {
		sqb.Where("translate("+sqb.Quote(col)+", '[]', '{}')::bigint[] && ?", pqx.IntArray(vals))
	}
}

func JSONInt64sContainsAny(sqb *sqlx.Builder, col string, vals ...int64) {
	if len(vals) > 0 {
		sqb.Where("translate("+sqb.Quote(col)+", '[]', '{}')::bigint[] && ?", pqx.Int64Array(vals))
	}
}

func JSONFlagsContainsAll(sqb *sqlx.Builder, col string, props ...string) {
	jsonFlagsContains(sqb, col, props, true)
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
		sb.WriteString("(")
		sb.WriteString(sqb.Quote(col))
		sb.WriteString("->'")
		sb.WriteString(p)
		sb.WriteString("')::integer = 1")
	}
	sb.WriteByte(')')

	sqb.Where(sb.String())
}
