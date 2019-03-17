package gui

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/gdamore/tcell"
	"github.com/mitchellh/go-homedir"
	"github.com/rivo/tview"
)

// Gui gui have some manager
type Gui struct {
	InputPath      tview.InputField
	HistoryManager *HistoryManager
	EntryManager   *EntryManager
	App            *tview.Application
}

// New create new gui
func New() *Gui {
	// init logger
	home, err := homedir.Dir()
	if err != nil {
		panic(fmt.Sprintf("cannot get home dir, cause: %s", err))
	}
	logFile, err := os.OpenFile(filepath.Join(home, "filemanager.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		panic(fmt.Sprintf("cannot open log file, cause: %s", err))
	}

	log.SetOutput(logFile)

	return &Gui{
		EntryManager:   NewEntryManager(),
		HistoryManager: NewHistoryManager(),
		App:            tview.NewApplication(),
	}
}

// Run run ff
func (gui *Gui) Run() (int, error) {
	// get current path
	currentDir, err := os.Getwd()
	if err != nil {
		return 1, err
	}

	gui.HistoryManager.Save(currentDir)

	inputPath := tview.NewInputField().SetText(currentDir)

	gui.EntryManager.SetEntries(currentDir)
	gui.EntryManager.SetColumns()

	gui.EntryManager.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			gui.App.Stop()
		}
	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		// go to input view
		case event.Key() == tcell.KeyTab:
			gui.App.SetFocus(inputPath)
		// go to prev history
		case event.Key() == tcell.KeyCtrlH:
			path := gui.HistoryManager.Previous()
			if path != "" {
				inputPath.SetText(path)
				gui.EntryManager.SetEntries(path)
				gui.EntryManager.SetColumns()
			}
		// go to next history
		case event.Key() == tcell.KeyCtrlL:
			path := gui.HistoryManager.Next()
			if path != "" {
				inputPath.SetText(path)
				gui.EntryManager.SetEntries(path)
				gui.EntryManager.SetColumns()
			}
		// go to specified dir
		// TODO save position info
		case event.Rune() == 'h':
			path := filepath.Dir(inputPath.GetText())
			if path != "" {
				inputPath.SetText(path)
				gui.EntryManager.SetEntries(path)
				gui.EntryManager.SetColumns()
			}
		// go to parent dir
		case event.Rune() == 'l':
			entries := gui.EntryManager.Entries()
			if len(entries) == 0 {
				return event
			}

			row, column := gui.EntryManager.GetSelection()
			entry := entries[row-1]

			if entry.IsDir {
				path := path.Join(inputPath.GetText(), gui.EntryManager.GetCell(row, column).Text)
				gui.HistoryManager.Save(path)
				inputPath.SetText(path)
				gui.EntryManager.SetEntries(path)
				gui.EntryManager.SetColumns()
			}
		// TODO mark file or dir
		case event.Rune() == ' ':
		}

		return event
	})

	inputPath.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			gui.App.Stop()
		}
		if key == tcell.KeyEnter {
			path := inputPath.GetText()
			gui.HistoryManager.Save(path)
			gui.EntryManager.SetEntries(path)
			gui.EntryManager.SetColumns()

			gui.App.SetFocus(gui.EntryManager)
		}

	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			gui.App.SetFocus(gui.EntryManager)
		}
		return event
	})

	grid := tview.NewGrid().SetRows(1, 0)
	grid.AddItem(inputPath, 0, 0, 1, 1, 0, 0, true)
	grid.AddItem(gui.EntryManager, 1, 0, 2, 2, 0, 0, true)

	if err := gui.App.SetRoot(grid, true).SetFocus(gui.EntryManager).Run(); err != nil {
		gui.App.Stop()
		return 1, err
	}

	return 0, nil
}
