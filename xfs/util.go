package xfs

import (
	"mime/multipart"
	"os"
	"time"

	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/net/httpx"
	"github.com/askasoft/pango/str"
)

func SaveLocalFile(xfs XFS, id string, filename string, tag ...string) (*File, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	data, err := fsu.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return xfs.SaveFile(id, filename, fi.ModTime(), data, tag...)
}

func SaveUploadedFile(xfs XFS, id string, file *multipart.FileHeader, tag ...string) (*File, error) {
	data, err := httpx.ReadMultipartFile(file)
	if err != nil {
		return nil, err
	}

	filename := str.ToValidUTF8(file.Filename, " ")
	return xfs.SaveFile(id, filename, time.Now(), data, tag...)
}
