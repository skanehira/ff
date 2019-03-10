package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var (
	entryManager   *EntryManager
	historyManager *HistoryManager
)

func initialize() (int, error) {
	// init logger
	home, err := homedir.Dir()
	if err != nil {
		log.Println("cannot get home dir, cause:", err)
		return 1, err
	}
	logFile, err := os.OpenFile(filepath.Join(home, "filemanager.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Println("cannot open log file, cause:", err)
		return 1, err
	}

	log.SetOutput(logFile)

	// init entry manager
	entryManager = NewEntryManager()
	// init history manager
	historyManager = NewHistoryManager()

	return 0, nil
}

func run() (int, error) {
	// initialize application
	exitCode, err := initialize()
	if err != nil {
		return exitCode, err
	}

	app := tview.NewApplication()
	// get current path
	currentDir, err := os.Getwd()
	if err != nil {
		return 1, err
	}

	historyManager.Save(currentDir)

	inputPath := tview.NewInputField().SetText(currentDir)

	entryManager.SetEntries(currentDir)
	entryManager.SetColumns()

	entryManager.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		// go to input view
		case event.Key() == tcell.KeyTab:
			app.SetFocus(inputPath)
		// go to prev history
		case event.Key() == tcell.KeyCtrlH:
			path := historyManager.Previous()
			if path != "" {
				inputPath.SetText(path)
				entryManager.SetEntries(path)
				entryManager.SetColumns()
			}
		// go to next history
		case event.Key() == tcell.KeyCtrlL:
			path := historyManager.Next()
			if path != "" {
				inputPath.SetText(path)
				entryManager.SetEntries(path)
				entryManager.SetColumns()
			}
		// go to specified dir
		// TODO save position info
		case event.Rune() == 'h':
			path := filepath.Dir(inputPath.GetText())
			if path != "" {
				inputPath.SetText(path)
				entryManager.SetEntries(path)
				entryManager.SetColumns()
			}
		// go to parent dir
		case event.Rune() == 'l':
			entries := entryManager.Entries()
			if len(entries) == 0 {
				return event
			}

			row, column := entryManager.GetSelection()
			entry := entries[row-1]

			if entry.IsDir {
				path := path.Join(inputPath.GetText(), entryManager.GetCell(row, column).Text)
				historyManager.Save(path)
				inputPath.SetText(path)
				entryManager.SetEntries(path)
				entryManager.SetColumns()
			}
		// TODO mark file or dir
		case event.Rune() == ' ':
		}

		return event
	})

	inputPath.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
		if key == tcell.KeyEnter {
			path := inputPath.GetText()
			historyManager.Save(path)
			entryManager.SetEntries(path)
			entryManager.SetColumns()

			app.SetFocus(entryManager)
		}

	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(entryManager)
		}
		return event
	})

	grid := tview.NewGrid().SetRows(1, 0)
	grid.AddItem(inputPath, 0, 0, 1, 1, 0, 0, true)
	grid.AddItem(entryManager, 1, 0, 2, 2, 0, 0, true)

	if err := app.SetRoot(grid, true).SetFocus(entryManager).Run(); err != nil {
		app.Stop()
		return 1, err
	}

	return 0, nil
}

func main() {
	exitCode, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(exitCode)
}
