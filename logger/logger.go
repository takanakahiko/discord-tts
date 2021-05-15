package logger

import (
	"log"
	"path/filepath"
	"runtime"
)

func PrintError(err error) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		fname := filepath.Base(file)
		log.Printf(`%s:%d : %s\n`, fname, line, err.Error())
		return
	}
	log.Printf(`%s\n`, err.Error())
}
