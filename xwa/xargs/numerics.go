package xargs

import (
	"errors"
	"strconv"
	"strings"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
)

type intrg [2]any

func (r intrg) Contains(n int64) bool {
	switch {
	case r[0] == nil:
		return n <= r[1].(int64)
	case r[1] == nil:
		return n >= r[0].(int64)
	default:
		return n >= r[0].(int64) && n <= r[1].(int64)
	}
}

type intrgs []intrg

func (rs intrgs) Contains(n int64) bool {
	for _, r := range rs {
		if r.Contains(n) {
			return true
		}
	}
	return false
}

type Integers struct {
	sep    byte
	ints   []int64
	ranges intrgs
}

func (ns Integers) IsEmpty() bool {
	return len(ns.ints) == 0 && len(ns.ranges) == 0
}

func (ns Integers) String() string {
	var sb strings.Builder

	for _, n := range ns.ints {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(num.Ltoa(n))
	}

	for _, r := range ns.ranges {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}

		if r[0] != nil {
			sb.WriteString(num.Ltoa(r[0].(int64)))
		}
		sb.WriteByte(ns.sep)
		if r[1] != nil {
			sb.WriteString(num.Ltoa(r[1].(int64)))
		}
	}

	return sb.String()
}

func (ns Integers) Contains(n int64) bool {
	return asg.Contains(ns.ints, n) || ns.ranges.Contains(n)
}

func (ns *Integers) Parse(val string) (err error) {
	ss := str.Fields(val)

	if len(ss) == 0 {
		return
	}

	var imin int64
	var imax int64
	for _, s := range ss {
		smin, smax, ok := str.CutByte(s, ns.sep)
		if ok {
			switch {
			case smin == "" && smax == "": // invalid
				err = errors.New("empty")
				return
			case smin == "":
				imax, err = strconv.ParseInt(smax, 0, 64)
				if err != nil {
					return
				}
				ns.ranges = append(ns.ranges, intrg{nil, imax})
			case smax == "":
				imin, err = strconv.ParseInt(smin, 0, 64)
				if err != nil {
					return
				}
				ns.ranges = append(ns.ranges, intrg{imin, nil})
			default:
				imin, err = strconv.ParseInt(smin, 0, 64)
				if err != nil {
					return
				}

				imax, err = strconv.ParseInt(smax, 0, 64)
				if err != nil {
					return
				}

				switch {
				case imin < imax:
					ns.ranges = append(ns.ranges, intrg{imin, imax})
				case imin > imax:
					ns.ranges = append(ns.ranges, intrg{imax, imin})
				default:
					ns.ints = append(ns.ints, imin)
				}
			}
		} else {
			imin, err = strconv.ParseInt(s, 0, 64)
			if err != nil {
				return
			}
			ns.ints = append(ns.ints, imin)
		}
	}
	return
}

func ParseIntegers(val string) (ns Integers, err error) {
	ns.sep = '~'
	err = ns.Parse(val)
	return
}

func ParseUintegers(val string) (ns Integers, err error) {
	ns.sep = '-'
	err = ns.Parse(val)
	return
}

//-------------------------------------------------------------------

type decrg [2]any

func (r decrg) Contains(n float64) bool {
	switch {
	case r[0] == nil:
		return n <= r[1].(float64)
	case r[1] == nil:
		return n >= r[0].(float64)
	default:
		return n >= r[0].(float64) && n <= r[1].(float64)
	}
}

type decrgs []decrg

func (rs decrgs) Contains(n float64) bool {
	for _, r := range rs {
		if r.Contains(n) {
			return true
		}
	}
	return false
}

type Decimals struct {
	sep    byte
	decs   []float64
	ranges decrgs
}

func (ds Decimals) IsEmpty() bool {
	return len(ds.decs) == 0 && len(ds.ranges) == 0
}

func (ds Decimals) String() string {
	var sb strings.Builder

	for _, n := range ds.decs {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(num.Ftoa(n))
	}

	for _, r := range ds.ranges {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}

		if r[0] != nil {
			sb.WriteString(num.Ftoa(r[0].(float64)))
		}
		sb.WriteByte(ds.sep)
		if r[1] != nil {
			sb.WriteString(num.Ftoa(r[1].(float64)))
		}
	}

	return sb.String()
}

func (ds Decimals) Contains(n float64) bool {
	return asg.Contains(ds.decs, n) || ds.ranges.Contains(n)
}

func (ds *Decimals) Parse(val string) (err error) {
	ss := str.Fields(val)

	if len(ss) == 0 {
		return
	}

	var fmin float64
	var fmax float64
	for _, s := range ss {
		smin, smax, ok := str.CutByte(s, ds.sep)
		if ok {
			switch {
			case smin == "" && smax == "": // invalid
				err = errors.New("empty")
				return
			case smin == "":
				fmax, err = strconv.ParseFloat(smax, 64)
				if err != nil {
					return
				}
				ds.ranges = append(ds.ranges, decrg{nil, fmax})
			case smax == "":
				fmin, err = strconv.ParseFloat(smin, 64)
				if err != nil {
					return
				}
				ds.ranges = append(ds.ranges, decrg{fmin, nil})
			default:
				fmin, err = strconv.ParseFloat(smin, 64)
				if err != nil {
					return
				}

				fmax, err = strconv.ParseFloat(smax, 64)
				if err != nil {
					return
				}

				switch {
				case fmin < fmax:
					ds.ranges = append(ds.ranges, decrg{fmin, fmax})
				case fmin > fmax:
					ds.ranges = append(ds.ranges, decrg{fmax, fmin})
				default:
					ds.decs = append(ds.decs, fmin)
				}
			}
		} else {
			fmin, err = strconv.ParseFloat(s, 64)
			if err != nil {
				return
			}
			ds.decs = append(ds.decs, fmin)
		}
	}
	return
}

func ParseDecimals(val string) (ds Decimals, err error) {
	ds.sep = '~'
	err = ds.Parse(val)
	return
}

func ParseUdecimals(val string) (ds Decimals, err error) {
	ds.sep = '-'
	err = ds.Parse(val)
	return
}
