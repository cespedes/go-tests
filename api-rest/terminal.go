package main

// Inspired by Goji's middleware, source:
// https://github.com/zenazn/goji/tree/master/web/middleware

import (
	"io/fs"
)

// If the input file descriptor is a character device, we assume
// that it is a TTY
func isatty(fd any) bool {
	f, ok := fd.(interface {
		Stat() (fs.FileInfo, error)
	})
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	m := fs.ModeDevice | fs.ModeCharDevice
	return fi.Mode()&m == m
}
