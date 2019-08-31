package main

import (
	"fmt"
	"os"

	"github.com/skanehira/ff/gui"
)

func main() {
	exitCode, err := gui.New().Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(exitCode)
}
