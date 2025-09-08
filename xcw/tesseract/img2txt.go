package tesseract

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/askasoft/pango/str"
)

// sudo apt install tesseract-ocr*

func ImgFileTextifyString(ctx context.Context, name string, langs ...string) (string, error) {
	bw := &bytes.Buffer{}
	err := ImgFileTextify(ctx, bw, name, langs...)
	return bw.String(), err
}

func ImgFileTextify(ctx context.Context, w io.Writer, name string, langs ...string) error {
	se := &strings.Builder{}
	args := buildTesseractArgs(name, langs...)
	cmd := exec.CommandContext(ctx, "tesseract", args...)
	cmd.Stdout = w
	cmd.Stderr = se
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tesseract: %q failed: %w - %s", cmd.String(), err, se.String())
	}
	return nil
}

func ImgReaderTextifyString(ctx context.Context, r io.Reader, langs ...string) (string, error) {
	bw := &bytes.Buffer{}
	err := ImgReaderTextify(ctx, bw, r, langs...)
	return bw.String(), err
}

func ImgReaderTextify(ctx context.Context, w io.Writer, r io.Reader, langs ...string) error {
	se := &strings.Builder{}
	args := buildTesseractArgs("-", langs...)
	cmd := exec.CommandContext(ctx, "tesseract", args...)
	cmd.Stdin = r
	cmd.Stdout = w
	cmd.Stderr = se
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tesseract: %q failed: %w - %s", cmd.String(), err, se.String())
	}
	return nil
}

func buildTesseractArgs(input string, langs ...string) []string {
	// See "man tesseract" for more options.
	// https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
	// tesseract --list-langs
	args := []string{
		input, // The input file (-: stdin)
		"-",   // The output file (stdout)
	}
	if len(langs) > 0 {
		args = append(args, "-l", str.Join(langs, "+"))
	}
	return args
}
