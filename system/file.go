package system

import (
	"errors"
	"io"
	"log"
	"os"
)

var (
	ErrFileExists    = errors.New("file already exists")
	ErrDirExists     = errors.New("directory already exists")
	ErrFileNotExists = errors.New("file is not exists")
)

func CopyFile(src, target string) error {
	if isExist(target) {
		return ErrFileExists
	}

	s, err := os.Open(src)
	if err != nil {
		log.Println(err)
		return err
	}
	defer s.Close()

	t, err := os.Create(target)
	if err != nil {
		return err
	}
	defer t.Close()

	if _, err := io.Copy(t, s); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func RemoveFile(file string) error {
	if !isExist(file) {
		return ErrFileNotExists
	}

	if err := os.Remove(file); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func NewFile(file string) error {
	if isExist(file) {
		return ErrFileExists
	}

	f, err := os.Create(file)
	if err != nil {
		log.Println(err)
		return err
	}
	defer f.Close()
	return nil
}

func Rename(oldpath, newpath string) error {
	if !isExist(oldpath) {
		return ErrFileNotExists
	}

	if isExist(newpath) {
		return ErrFileExists
	}

	if err := os.Rename(oldpath, newpath); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func isExist(name string) bool {
	_, err := os.Stat(name)
	if err != nil {
		log.Println(err)
	}
	return !os.IsNotExist(err)
}
