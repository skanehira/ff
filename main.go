package main

import (
	"fmt"
	"os"

	"github.com/skanehira/ff/gui"
)

func main() {
	gui := gui.New()
	exitCode, err := gui.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(exitCode)
}
