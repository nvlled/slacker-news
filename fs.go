package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
)

func fsExists(fsys fs.FS, filename string) bool {
	file, err := fsys.Open(filename)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Print(err)
		}
		return false
	}
	defer file.Close()
	return true
}

func fsStat(fsys fs.FS, filename string) (fs.FileInfo, error) {
	file, err := fsys.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil

		}
		return nil, err
	}
	defer file.Close()
	return file.Stat()
}
