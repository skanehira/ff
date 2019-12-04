package system

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/otiai10/copy"
)

var (
	ErrFileExists    = errors.New("file already exists")
	ErrDirExists     = errors.New("directory already exists")
	ErrFileNotExists = errors.New("file is not exists")
)

var OpenCmd string

func Copy(src, target string) error {
	return copy.Copy(src, target)
}

func RemoveFile(file string) error {
	if !IsExist(file) {
		return ErrFileNotExists
	}

	if err := os.Remove(file); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func NewFile(file string) error {
	if IsExist(file) {
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
	if !IsExist(oldpath) {
		return ErrFileNotExists
	}

	if IsExist(newpath) {
		return ErrFileExists
	}

	if err := os.Rename(oldpath, newpath); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func IsExist(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func NewDir(dir string) error {
	// TODO use inputed permission
	return os.Mkdir(dir, 0777)
}

func RemoveDirAll(dir string) error {
	return os.RemoveAll(dir)
}

func Open(name string) error {
	cmd := exec.Command(OpenCmd, name)
	buf := bytes.Buffer{}
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err, buf.String())
	}

	return nil
}
