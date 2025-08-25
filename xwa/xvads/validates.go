package xvads

import (
	"regexp"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pangox/xwa/xargs"
)

func RegisterValidations(vad *vad.Validate) {
	vad.RegisterValidation("ini", ValidateINI)
	vad.RegisterValidation("cidrs", ValidateCIDRs)
	vad.RegisterValidation("integers", ValidateIntegers)
	vad.RegisterValidation("uintegers", ValidateUintegers)
	vad.RegisterValidation("decimals", ValidateDecimals)
	vad.RegisterValidation("udecimals", ValidateUdecimals)
	vad.RegisterValidation("regexps", ValidateRegexps)
}

func ValidateINI(fl vad.FieldLevel) bool {
	vad.MustStringField("ini", fl)

	err := ini.NewIni().LoadData(str.NewReader(fl.Field().String()))
	return err == nil
}

func ValidateCIDRs(fl vad.FieldLevel) bool {
	vad.MustStringField("cidrs", fl)

	for _, s := range str.Fields(fl.Field().String()) {
		if !vad.IsCIDR(s) {
			return false
		}
	}
	return true
}

func ValidateIntegers(fl vad.FieldLevel) bool {
	vad.MustStringField("integers", fl)

	_, err := xargs.ParseIntegers(fl.Field().String())
	return err == nil
}

func ValidateUintegers(fl vad.FieldLevel) bool {
	vad.MustStringField("uintegers", fl)

	_, err := xargs.ParseUintegers(fl.Field().String())
	return err == nil
}

func ValidateDecimals(fl vad.FieldLevel) bool {
	vad.MustStringField("decimals", fl)

	_, err := xargs.ParseDecimals(fl.Field().String())
	return err == nil
}

func ValidateUdecimals(fl vad.FieldLevel) bool {
	vad.MustStringField("udecimals", fl)

	_, err := xargs.ParseUdecimals(fl.Field().String())
	return err == nil
}

func ValidateRegexps(fl vad.FieldLevel) bool {
	vad.MustStringField("regexps", fl)

	exprs := str.RemoveEmpties(str.FieldsAny(fl.Field().String(), "\r\n"))
	for _, expr := range exprs {
		_, err := regexp.Compile(expr)
		if err != nil {
			return false
		}
	}
	return true
}
