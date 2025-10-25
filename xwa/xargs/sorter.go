package xargs

import (
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/str"
)

type Sorter struct {
	Col string `json:"c,omitempty" form:"c"`
	Dir string `json:"d,omitempty" form:"d,lower"`
}

func (s *Sorter) String() string {
	return s.Col + " " + s.Dir
}

func (s *Sorter) IsAsc() bool {
	return s.Dir == "asc"
}

func (s *Sorter) IsDesc() bool {
	return s.Dir == "desc"
}

func (s *Sorter) Normalize(columns ...string) {
	if len(columns) > 0 {
		if !asg.Contains(columns, s.Col) {
			s.Col = columns[0]
		}
	}

	s.Dir = str.ToLower(s.Dir)
	if s.Dir != "asc" && s.Dir != "desc" {
		s.Dir = "asc"
	}
}
