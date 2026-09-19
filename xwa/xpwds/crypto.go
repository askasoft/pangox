package xpwds

import (
	"github.com/askasoft/pango/cpt"
	"github.com/askasoft/pango/str"
)

var (
	NewCryptor = cpt.NewAes256GCMCryptor
	Prefix     = "enc:"
	HkdfInfo   = "userpass"
)

func IsEncrypted(s string) bool {
	return str.StartsWith(s, Prefix)
}

func Encrypt(secret, value string) (string, error) {
	if value == "" || IsEncrypted(value) {
		return value, nil
	}

	c, err := NewCryptor(secret, HkdfInfo)
	if err != nil {
		return value, err
	}

	enc, err := c.EncryptString(value)
	if err != nil {
		return value, err
	}

	return Prefix + enc, nil
}

func Decrypt(secret, value string) (string, error) {
	if !IsEncrypted(value) {
		return value, nil
	}

	c, err := NewCryptor(secret, HkdfInfo)
	if err != nil {
		return value, err
	}
	return c.DecryptString(value[len(Prefix):])
}
