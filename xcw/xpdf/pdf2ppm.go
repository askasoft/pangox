package xpdf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// sudo apt install poppler-utils
// Usage: pdftoppm [options] [PDF-file [PPM-file-prefix]]
//   -f <int>                                 : first page to print
//   -l <int>                                 : last page to print
//   -o                                       : print only odd pages
//   -e                                       : print only even pages
//   -singlefile                              : write only the first page and do not add digits
//   -scale-dimension-before-rotation         : for rotated pdf, resize dimensions before the rotation
//   -r <fp>                                  : resolution, in DPI (default is 150)
//   -rx <fp>                                 : X resolution, in DPI (default is 150)
//   -ry <fp>                                 : Y resolution, in DPI (default is 150)
//   -scale-to <int>                          : scales each page to fit within scale-to*scale-to pixel box
//   -scale-to-x <int>                        : scales each page horizontally to fit in scale-to-x pixels
//   -scale-to-y <int>                        : scales each page vertically to fit in scale-to-y pixels
//   -x <int>                                 : x-coordinate of the crop area top left corner
//   -y <int>                                 : y-coordinate of the crop area top left corner
//   -W <int>                                 : width of crop area in pixels (default is 0)
//   -H <int>                                 : height of crop area in pixels (default is 0)
//   -sz <int>                                : size of crop square in pixels (sets W and H)
//   -cropbox                                 : use the crop box rather than media box
//   -hide-annotations                        : do not show annotations
//   -mono                                    : generate a monochrome PBM file
//   -gray                                    : generate a grayscale PGM file
//   -displayprofile <string>                 : ICC color profile to use as the display profile
//   -defaultgrayprofile <string>             : ICC color profile to use as the DefaultGray color space
//   -defaultrgbprofile <string>              : ICC color profile to use as the DefaultRGB color space
//   -defaultcmykprofile <string>             : ICC color profile to use as the DefaultCMYK color space
//   -sep <string>                            : single character separator between name and page number, default -
//   -forcenum                                : force page number even if there is only one page
//   -png                                     : generate a PNG file
//   -jpeg                                    : generate a JPEG file
//   -jpegcmyk                                : generate a CMYK JPEG file
//   -jpegopt <string>                        : jpeg options, with format <opt1>=<val1>[,<optN>=<valN>]*
//   -overprint                               : enable overprint
//   -tiff                                    : generate a TIFF file
//   -tiffcompression <string>                : set TIFF compression: none, packbits, jpeg, lzw, deflate
//   -freetype <string>                       : enable FreeType font rasterizer: yes, no
//   -thinlinemode <string>                   : set thin line mode: none, solid, shape. Default: none
//   -aa <string>                             : enable font anti-aliasing: yes, no
//   -aaVector <string>                       : enable vector anti-aliasing: yes, no
//   -opw <string>                            : owner password (for encrypted files)
//   -upw <string>                            : user password (for encrypted files)
//   -q                                       : don't print any messages or errors
//   -progress                                : print progress info
//   -v                                       : print copyright and version info
//   -h                                       : print usage information
//   -help                                    : print usage information
//   --help                                   : print usage information
//   -?                                       : print usage information

// PdfFileImagify Convert pdf file to images.
// options: see "pdftoppm -h"
func PdfFileImagify(ctx context.Context, pdffile, outdir string, options ...string) error {
	se := &strings.Builder{}
	args := buildPdfToPpmArgs(pdffile, outdir, options...)
	cmd := exec.CommandContext(ctx, "pdftoppm", args...)
	cmd.Stderr = se
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("xpdf: %q failed: %w - %s", cmd.String(), err, se.String())
	}
	return nil
}

// PdfBytesImagify Convert pdf data to images.
// options: see "pdftoppm -h"
func PdfBytesImagify(ctx context.Context, bs []byte, outdir string, options ...string) error {
	return PdfReaderImagify(ctx, bytes.NewReader(bs), outdir, options...)
}

// PdfReaderImagify Convert pdf reader to images.
// options: see "pdftoppm -h"
func PdfReaderImagify(ctx context.Context, r io.Reader, outdir string, options ...string) error {
	se := &strings.Builder{}
	args := buildPdfToPpmArgs("-", outdir, options...)
	cmd := exec.CommandContext(ctx, "pdftoppm", args...)
	cmd.Stdin = r
	cmd.Stderr = se
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("xpdf: %q failed: %w - %s", cmd.String(), err, se.String())
	}
	return nil
}

func buildPdfToPpmArgs(input, outdir string, options ...string) []string {
	return append(options, input, outdir)
}
