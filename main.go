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

var entryManager *EntryManager

// MoveHistory have the move history
type MoveHistory struct {
	idx       int
	histories []string
}

// Save save the move history
func (p *MoveHistory) Save(path string) {
	count := len(p.histories)

	// if not have history
	if count == 0 {
		p.histories = append(p.histories, path)
	} else if p.idx == count-1 {
		p.histories = append(p.histories, path)
		p.idx++
	} else {
		p.histories = append(p.histories[:p.idx+1], path)
		p.idx = len(p.histories) - 1
	}
}

// Previous return the previous history
func (p *MoveHistory) Previous() string {
	count := len(p.histories)
	if count == 0 {
		return ""
	}

	p.idx--
	if p.idx < 0 {
		p.idx = 0
	}
	return p.histories[p.idx]
}

// Next return the next history
func (p *MoveHistory) Next() string {
	count := len(p.histories)
	if count == 0 {
		return ""
	}

	p.idx++
	if p.idx >= count {
		p.idx = count - 1
	}
	return p.histories[p.idx]
}

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

	history := &MoveHistory{}
	history.Save(currentDir)

	inputPath := tview.NewInputField().SetText(currentDir)

	entryManager.SetEntries(currentDir)
	entryManager.SetColumns()

	entryManager.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}

	}).SetSelectedFunc(func(row int, column int) {
		entries := entryManager.Entries()
		if len(entries) == 0 {
			return
		}

		entry := entries[row-1]

		if entry.IsDir {
			path := path.Join(inputPath.GetText(), entryManager.GetCell(row, column).Text)
			history.Save(path)
			inputPath.SetText(path)
			entryManager.SetEntries(path)
			entryManager.SetColumns()
		}
	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(inputPath)
		}

		if event.Rune() == 'h' {
			path := history.Previous()
			if path != "" {
				inputPath.SetText(path)
				entryManager.SetEntries(path)
				entryManager.SetColumns()
			}
		}

		if event.Rune() == 'l' {
			path := history.Next()
			if path != "" {
				inputPath.SetText(path)
				entryManager.SetEntries(path)
				entryManager.SetColumns()
			}
		}
		return event
	})

	inputPath.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
		if key == tcell.KeyEnter {
			path := inputPath.GetText()
			history.Save(path)
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
