package xargs

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/str"
)

type Keywords []string

func (ks Keywords) String() string {
	if len(ks) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, k := range ks {
		if str.ContainsFunc(k, unicode.IsSpace) {
			k = strconv.Quote(k)
		}
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(k)
	}
	return sb.String()
}

func (ks Keywords) Contains(v string) bool {
	if len(ks) == 0 {
		return false
	}

	f := func(s string) bool {
		return str.ContainsFold(v, s)
	}
	return asg.ContainsFunc(ks, f)
}

func (ks Keywords) ContainsAny(vs ...string) bool {
	if len(ks) == 0 {
		return false
	}

	f := func(s string) bool {
		return asg.ContainsFunc(vs, func(v string) bool {
			return str.ContainsFold(v, s)
		})
	}
	return asg.ContainsFunc(ks, f)
}

func ParseKeywords(val string) (keys Keywords) {
	var key string

	for val != "" {
		key, val, _ = NextKeyword(val)

		if key == "" {
			continue
		}

		keys = append(keys, key)
	}

	return
}

func NextKeyword(s string) (string, string, bool) {
	s = str.Strip(s)

	if s == "" {
		return "", "", false
	}

	if s[0] == '"' {
		i := str.IndexByte(s[1:], '"')
		if i >= 0 {
			return s[1 : i+1], s[i+2:], true
		}
	}

	i := str.IndexFunc(s, unicode.IsSpace)
	if i >= 0 {
		return s[:i], s[i:], false
	}

	return s, "", false
}
