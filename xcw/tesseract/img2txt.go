package tesseract

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// sudo apt install tesseract-ocr*
// Usage:
//        tesseract FILE OUTPUTBASE [OPTIONS]... [CONFIGFILE]...

// DESCRIPTION
//        tesseract(1) is a commercial quality OCR engine originally developed at HP between 1985 and 1995. In 1995, this engine was among the top 3 evaluated by UNLV. It was
//        open-sourced by HP and UNLV in 2005, and has been developed at Google since then.

// IN/OUT ARGUMENTS
//        FILE
//            The name of the input file. This can either be an image file or a text file.
//            Most image file formats (anything readable by Leptonica) are supported.
//            A text file lists the names of all input images (one image name per line). The results will be combined in a single file for each output file format (txt, pdf,
//            hocr, xml).
//            If FILE is stdin or - then the standard input is used.
//        OUTPUTBASE
//            The basename of the output file (to which the appropriate extension will be appended). By default the output will be a text file with .txt added to the basename
//            unless there are one or more parameters set which explicitly specify the desired output.
//            If OUTPUTBASE is stdout or - then the standard output is used.
// OPTIONS
//        -c CONFIGVAR=VALUE
//            Set value for parameter CONFIGVAR to VALUE. Multiple -c arguments are allowed.
//        --dpi N
//            Specify the resolution N in DPI for the input image(s). A typical value for N is 300. Without this option, the resolution is read from the metadata included in the
//            image. If an image does not include that information, Tesseract tries to guess it.
//        -l LANG, -l SCRIPT
//            The language or script to use. If none is specified, eng (English) is assumed. Multiple languages may be specified, separated by plus characters. Tesseract uses
//            3-character ISO 639-2 language codes (see LANGUAGES AND SCRIPTS).
//        --psm N
//            Set Tesseract to only run a subset of layout analysis and assume a certain form of image. The options for N are:
//                0 = Orientation and script detection (OSD) only.
//                1 = Automatic page segmentation with OSD.
//                2 = Automatic page segmentation, but no OSD, or OCR. (not implemented)
//                3 = Fully automatic page segmentation, but no OSD. (Default)
//                4 = Assume a single column of text of variable sizes.
//                5 = Assume a single uniform block of vertically aligned text.
//                6 = Assume a single uniform block of text.
//                7 = Treat the image as a single text line.
//                8 = Treat the image as a single word.
//                9 = Treat the image as a single word in a circle.
//                10 = Treat the image as a single character.
//                11 = Sparse text. Find as much text as possible in no particular order.
//                12 = Sparse text with OSD.
//                13 = Raw line. Treat the image as a single text line,
//                     bypassing hacks that are Tesseract-specific.
//        --oem N
//            Specify OCR Engine mode. The options for N are:
//                0 = Original Tesseract only.
//                1 = Neural nets LSTM only.
//                2 = Tesseract + LSTM.
//                3 = Default, based on what is available.
//        --tessdata-dir PATH
//            Specify the location of tessdata path.
//        --user-patterns FILE
//            Specify the location of user patterns file.
//        --user-words FILE
//            Specify the location of user words file.
//        CONFIGFILE
//            The name of a config to use. The name can be a file in tessdata/configs or tessdata/tessconfigs, or an absolute or relative file path. A config is a plain text file
//            which contains a list of parameters and their values, one per line, with a space separating parameter from value.
//            Interesting config files include:
//            •   alto — Output in ALTO format (OUTPUTBASE.xml).
//            •   hocr — Output in hOCR format (OUTPUTBASE.hocr).
//            •   pdf — Output PDF (OUTPUTBASE.pdf).
//            •   tsv — Output TSV (OUTPUTBASE.tsv).
//            •   txt — Output plain text (OUTPUTBASE.txt).
//            •   get.images — Write processed input images to file (tessinput.tif).
//            •   logfile — Redirect debug messages to file (tesseract.log).
//            •   lstm.train — Output files used by LSTM training (OUTPUTBASE.lstmf).
//            •   makebox — Write box file (OUTPUTBASE.box).
//            •   quiet — Redirect debug messages to /dev/null.
//        It is possible to select several config files, for example tesseract image.png demo alto hocr pdf txt will create four output files demo.alto, demo.hocr, demo.pdf and
//        demo.txt with the OCR results.
//        Nota bene: The options -l LANG, -l SCRIPT and --psm N must occur before any CONFIGFILE.
// SINGLE OPTIONS
//        -h, --help
//            Show help message.
//        --help-extra
//            Show extra help for advanced users.
//        --help-psm
//            Show page segmentation modes.
//        --help-oem
//            Show OCR Engine modes.
//        -v, --version
//            Returns the current version of the tesseract(1) executable.
//        --list-langs
//            List available languages for tesseract engine. Can be used with --tessdata-dir PATH.
//        --print-parameters
//            Print tesseract parameters.
// LANGUAGES AND SCRIPTS
//        To recognize some text with Tesseract, it is normally necessary to specify the language(s) or script(s) of the text (unless it is English text which is supported by
//        default) using -l LANG or -l SCRIPT.
//        Selecting a language automatically also selects the language specific character set and dictionary (word list).
//        Selecting a script typically selects all characters of that script which can be from different languages. The dictionary which is included also contains a mix from
//        different languages. In most cases, a script also supports English. So it is possible to recognize a language that has not been specifically trained for by using
//        traineddata for the script it is written in.
//        More than one language or script may be specified by using +. Example: tesseract myimage.png myimage -l eng+deu+fra.
//        https://github.com/tesseract-ocr/tessdata_fast provides fast language and script models which are also part of Linux distributions.
//        For Tesseract 4, tessdata_fast includes traineddata files for the following languages:
//        afr (Afrikaans), amh (Amharic), ara (Arabic), asm (Assamese), aze (Azerbaijani), aze_cyrl (Azerbaijani - Cyrilic), bel (Belarusian), ben (Bengali), bod (Tibetan), bos
//        (Bosnian), bre (Breton), bul (Bulgarian), cat (Catalan; Valencian), ceb (Cebuano), ces (Czech), chi_sim (Chinese simplified), chi_tra (Chinese traditional), chr
//        (Cherokee), cym (Welsh), dan (Danish), deu (German), dzo (Dzongkha), ell (Greek, Modern, 1453-), eng (English), enm (English, Middle, 1100-1500), epo (Esperanto), equ
//        (Math / equation detection module), est (Estonian), eus (Basque), fas (Persian), fin (Finnish), fra (French), frk (Frankish), frm (French, Middle, ca.1400-1600), gle
//        (Irish), glg (Galician), grc (Greek, Ancient, to 1453), guj (Gujarati), hat (Haitian; Haitian Creole), heb (Hebrew), hin (Hindi), hrv (Croatian), hun (Hungarian), iku
//        (Inuktitut), ind (Indonesian), isl (Icelandic), ita (Italian), ita_old (Italian - Old), jav (Javanese), jpn (Japanese), kan (Kannada), kat (Georgian), kat_old (Georgian
//        - Old), kaz (Kazakh), khm (Central Khmer), kir (Kirghiz; Kyrgyz), kmr (Kurdish Kurmanji), kor (Korean), kor_vert (Korean vertical), kur (Kurdish), lao (Lao), lat
//        (Latin), lav (Latvian), lit (Lithuanian), ltz (Luxembourgish), mal (Malayalam), mar (Marathi), mkd (Macedonian), mlt (Maltese), mon (Mongolian), mri (Maori), msa
//        (Malay), mya (Burmese), nep (Nepali), nld (Dutch; Flemish), nor (Norwegian), oci (Occitan post 1500), ori (Oriya), osd (Orientation and script detection module), pan
//        (Panjabi; Punjabi), pol (Polish), por (Portuguese), pus (Pushto; Pashto), que (Quechua), ron (Romanian; Moldavian; Moldovan), rus (Russian), san (Sanskrit), sin
//        (Sinhala; Sinhalese), slk (Slovak), slv (Slovenian), snd (Sindhi), spa (Spanish; Castilian), spa_old (Spanish; Castilian - Old), sqi (Albanian), srp (Serbian), srp_latn
//        (Serbian - Latin), sun (Sundanese), swa (Swahili), swe (Swedish), syr (Syriac), tam (Tamil), tat (Tatar), tel (Telugu), tgk (Tajik), tgl (Tagalog), tha (Thai), tir
//        (Tigrinya), ton (Tonga), tur (Turkish), uig (Uighur; Uyghur), ukr (Ukrainian), urd (Urdu), uzb (Uzbek), uzb_cyrl (Uzbek - Cyrilic), vie (Vietnamese), yid (Yiddish), yor
//        (Yoruba)
//        To use a non-standard language pack named foo.traineddata, set the TESSDATA_PREFIX environment variable so the file can be found at
//        TESSDATA_PREFIX/tessdata/foo.traineddata and give Tesseract the argument -l foo.
//        For Tesseract 4, tessdata_fast includes traineddata files for the following scripts:
//        Arabic, Armenian, Bengali, Canadian_Aboriginal, Cherokee, Cyrillic, Devanagari, Ethiopic, Fraktur, Georgian, Greek, Gujarati, Gurmukhi, HanS (Han simplified), HanS_vert
//        (Han simplified, vertical), HanT (Han traditional), HanT_vert (Han traditional, vertical), Hangul, Hangul_vert (Hangul vertical), Hebrew, Japanese, Japanese_vert
//        (Japanese vertical), Kannada, Khmer, Lao, Latin, Malayalam, Myanmar, Oriya (Odia), Sinhala, Syriac, Tamil, Telugu, Thaana, Thai, Tibetan, Vietnamese.
//        The same languages and scripts are available from https://github.com/tesseract-ocr/tessdata_best. tessdata_best provides slow language and script models. These models
//        are needed for training. They also can give better OCR results, but the recognition takes much more time.
//        Both tessdata_fast and tessdata_best only support the LSTM OCR engine.
//        There is a third repository, https://github.com/tesseract-ocr/tessdata, with models which support both the Tesseract 3 legacy OCR engine and the Tesseract 4 LSTM OCR
//        engine.
// ENVIRONMENT VARIABLES
//        TESSDATA_PREFIX
//            If the TESSDATA_PREFIX is set to a path, then that path is used to find the tessdata directory with language and script recognition models and config files. Using
//            --tessdata-dir PATH is the recommended alternative.
//        OMP_THREAD_LIMIT
//            If the tesseract executable was built with multithreading support, it will normally use four CPU cores for the OCR process. While this can be faster for a single
//            image, it gives bad performance if the host computer provides less than four CPU cores or if OCR is made for many images. Only a single CPU core is used with
//            OMP_THREAD_LIMIT=1.

func ImgFileTextifyString(ctx context.Context, imgfile string, options ...string) (string, error) {
	bw := &bytes.Buffer{}
	err := ImgFileTextify(ctx, bw, imgfile, options...)
	return bw.String(), err
}

func ImgFileTextify(ctx context.Context, w io.Writer, imgfile string, options ...string) error {
	se := &strings.Builder{}
	args := buildTesseractArgs(imgfile, options...)
	cmd := exec.CommandContext(ctx, "tesseract", args...)
	cmd.Stdout = w
	cmd.Stderr = se
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tesseract: %q failed: %w - %s", cmd.String(), err, se.String())
	}
	return nil
}

func ImgReaderTextifyString(ctx context.Context, r io.Reader, options ...string) (string, error) {
	bw := &bytes.Buffer{}
	err := ImgReaderTextify(ctx, bw, r, options...)
	return bw.String(), err
}

func ImgReaderTextify(ctx context.Context, w io.Writer, r io.Reader, options ...string) error {
	se := &strings.Builder{}
	args := buildTesseractArgs("-", options...)
	cmd := exec.CommandContext(ctx, "tesseract", args...)
	cmd.Stdin = r
	cmd.Stdout = w
	cmd.Stderr = se
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tesseract: %q failed: %w - %s", cmd.String(), err, se.String())
	}
	return nil
}

// See "man tesseract" for more options.
// https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
// tesseract --list-options
func buildTesseractArgs(input string, options ...string) []string {
	args := []string{
		input, // The input file (-: stdin)
		"-",   // The output file (stdout)
	}
	return append(args, options...)
}
