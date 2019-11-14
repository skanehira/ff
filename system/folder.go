package system

import "os"

func MakeDir(dir string) error {
	// TODO use inputed permission
	return os.Mkdir(dir, 0666)
}
