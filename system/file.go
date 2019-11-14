package system

import (
	"io"
	"os"
)

func CopyFile(src, target string) error {
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
