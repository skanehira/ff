package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/skanehira/ff/gui"
)

var (
	enableLog     = flag.Bool("log", false, "enable log")
	enablePreview = flag.Bool("preview", false, "enable preview panel")
	ignorecase    = flag.Bool("ignorecase", false, "ignore case when searcing")
)

var (
	ErrGetHomeDir  = errors.New("cannot get home dir")
	ErrOpenLogFile = errors.New("cannot open log file")
)

func run() int {
	if err := initLogger(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := gui.New(*enablePreview, *ignorecase).Run(); err != nil {
		return 1
	}

	return 0
}

func initLogger() error {
	var logWriter io.Writer
	if *enableLog {
		home, err := homedir.Dir()
		if err != nil {
			return fmt.Errorf("%s: %s", ErrGetHomeDir, err)
		}

		logWriter, err = os.OpenFile(filepath.Join(home, "ff.log"),
			os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

		if err != nil {
			return fmt.Errorf("%s: %s", ErrOpenLogFile, err)
		}
		log.SetFlags(log.Lshortfile)
	} else {
		// no print log
		logWriter = ioutil.Discard
	}

	log.SetOutput(logWriter)
	return nil
}

func main() {
	flag.Parse()
	os.Exit(run())
}
