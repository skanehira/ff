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

	"github.com/skanehira/ff/gui"
	"github.com/skanehira/ff/system"
	"gopkg.in/yaml.v2"
)

var (
	enableLog     = flag.Bool("log", false, "enable log")
	enablePreview = flag.Bool("preview", false, "enable preview panel")
	ignorecase    = flag.Bool("ignorecase", false, "ignore case when searcing")
)

var (
	ErrOpenLogFile = errors.New("cannot open log file")
)

func printError(err error) {
	fmt.Fprintln(os.Stderr, err)
}

func initConfig() gui.Config {
	var config gui.Config

	configDir, err := os.UserConfigDir()

	if err != nil {
		printError(err)
		config = gui.DefaultConfig()
	} else {
		configDir = filepath.Join(configDir, "ff")

		configFile := filepath.Join(configDir, "config.yaml")
		if system.IsExist(configFile) {
			b, err := ioutil.ReadFile(configFile)
			if err != nil {
				printError(err)
				config = gui.DefaultConfig()
			} else {
				if err := yaml.Unmarshal(b, &config); err != nil {
					printError(err)
					config = gui.DefaultConfig()
				}
			}
		} else {
			config = gui.DefaultConfig()
		}
		config.ConfigDir = configDir
		config.ConfigFile = configFile
	}

	// override config when use flags
	if *enablePreview {
		config.Preview.Enable = *enablePreview
	}

	if *enableLog && configDir != "" {
		config.Log.Enable = *enableLog
		if config.Log.File == "" {
			config.Log.File = filepath.Join(configDir, "ff.log")
		}
	}

	if *ignorecase {
		config.IgnoreCase = *ignorecase
	}

	return config
}

func initLogger(config gui.Config) error {
	var logWriter io.Writer
	var err error
	if config.Log.Enable {
		logWriter, err = os.OpenFile(os.ExpandEnv(config.Log.File),
			os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

		if err != nil {
			return fmt.Errorf("%s: %s", ErrOpenLogFile, err)
		}
		log.SetFlags(log.Lshortfile)
	} else {
		// don't print log
		logWriter = ioutil.Discard
	}

	log.SetOutput(logWriter)
	return nil
}

func run() int {
	flag.Parse()

	config := initConfig()
	if err := initLogger(config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := gui.New(config).Run(); err != nil {
		return 1
	}

	return 0
}

func main() {
	os.Exit(run())
}
