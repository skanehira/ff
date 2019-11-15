package system

import "os"

func NewDir(dir string) error {
	// TODO use inputed permission
	return os.Mkdir(dir, 0666)
}

func RemoveDirAll(dir string) error {
	return os.RemoveAll(dir)
}
