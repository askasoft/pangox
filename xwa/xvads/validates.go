package xvads

import (
	"regexp"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pangox/xwa/xargs"
)

func ValidateCIDRs(fl vad.FieldLevel) bool {
	for _, s := range str.Fields(fl.Field().String()) {
		if !vad.IsCIDR(s) {
			return false
		}
	}
	return true
}

func ValidateINI(fl vad.FieldLevel) bool {
	err := ini.NewIni().LoadData(str.NewReader(fl.Field().String()))
	return err == nil
}

func ValidateIntegers(fl vad.FieldLevel) bool {
	_, err := xargs.ParseIntegers(fl.Field().String())
	return err == nil
}

func ValidateUintegers(fl vad.FieldLevel) bool {
	_, err := xargs.ParseUintegers(fl.Field().String())
	return err == nil
}

func ValidateDecimals(fl vad.FieldLevel) bool {
	_, err := xargs.ParseDecimals(fl.Field().String())
	return err == nil
}

func ValidateUdecimals(fl vad.FieldLevel) bool {
	_, err := xargs.ParseUdecimals(fl.Field().String())
	return err == nil
}

func ValidateRegexps(fl vad.FieldLevel) bool {
	exprs := str.RemoveEmpties(str.FieldsAny(fl.Field().String(), "\r\n"))
	for _, expr := range exprs {
		_, err := regexp.Compile(expr)
		if err != nil {
			return false
		}
	}
	return true
}
