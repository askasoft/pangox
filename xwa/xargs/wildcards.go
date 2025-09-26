package xargs

import (
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/str/wildcard"
)

type Wildcards []string

func (ws Wildcards) String() string {
	return str.Join(ws, "\n")
}

func (ws Wildcards) Match(v string) bool {
	if len(ws) == 0 {
		return false
	}

	f := func(s string) bool {
		return wildcard.Match(s, v)
	}
	return asg.ContainsFunc(ws, f)
}

func (ws Wildcards) MatchAny(vs ...string) bool {
	if len(ws) == 0 {
		return false
	}

	f := func(s string) bool {
		return asg.ContainsFunc(vs, func(v string) bool {
			return wildcard.Match(s, v)
		})
	}
	return asg.ContainsFunc(ws, f)
}

func ParseWildcards(val string) Wildcards {
	return str.RemoveEmpties(str.Strips(str.FieldsAny(val, "\r\n")))
}
