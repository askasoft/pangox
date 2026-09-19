package xcpts

import (
	"crypto/sha256"
	"fmt"

	"github.com/askasoft/pango/bye"
	"github.com/askasoft/pango/cpt"
	"github.com/askasoft/pango/gog"
	"github.com/askasoft/pango/str"
)

var (
	Cryptor = cpt.NewAes256GCMCryptor
	Prefix  = "enc:"
)

func Hash(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256(str.UnsafeBytes(s)))
}

func IsEncryptedString(s string) bool {
	return len(s) > 0 && str.StartsWith(s, Prefix)
}

func IsEncryptedBytes(s []byte) bool {
	return len(s) > 0 && bye.StartsWith(s, str.UnsafeBytes(Prefix))
}

func EncryptString(secret, hkdfinfo, plain string) (string, error) {
	if IsEncryptedString(plain) {
		return plain, nil
	}

	c, err := Cryptor(secret, hkdfinfo)
	if err != nil {
		return plain, err
	}

	enc, err := c.EncryptString(plain)
	if err != nil {
		return plain, err
	}

	return Prefix + enc, nil
}

func DecryptString(secret, hkdfinfo, value string) (string, error) {
	if !IsEncryptedString(value) {
		return value, nil
	}

	c, err := Cryptor(secret, hkdfinfo)
	if err != nil {
		return value, err
	}
	return c.DecryptString(value[len(Prefix):])
}

func EncryptBytes(secret, hkdfinfo string, plain []byte) ([]byte, error) {
	if IsEncryptedBytes(plain) {
		return plain, nil
	}

	c, err := Cryptor(secret, hkdfinfo)
	if err != nil {
		return plain, err
	}

	enc, err := c.EncryptBytes(plain)
	if err != nil {
		return plain, err
	}

	out := make([]byte, 0, len(Prefix)+len(enc))
	out = append(out, str.UnsafeBytes(Prefix)...)
	out = append(out, enc...)
	return out, nil
}

func DecryptBytes(secret, hkdfinfo string, value []byte) ([]byte, error) {
	if !IsEncryptedBytes(value) {
		return value, nil
	}

	c, err := Cryptor(secret, hkdfinfo)
	if err != nil {
		return value, err
	}

	return c.DecryptBytes(value[len(Prefix):])
}

func MustEncryptString(secret, hkdfinfo, value string) string {
	return gog.Must(EncryptString(secret, hkdfinfo, value))
}

func MustDecryptString(secret, hkdfinfo, value string) string {
	return gog.Must(DecryptString(secret, hkdfinfo, value))
}

func MustEncryptBytes(secret, hkdfinfo string, plain []byte) []byte {
	return gog.Must(EncryptBytes(secret, hkdfinfo, plain))
}

func MustDecryptBytes(secret, hkdfinfo string, value []byte) []byte {
	return gog.Must(DecryptBytes(secret, hkdfinfo, value))
}
