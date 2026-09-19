package xcpts

import (
	"github.com/askasoft/pango/bye"
	"github.com/askasoft/pango/cpt"
	"github.com/askasoft/pango/gog"
	"github.com/askasoft/pango/str"
)

type XCryptor struct {
	Cryptor cpt.Cryptor
	Prefix  string
}

func NewCryptor(secret, hkdfinfo string) (*XCryptor, error) {
	c, err := Cryptor(secret, hkdfinfo)
	if err != nil {
		return nil, err
	}
	return &XCryptor{Cryptor: c, Prefix: Prefix}, nil
}

func (xc *XCryptor) IsEncryptedString(s string) bool {
	return len(s) > 0 && str.StartsWith(s, xc.Prefix)
}

func (xc *XCryptor) IsEncryptedBytes(s []byte) bool {
	return len(s) > 0 && bye.StartsWith(s, str.UnsafeBytes(xc.Prefix))
}

func (xc *XCryptor) EncryptString(plain string) (string, error) {
	if xc.IsEncryptedString(plain) {
		return plain, nil
	}

	enc, err := xc.Cryptor.EncryptString(plain)
	if err != nil {
		return plain, err
	}

	return xc.Prefix + enc, nil
}

func (xc *XCryptor) DecryptString(value string) (string, error) {
	if !xc.IsEncryptedString(value) {
		return value, nil
	}

	return xc.Cryptor.DecryptString(value[len(xc.Prefix):])
}

func (xc *XCryptor) EncryptBytes(plain []byte) ([]byte, error) {
	if xc.IsEncryptedBytes(plain) {
		return plain, nil
	}

	enc, err := xc.Cryptor.EncryptBytes(plain)
	if err != nil {
		return plain, err
	}

	out := make([]byte, 0, len(xc.Prefix)+len(enc))
	out = append(out, str.UnsafeBytes(xc.Prefix)...)
	out = append(out, enc...)
	return out, nil
}

func (xc *XCryptor) DecryptBytes(value []byte) ([]byte, error) {
	if !xc.IsEncryptedBytes(value) {
		return value, nil
	}

	return xc.Cryptor.DecryptBytes(value[len(xc.Prefix):])
}

func (xc *XCryptor) MustEncryptString(plain string) string {
	return gog.Must(xc.EncryptString(plain))
}

func (xc *XCryptor) MustDecryptString(value string) string {
	return gog.Must(xc.DecryptString(value))
}

func (xc *XCryptor) MustEncryptBytes(plain []byte) []byte {
	return gog.Must(xc.EncryptBytes(plain))
}

func (xc *XCryptor) MustDecryptBytes(value []byte) []byte {
	return gog.Must(xc.DecryptBytes(value))
}
