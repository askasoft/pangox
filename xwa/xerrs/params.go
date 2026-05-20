package xerrs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xin/binding"
)

type ParamError struct {
	Param   string `json:"param,omitempty"`
	Label   string `json:"label,omitempty"`
	Message string `json:"message,omitempty"`
}

func (pe *ParamError) Error() string {
	if pe.Label == "" || pe.Label == pe.Param {
		return pe.Param + ": " + pe.Message
	}
	return pe.Param + " [" + pe.Label + "]: " + pe.Message
}

func InvalidParamError(locale, namespace, field string, errkey ...string) *ParamError {
	label := tbs.GetText(locale, namespace+field, field)
	fe := &ParamError{
		Param:   field,
		Label:   label,
		Message: tbs.GetText(locale, asg.First(errkey, "error.param.invalid")),
	}
	return fe
}

func InvalidIDError(locale string) error {
	return tbs.Error(locale, "error.param.id")
}

func InvalidRequestError(locale string) error {
	return tbs.Error(locale, "error.request.invalid")
}

// AddBindErrors translate bind or validate errors and add it to context
func AddBindErrors(c *xin.Context, err error, ns string) {
	TranslateBindErrors(c.Locale, err, ns, func(err error) {
		c.AddError(err)
	})
}

// FormatBindErrors translate bind or validate errors and merge it to a new error
func FormatBindErrors(locale string, err error, ns string) error {
	var sb strings.Builder
	TranslateBindErrors(locale, err, ns, func(err error) {
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(err.Error())
	})
	return errors.New(sb.String())
}

// TranslateBindErrors translate bind or validate errors
// FieldBindErrors:
//  1. {xxx}.error.{field}
//  2. error.param.invalid
//
// ValidationErrors:
//  1. {xxx}.error.{field}.{tag}
//  2. {xxx}.error.param.{tag}
//  3. error.param.{tag}
//  4. error.param.invalid
func TranslateBindErrors(locale string, err error, ns string, tf func(error)) {
	if fbes, ok := binding.AsFieldBindErrors(err); ok {
		for _, fbe := range *fbes {
			tf(ConvertFieldBindError(locale, ns, fbe.Field))
		}
		return
	}

	if ves, ok := vad.AsValidationErrors(err); ok {
		for _, fe := range *ves {
			fk := str.SnakeCase(fe.Field())

			if le, ok := fe.Cause().(ILocaleError); ok {
				fn := tbs.GetText(locale, ns+fk, fk)
				em := le.LocaleError(locale)
				tf(&ParamError{Param: fk, Label: fn, Message: em})
				continue
			}

			tf(ConvertValidationError(locale, ns, fk, fe.Tag(), fe.Param()))
		}
		return
	}

	tf(err)
}

// ConvertFieldBindError convert FieldBindError to ParamError.
//  1. {xxx}.error.{field}
//  2. error.param.invalid
func ConvertFieldBindError(locale string, ns, field string) error {
	fk := str.SnakeCase(field)
	fn := tbs.GetText(locale, ns+fk, fk)
	fm := tbs.GetText(locale, ns+"error."+fk)
	if fm == "" {
		fm = tbs.GetText(locale, "error.param.invalid")
	}
	return &ParamError{Param: fk, Label: fn, Message: fm}
}

// ConvertValidationError convert ValidationError to ParamError.
//  1. {xxx}.error.{field}.{tag}
//  2. {xxx}.error.param.{tag}
//  3. error.param.{tag}
//  4. error.param.invalid
func ConvertValidationError(locale string, ns, field, tag, param string) error {
	fn := ""
	fm := tbs.GetText(locale, ns+"error."+field+"."+tag)
	if fm == "" {
		fm = tbs.GetText(locale, ns+"error."+field)
		if fm == "" {
			fn = tbs.GetText(locale, ns+field, field)
			fm = tbs.GetText(locale, ns+"error.param."+tag)
			if fm == "" {
				fm = tbs.GetText(locale, "error.param."+tag)
				if fm == "" {
					fm = tbs.GetText(locale, "error.param.invalid")
				}
			}
		}
	}

	em := fm
	if str.Contains(fm, "%s") {
		if str.EndsWith(tag, "field") {
			tk := str.SnakeCase(param)
			param = tbs.GetText(locale, ns+tk, tk)
		}
		em = fmt.Sprintf(fm, param)
	}

	return &ParamError{Param: field, Label: fn, Message: em}
}

func E(c *xin.Context) xin.H {
	errs := []any{}
	for _, e := range c.Errors {
		if pe, ok := e.(*ParamError); ok { //nolint: errorlint
			errs = append(errs, pe)
		} else {
			errs = append(errs, e.Error())
		}
	}

	var err any
	if len(errs) == 1 {
		err = errs[0]
	} else {
		err = errs
	}

	h := xin.H{
		"error": err,
	}
	return h
}
