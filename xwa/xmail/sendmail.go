package xmail

import (
	"crypto/tls"
	"errors"
	"html"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/log"
	"github.com/askasoft/pango/net/email"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pangox/xwa/xtpls"
)

func SendTemplateHTMLEmail(locale, tplName, toAddr string, data any) error {
	var sb str.Builder

	if err := xtpls.XHT.Render(&sb, locale, tplName, data); err != nil {
		return err
	}

	sub, msg, _ := str.CutByte(str.Strip(sb.String()), '\n')
	sub = str.Strip(sub)
	sub = str.TrimPrefix(sub, "<s>")
	sub = str.TrimSuffix(sub, "</s>")
	sub = html.EscapeString(sub)
	sub = str.Strip(sub)
	msg = str.Strip(msg)

	return SendHTMLEmail(toAddr, sub, msg)
}

func SendHTMLEmail(toAddr, subject, message string) error {
	log.GetLogger("SMTP").Infof("Send email to %q - %q", toAddr, subject)

	sec := ini.GetSection("smtp")
	if sec == nil {
		return errors.New("missing [smtp] settings")
	}

	em := &email.Email{}
	if err := em.SetFrom(sec.GetString("fromaddr")); err != nil {
		return err
	}
	if err := em.AddTo(toAddr); err != nil {
		return err
	}
	em.Subject = subject
	em.SetHTMLMsg(message)

	sender := &email.SMTPSender{
		Host:     sec.GetString("host", "localhost"),
		Port:     sec.GetInt("port", 25),
		Username: sec.GetString("username"),
		Password: sec.GetString("password"),
	}
	sender.Helo = "localhost"
	sender.Timeout = sec.GetDuration("timeout")
	if sec.GetBool("insecure") {
		sender.TLSConfig = &tls.Config{ServerName: sender.Host, InsecureSkipVerify: true} //nolint: gosec
	}

	if err := sender.Dial(); err != nil {
		return err
	}
	defer sender.Close()

	if err := sender.Login(); err != nil {
		return err
	}

	return sender.Send(em)
}
