package xcpts

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestHash(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "",
			want:  fmt.Sprintf("%x", sha256.Sum256([]byte(""))),
		},
		{
			input: "hello",
			want:  fmt.Sprintf("%x", sha256.Sum256([]byte("hello"))),
		},
		{
			input: "GoLang",
			want:  fmt.Sprintf("%x", sha256.Sum256([]byte("GoLang"))),
		},
		{
			input: "The quick brown fox jumps over the lazy dog",
			want:  fmt.Sprintf("%x", sha256.Sum256([]byte("The quick brown fox jumps over the lazy dog"))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Hash(tt.input)
			if got != tt.want {
				t.Errorf("Hash(%q) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	username := "x@x.com"
	password := "trusttrusttrusttrust"

	encpass := MustEncrypt(username, password)

	decpass := MustDecrypt(username, encpass)
	if password != decpass {
		t.Errorf("%s: E(%s) != D(%s)", password, encpass, decpass)
	}
}
