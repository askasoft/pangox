package xargs

import (
	"time"

	"github.com/askasoft/pango/tmu"
)

type DateRangeArg struct {
	DateMin time.Time `json:"date_min,omitempty" form:"date_min,strip"`
	DateMax time.Time `json:"date_max,omitempty" form:"date_max,strip" validate:"omitempty,gtefield=DateMin"`
}

func (dra *DateRangeArg) Normalize() {
	if !dra.DateMin.IsZero() {
		dra.DateMin = tmu.TruncateHours(dra.DateMin)
	}
	if !dra.DateMax.IsZero() {
		dra.DateMax = tmu.TruncateHours(dra.DateMax).Add(time.Hour*24 - time.Nanosecond)
	}
}

type TimeRangeArg struct {
	TimeMin time.Time `json:"time_min,omitempty" form:"time_min,strip"`
	TimeMax time.Time `json:"time_max,omitempty" form:"time_max,strip" validate:"omitempty,gtefield=TimeMin"`
}
