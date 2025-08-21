package xpwds

import (
	"crypto/rand"

	"github.com/askasoft/pango/str"
)

const (
	PASSWORD_INVALID_LENGTH    = "-"
	PASSWORD_NEED_UPPER_LETTER = "U"
	PASSWORD_NEED_LOWER_LETTER = "L"
	PASSWORD_NEED_NUMBER       = "N"
	PASSWORD_NEED_SYMBOL       = "S"
)

var RandomStrings = []string{
	str.UpperLetters,
	str.LowerLetters,
	str.Numbers,
	str.Symbols,
}

func RandomPassword(n int) string {
	bs := make([]byte, n)
	_, _ = rand.Read(bs)

	for i, b := range bs {
		rs := RandomStrings[i%len(RandomStrings)]
		bs[i] = rs[int(b)%len(rs)]
	}
	return str.UnsafeString(bs)
}

type PasswordPolicy struct {
	MinLength int
	MaxLength int
	Strengths []string
}

func (pp *PasswordPolicy) ValidatePassword(pwd string) (vs []string) {
	if len(pwd) < pp.MinLength || len(pwd) > pp.MaxLength {
		vs = append(vs, PASSWORD_INVALID_LENGTH)
	}

	if pwd != "" {
		for _, p := range pp.Strengths {
			switch p {
			case PASSWORD_NEED_UPPER_LETTER:
				if !str.ContainsAny(pwd, str.UpperLetters) {
					vs = append(vs, p)
				}
			case PASSWORD_NEED_LOWER_LETTER:
				if !str.ContainsAny(pwd, str.LowerLetters) {
					vs = append(vs, p)
				}
			case PASSWORD_NEED_NUMBER:
				if !str.ContainsAny(pwd, str.Numbers) {
					vs = append(vs, p)
				}
			case PASSWORD_NEED_SYMBOL:
				if !str.ContainsAny(pwd, str.Symbols) {
					vs = append(vs, p)
				}
			}
		}
	}

	return
}
