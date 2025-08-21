package xpwds

import (
	"testing"

	"github.com/askasoft/pango/str"
)

func TestRandomPassword(t *testing.T) {
	for range 10 {
		pwd := RandomPassword(12)
		if len(pwd) != 12 {
			t.Errorf("Expected password length 12, got %d", len(pwd))
		}
		if !str.ContainsAny(pwd, str.UpperLetters) {
			t.Error("Password must contain at least one uppercase letter")
		}
		if !str.ContainsAny(pwd, str.LowerLetters) {
			t.Error("Password must contain at least one lowercase letter")
		}
		if !str.ContainsAny(pwd, str.Numbers) {
			t.Error("Password must contain at least one number")
		}
		if !str.ContainsAny(pwd, str.Symbols) {
			t.Error("Password must contain at least one symbol")
		}
	}
}
