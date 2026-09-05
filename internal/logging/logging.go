// Package logging writes a plain-text log file (nscardprice.log) under a
// configurable directory, so error traces live in one place instead of
// scattered stderr redirects. It is a thin wrapper over the standard log
// package; call Setup once per process.
package logging

import (
	"log"
	"os"
	"path/filepath"
	"sync"
)

// fileName is the fixed log file name inside the configured directory.
const fileName = "nscardprice.log"

var (
	mu     sync.Mutex
	logger *log.Logger
	file   *os.File
)

// Setup opens dir/nscardprice.log for appending (creating the directory when
// needed) and routes subsequent Errorf/Infof calls there. Passing an empty
// dir disables file logging and closes any previously opened file.
func Setup(dir string) error {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		_ = file.Close()
		file, logger = nil, nil
	}
	if dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, fileName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	file = f
	logger = log.New(f, "", log.Ldate|log.Ltime)
	return nil
}

// Errorf records an error line in the log file. It is a no-op when Setup has
// not run or was given an empty directory.
func Errorf(format string, args ...any) {
	if l := lockedLogger(); l != nil {
		l.Printf("ERROR "+format, args...)
	}
}

// Infof records an informational line in the log file.
func Infof(format string, args ...any) {
	if l := lockedLogger(); l != nil {
		l.Printf("INFO "+format, args...)
	}
}

func lockedLogger() *log.Logger {
	mu.Lock()
	defer mu.Unlock()
	return logger
}
