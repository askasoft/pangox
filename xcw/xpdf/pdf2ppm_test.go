package xpdf

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

func testCheckPdfToPpm(t *testing.T) {
	path, err := exec.LookPath("pdftoppm")
	if path == "" || err != nil {
		t.Skip("Failed to find pdftoppm", path, err)
	}
}

func TestPdfFileImagify(t *testing.T) {
	testCheckPdfToPpm(t)

	cs := []string{"hello.pdf", "table.pdf"}

	for i, c := range cs {
		fn := testFilename(c)

		od := fn + ".png/"
		os.MkdirAll(od, 0770)

		err := PdfFileImagify(context.Background(), fn, od, "-png")
		if err != nil {
			t.Errorf("[%d] PdfFileImagify(%s): %v", i, fn, err)
			continue
		}

		defer os.RemoveAll(od)
	}
}

func TestPdfReaderImagify(t *testing.T) {
	testCheckPdfToPpm(t)

	cs := []string{"hello.pdf", "table.pdf"}

	for i, c := range cs {
		fn := testFilename(c)
		fr, err := os.Open(fn)
		if err != nil {
			t.Errorf("[%d] PdfReaderImagify(%s): %v", i, fn, err)
			continue
		}
		defer fr.Close()

		od := fn + ".png/"
		os.MkdirAll(od, 0770)

		err = PdfReaderImagify(context.Background(), fr, od, "-png")
		if err != nil {
			t.Errorf("[%d] PdfReaderImagify(%s): %v", i, fn, err)
			continue
		}

		defer os.RemoveAll(od)
	}
}

func TestPdfFileImagifyBad(t *testing.T) {
	testCheckPdfToPpm(t)

	cs := []string{"bad.pdf"}

	for i, c := range cs {
		fn := testFilename(c)

		od := fn + ".png/"
		os.MkdirAll(od, 0770)
		defer os.RemoveAll(od)

		err := PdfFileImagify(context.Background(), fn, od, "-png")
		if err == nil {
			t.Fatalf("[%d] Expected error for PdfFileImagify(%s), got nil", i, fn)
		}
	}
}
