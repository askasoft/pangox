package xwa

import (
	"os"
	"testing"

	"github.com/askasoft/pango/fsu"
)

func TestInitLogs(t *testing.T) {
	LogConfigFile = "test.ini"

	defer os.Remove(LogConfigFile)

	fsu.WriteString(LogConfigFile, `
writer = stdout

[level]
* = TRACE

[writer.stdout]
format = %t{2006-01-02T15:04:05} [%p] - %m%n%T
`, 0666)

	if err := InitLogs(); err != nil {
		t.Fatal(err)
	}
}
