package xsqbs

import (
	"time"

	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/sqx"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tmu"
	"github.com/askasoft/pango/xin/taglib/args"
	"github.com/askasoft/pangox/xwa/xargs"
)

func AddPager(sqb *sqlx.Builder, p *args.Pager) {
	sqb.Offset(p.Start()).Limit(p.Limit)
}

func AddOrder(sqb *sqlx.Builder, st *args.Sorter, defcol string) {
	cols := str.FieldsByte(st.Col, ',')

	defs := false
	for _, col := range cols {
		b, a, ok := str.CutByte(col, '.')
		if ok {
			col = b + "->>'" + a + "'"
		}

		sqb.Order(col, st.IsDesc())
		if col == defcol {
			defs = true
		}
	}

	if !defs {
		sqb.Order(defcol, st.IsDesc())
	}
}

func AddEq(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.Eq(col, val)
	}
}

func AddNeq(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.Neq(col, val)
	}
}

func AddIn[T any](sqb *sqlx.Builder, col string, vals []T) {
	if len(vals) > 0 {
		sqb.In(col, vals)
	}
}

func AddStartsLike(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.Like(col, sqx.StartsLike(val))
	}
}

func AddEndsLike(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.Like(col, sqx.EndsLike(val))
	}
}

func AddLike(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.Like(col, sqx.StringLike(val))
	}
}

func AddStartsILike(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.ILike(col, sqx.StartsLike(val))
	}
}

func AddEndsILike(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.ILike(col, sqx.EndsLike(val))
	}
}

func AddILike(sqb *sqlx.Builder, col string, val string) {
	if val != "" {
		sqb.ILike(col, sqx.StringLike(val))
	}
}

func AddDateRange(sqb *sqlx.Builder, col string, tmin, tmax time.Time) {
	if !tmin.IsZero() {
		tmin = tmu.TruncateHours(tmin)
	}
	if !tmax.IsZero() {
		tmax = tmu.TruncateHours(tmax).Add(time.Hour * 24)
	}
	AddTimeRange(sqb, col, tmin, tmax)
}

func AddTimeRange(sqb *sqlx.Builder, col string, tmin, tmax time.Time) {
	switch {
	case tmin.IsZero() && tmax.IsZero():
	case tmin.IsZero():
		sqb.Lte(col, tmax)
	case tmax.IsZero():
		sqb.Gte(col, tmin)
	default:
		switch {
		case tmin.Before(tmax):
			sqb.Btw(col, tmin, tmax)
		case tmin.After(tmax):
			sqb.Btw(col, tmax, tmin)
		default:
			sqb.Eq(col, tmin)
		}
	}
}

func AddIntRange(sqb *sqlx.Builder, col string, smin, smax string) {
	q, a := buildIntRange(smin, smax)
	if q != "" {
		sqb.Where(sqb.Quote(col)+q, a...)
	}
}

func buildIntRange(smin, smax string) (string, []any) {
	switch {
	case smin == "" && smax == "": // invalid
		return "", nil
	case smin == "":
		return " <= ?", []any{num.Atol(smax)}
	case smax == "":
		return " >= ?", []any{num.Atol(smin)}
	default:
		imin := num.Atol(smin)
		imax := num.Atol(smax)
		switch {
		case imin < imax:
			return " BETWEEN ? AND ?", []any{imin, imax}
		case imin > imax:
			return " BETWEEN ? AND ?", []any{imax, imin}
		default:
			return " = ?", []any{imin}
		}
	}
}

func AddFloatRange(sqb *sqlx.Builder, col string, smin, smax string) {
	q, a := buildFloatRange(smin, smax)
	if q != "" {
		sqb.Where(sqb.Quote(col)+" "+q, a...)
	}
}

func buildFloatRange(smin, smax string) (string, []any) {
	switch {
	case smin == "" && smax == "": // invalid
		return "", nil
	case smin == "":
		return " <= ?", []any{num.Atof(smax)}
	case smax == "":
		return " >= ?", []any{num.Atof(smin)}
	default:
		fmin := num.Atof(smin)
		fmax := num.Atof(smax)
		switch {
		case fmin < fmax:
			return " BETWEEN ? AND ?", []any{fmin, fmax}
		case fmin > fmax:
			return " BETWEEN ? AND ?", []any{fmax, fmin}
		default:
			return " = ?", []any{fmin}
		}
	}
}

func AddIntegers(sqb *sqlx.Builder, col string, val string) {
	AddIntegersEx(sqb, col, val, false)
}

func AddIntegersEx(sqb *sqlx.Builder, col string, val string, not bool) {
	addIntegersEx(sqb, col, val, '~', not)
}

func AddUintegers(sqb *sqlx.Builder, col string, val string) {
	AddUintegersEx(sqb, col, val, false)
}

func AddUintegersEx(sqb *sqlx.Builder, col string, val string, not bool) {
	addIntegersEx(sqb, col, val, '-', not)
}

func addIntegersEx(sqb *sqlx.Builder, col string, val string, sep byte, not bool) {
	ss := str.Fields(val)

	if len(ss) == 0 {
		return
	}

	var sb str.Builder
	var args []any

	if not {
		sb.WriteString("NOT ")
	}

	sb.WriteByte('(')
	for _, s := range ss {
		if len(args) > 0 {
			sb.WriteString(" OR ")
		}

		smin, smax, ok := str.CutByte(s, sep)
		if ok {
			q, a := buildIntRange(smin, smax)
			if q != "" {
				sb.WriteString(sqb.Quote(col))
				sb.WriteString(q)
				args = append(args, a...)
			}
		} else {
			sb.WriteString(sqb.Quote(col))
			sb.WriteString(" = ?")
			args = append(args, num.Atol(s))
		}
	}
	sb.WriteByte(')')

	sqb.Where(sb.String(), args...)
}

func AddDecimals(sqb *sqlx.Builder, col string, val string) {
	AddDecimalsEx(sqb, col, val, false)
}

func AddDecimalsEx(sqb *sqlx.Builder, col string, val string, not bool) {
	addDecimalsEx(sqb, col, val, '~', not)
}

func AddUdecimals(sqb *sqlx.Builder, col string, val string) {
	AddUdecimalsEx(sqb, col, val, false)
}

func AddUdecimalsEx(sqb *sqlx.Builder, col string, val string, not bool) {
	addDecimalsEx(sqb, col, val, '-', not)
}

func addDecimalsEx(sqb *sqlx.Builder, col string, val string, sep byte, not bool) {
	ss := str.Fields(val)

	if len(ss) == 0 {
		return
	}

	var sb str.Builder
	var args []any

	if not {
		sb.WriteString("NOT ")
	}

	sb.WriteByte('(')
	for _, s := range ss {
		if len(args) > 0 {
			sb.WriteString(" OR ")
		}

		smin, smax, ok := str.CutByte(s, sep)
		if ok {
			q, a := buildFloatRange(smin, smax)
			if q != "" {
				sb.WriteString(sqb.Quote(col))
				sb.WriteString(q)
				args = append(args, a...)
			}
		} else {
			sb.WriteString(sqb.Quote(col))
			sb.WriteString(" = ?")
			args = append(args, num.Atof(s))
		}
	}
	sb.WriteByte(')')

	sqb.Where(sb.String(), args...)
}

func AddKeywordsEx(sqb *sqlx.Builder, like, col, val string, not bool) {
	val = str.Strip(val)
	if val == "" {
		return
	}

	var (
		sb   str.Builder
		args []any
		key  string
	)

	if not {
		sb.WriteString("NOT ")
	}

	sb.WriteByte('(')
	for val != "" {
		key, val, _ = xargs.NextKeyword(val)

		if key == "" {
			continue
		}

		if len(args) > 0 {
			sb.WriteString(" OR ")
		}
		sb.WriteString(sqb.Quote(col))
		sb.WriteString(" ")
		sb.WriteString(like)
		sb.WriteString(" ?")
		args = append(args, sqx.StringLike(key))
	}
	if len(args) > 0 {
		sb.WriteByte(')')
		sqb.Where(sb.String(), args...)
	}
}

func AddAndwordsEx(sqb *sqlx.Builder, like, col, val string, not bool) {
	val = str.Strip(val)
	if val == "" {
		return
	}

	var (
		sb     str.Builder
		args   []any
		key    string
		and    bool
		quoted bool
	)

	if not {
		sb.WriteString("NOT ")
	}

	sb.WriteByte('(')
	for val != "" {
		key, val, quoted = xargs.NextKeyword(val)

		if key == "" {
			continue
		}
		if !quoted && key == "&" {
			and = true
			continue
		}

		if len(args) > 0 {
			sb.WriteString(str.If(and, " AND ", " OR "))
		}
		sb.WriteString(sqb.Quote(col))
		sb.WriteString(" ")
		sb.WriteString(like)
		sb.WriteString(" ?")
		args = append(args, sqx.StringLike(key))

		and = false
	}
	if len(args) > 0 {
		sb.WriteByte(')')
		sqb.Where(sb.String(), args...)
	}
}
