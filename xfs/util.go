package xfs

import (
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
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

	fn := str.ToValidUTF8(file.Filename, " ")
	return xfs.SaveFile(id, fn, time.Now(), data, tag...)
}

func getLogger(loggers ...log.Logger) log.Logger {
	if len(loggers) > 0 {
		return loggers[0]
	}
	return log.GetLogger("XFS")
}

func CleanOutdatedLocalFiles(dir string, before time.Time, loggers ...log.Logger) {
	logger := getLogger(loggers...)

	logger.Debugf("CleanOutdatedLocalFiles('%s', '%s')", dir, before.Format(time.RFC3339))

	f, err := os.Open(dir)
	if err != nil {
		logger.Errorf("Open('%s') failed: %v", dir, err)
		return
	}
	defer f.Close()

	des, err := f.ReadDir(-1)
	if err != nil {
		logger.Error("ReadDir('%s') failed: %v", dir, err)
		return
	}

	for _, de := range des {
		path := filepath.Join(dir, de.Name())

		fi, err := de.Info()
		if err != nil {
			log.Errorf("DirEntry('%s').Info() failed: %v", path, err)
			continue
		}

		if de.IsDir() {
			CleanOutdatedLocalFiles(path, before, logger)

			err = fsu.DirIsEmpty(path)
			if errors.Is(err, fsu.ErrDirNotEmpty) {
				continue
			}

			if err != nil {
				log.Errorf("DirIsEmpty('%s') failed: %v", path, err)
				continue
			}
		}

		if fi.ModTime().Before(before) {
			if err := os.Remove(path); err != nil {
				log.Errorf("Remove('%s') failed: %v", path, err)
			} else {
				log.Debugf("Remove('%s') OK", path)
			}
		}
	}
}

func CleanOutdatedPrefixFiles(xfs XFS, prefix string, before time.Time, loggers ...log.Logger) {
	logger := getLogger(loggers...)

	tm := before.Format(time.RFC3339)

	logger.Debugf("CleanOutdatedPrefixFiles('%s', '%s')", prefix, tm)

	var (
		cnt int64
		err error
	)

	if prefix == "" {
		cnt, err = xfs.DeleteBefore(before)
	} else {
		cnt, err = xfs.DeletePrefixBefore(prefix, before)
	}

	if err != nil {
		logger.Errorf("CleanOutdatedPrefixFiles('%s', '%s') failed: %v", prefix, tm, err)
		return
	}

	logger.Infof("CleanOutdatedPrefixFiles('%s', '%s'): %d", prefix, tm, cnt)
}

func CleanOutdatedTaggedFiles(xfs XFS, tag string, before time.Time, loggers ...log.Logger) {
	logger := getLogger(loggers...)

	tm := before.Format(time.RFC3339)

	logger.Debugf("CleanOutdatedTaggedFiles('%s', '%s')", tag, tm)

	var (
		cnt int64
		err error
	)

	if tag == "" {
		cnt, err = xfs.DeleteBefore(before)
	} else {
		cnt, err = xfs.DeleteTaggedBefore(tag, before)
	}

	if err != nil {
		logger.Errorf("CleanOutdatedTaggedFiles('%s', '%s') failed: %v", tag, tm, err)
		return
	}

	logger.Infof("CleanOutdatedTaggedFiles('%s', '%s'): %d", tag, tm, cnt)
}
