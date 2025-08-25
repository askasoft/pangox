package xmwas

import (
	"net/http"

	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func InvalidToken(c *xin.Context) {
	err := tbs.GetText(c.Locale, "error.forbidden.token", "Invalid Token.")

	if xin.IsAjax(c) {
		c.JSON(http.StatusForbidden, xin.H{"error": err})
	} else {
		c.String(http.StatusForbidden, err)
	}
	c.Abort()
}

func BodyTooLarge(c *xin.Context) {
	err := tbs.Format(c.Locale, "error.request.toolarge", num.HumanSize(float64(XSL.MaxBodySize)))

	if xin.IsAjax(c) {
		c.JSON(http.StatusRequestEntityTooLarge, xin.H{"error": err})
	} else {
		c.String(http.StatusRequestEntityTooLarge, err)
	}
	c.Abort()
}
