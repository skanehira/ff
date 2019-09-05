package main

import (
	"fmt"
	"os"

	"github.com/skanehira/ff/gui"
)

func main() {
	gui, err := gui.New()
	exitCode := 0

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		exitCode = 1
	} else {
		exitCode, err = gui.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	os.Exit(exitCode)
}
