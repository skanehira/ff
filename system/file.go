package system

import (
	"errors"
	"io"
	"os"
)

var (
	ErrFileExists    = errors.New("file already exists")
	ErrDirExists     = errors.New("directory already exists")
	ErrFileNotExists = errors.New("file is not exists")
)

func CopyFile(src, target string) error {
	_, err := os.Stat(src)
	if !os.IsExist(err) {
		return ErrFileExists
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	t, err := os.Create(target)
	if err != nil {
		return err
	}
	defer t.Close()

	if _, err := io.Copy(t, s); err != nil {
		return err
	}

	return nil
}

func RemoveFile(file string) error {
	if !isExist(file) {
		return ErrFileNotExists
	}

	if err := os.Remove(file); err != nil {
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
		return err
	}
	defer f.Close()
	return nil
}

func isExist(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}
