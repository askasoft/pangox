package xfsus

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/askasoft/pango/fsu"
	"github.com/askasoft/pango/log"
)

func getLogger(loggers ...log.Logger) log.Logger {
	if len(loggers) > 0 {
		return loggers[0]
	}
	return log.Default()
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
