package xargs

import (
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
)

type IDArg struct {
	ID string `json:"id,omitempty" form:"id,strip"`

	ids []int64
	all bool
}

func (ida *IDArg) String() string {
	if ida.all {
		return "*"
	}
	return asg.Join(ida.ids, ",")
}

func (ida *IDArg) IDs() []int64 {
	return ida.ids
}

func (ida *IDArg) hasValidID() bool {
	return len(ida.ids) > 0 || ida.all
}

func (ida *IDArg) ParseID() error {
	ida.ids, ida.all = splitIDs(ida.ID)
	if !ida.hasValidID() {
		return ErrInvalidID
	}
	return nil
}

func (ida *IDArg) Bind(c *xin.Context) error {
	if err := c.Bind(ida); err != nil {
		return err
	}
	return ida.ParseID()
}

func splitIDs(id string) ([]int64, bool) {
	if id == "" {
		return nil, false
	}
	if id == "*" {
		return nil, true
	}

	ss := str.FieldsByte(id, ',')
	ids := make([]int64, 0, len(ss))
	for _, s := range ss {
		id := num.Atol(s)
		if id != 0 {
			ids = append(ids, id)
		}
	}
	return ids, false
}

type PKArg struct {
	ID string `json:"id,omitempty" form:"id,strip"`

	pks []string
	all bool
}

func (pka *PKArg) String() string {
	if pka.all {
		return "[*]"
	}
	return "[" + asg.Join(pka.pks, ",") + "]"
}

func (pka *PKArg) PKs() []string {
	return pka.pks
}

func (pka *PKArg) hasValidID() bool {
	return len(pka.pks) > 0 || pka.all
}

func (pka *PKArg) ParseID() error {
	pka.pks, pka.all = splitPKs(pka.ID)
	if !pka.hasValidID() {
		return ErrInvalidID
	}
	return nil
}

func (pka *PKArg) Bind(c *xin.Context) error {
	if err := c.Bind(pka); err != nil {
		return err
	}
	return pka.ParseID()
}

func splitPKs(id string) ([]string, bool) {
	if id == "" {
		return nil, false
	}
	if id == "*" {
		return nil, true
	}

	pks := str.RemoveEmpties(str.Strips(str.FieldsByte(id, ',')))
	return pks, false
}
