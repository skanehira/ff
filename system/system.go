package system

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/otiai10/copy"
)

var (
	ErrFileExists    = errors.New("file already exists")
	ErrDirExists     = errors.New("directory already exists")
	ErrFileNotExists = errors.New("file is not exists")
)

func CopyFile(src, target string) error {
	return copy.Copy(src, target)
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

func NewDir(dir string) error {
	// TODO use inputed permission
	return os.Mkdir(dir, 0777)
}

func RemoveDirAll(dir string) error {
	return os.RemoveAll(dir)
}

func Open(name string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", name).Run()
	case "windows":
		return exec.Command("cmd", "/c", "start", name).Run()
	case "linux":
		return exec.Command("xdg-open", name).Run()
	}

	return nil
}
