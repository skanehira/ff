package system

import (
	"errors"
	"io"
	"os"
)

var (
	ErrFileExist = errors.New("file already exist")
)

func CopyFile(src, target string) error {
	_, err := os.Stat(src)
	if !os.IsExist(err) {
		return ErrFileExist
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
