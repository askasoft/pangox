package xcpts

import (
	"crypto/sha256"
	"fmt"

	"github.com/askasoft/pango/cpt"
	"github.com/askasoft/pango/str"
)

func Hash(s string) string {
	h := sha256.New()
	h.Write(str.UnsafeBytes(s))
	var buf [sha256.Size]byte
	return fmt.Sprintf("%x", h.Sum(buf[:0]))
}

func Encrypt(secret, s string) (string, error) {
	cryptor := cpt.NewAes128CBCEncryptor(secret)
	return cryptor.EncryptString(s)
}

func MustEncrypt(secret, s string) string {
	es, err := Encrypt(secret, s)
	if err != nil {
		panic(err)
	}
	return es
}

func Decrypt(secret, s string) (string, error) {
	cryptor := cpt.NewAes128CBCDecryptor(secret)
	return cryptor.DecryptString(s)
}

func MustDecrypt(secret, s string) string {
	ds, err := Decrypt(secret, s)
	if err != nil {
		panic(err)
	}
	return ds
}
