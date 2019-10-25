package gui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/gdamore/tcell"
	"github.com/mitchellh/go-homedir"
	"github.com/rivo/tview"
)

// Register memory resources
type Register struct {
	MoveSources []*Entry
	CopySources []*Entry
}

// ClearMoveResources clear resources
func (r *Register) ClearMoveResources() {
	r.MoveSources = []*Entry{}
}

// ClearCopyResources clear resouces
func (r *Register) ClearCopyResources() {
	r.MoveSources = []*Entry{}
}

// Gui gui have some manager
type Gui struct {
	InputPath      *tview.InputField
	Register       *Register
	HistoryManager *HistoryManager
	EntryManager   *EntryManager
	Preview        *Preview
	App            *tview.Application
}

func hasEntry(gui *Gui) bool {
	if len(gui.EntryManager.Entries()) != 0 {
		return true
	}
	return false
}

func initLogger() error {
	// init logger
	home, err := homedir.Dir()
	if err != nil {
		return errors.New(fmt.Sprintf("cannot get home dir, cause: %s", err))
	}
	logFile, err := os.OpenFile(filepath.Join(home, "ff.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return errors.New(fmt.Sprintf("cannot open log file, cause: %s", err))
	}

	log.SetOutput(logFile)

	return nil
}

// New create new gui
func New() (*Gui, error) {
	if err := initLogger(); err != nil {
		return nil, err
	}

	return &Gui{
		EntryManager:   NewEntryManager(),
		HistoryManager: NewHistoryManager(),
		App:            tview.NewApplication(),
		Preview:        NewPreview(),
		Register:       &Register{},
	}, nil
}

// ExecCmd execute command
func (gui *Gui) ExecCmd(attachStd bool, cmd string, args ...string) error {
	command := exec.Command(cmd, args...)

	if attachStd {
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	return command.Run()
}

// Stop stop ff
func (gui *Gui) Stop() {
	gui.App.Stop()
}

// Run run ff
func (gui *Gui) Run() (int, error) {
	// get current path
	currentDir, err := os.Getwd()
	if err != nil {
		return 1, err
	}

	gui.InputPath = tview.NewInputField().SetText(currentDir)

	gui.HistoryManager.Save(0, currentDir)
	gui.EntryManager.SetEntries(currentDir)

	gui.EntryManager.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			gui.App.Stop()
		}

	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		globalKeybinding(gui, event)
		switch {
		// cut entry
		case event.Rune() == 'd':
			if !hasEntry(gui) {
				return event
			}
			//gui.EntryManager.SetViewable(false)
			gui.Register.MoveSources = append(gui.Register.MoveSources, gui.EntryManager.GetSelectEntry())

		// copy entry
		case event.Rune() == 'y':
			if !hasEntry(gui) {
				return event
			}
			gui.Register.CopySources = append(gui.Register.CopySources, gui.EntryManager.GetSelectEntry())

		// paset entry
		case event.Rune() == 'p':
			//for _, source := range gui.Register.MoveSources {
			//	dest := filepath.Join(gui.InputPath.GetText(), filepath.Base(source))
			//	if err := os.Rename(source, dest); err != nil {
			//		log.Printf("cannot copy or move the file: %s", err)
			//	}
			//}

			// TODO implement file copy
			//for _, source := range gui.Register.CopyResources {
			//dest := filepath.Join(gui.InputPath.GetText(), filepath.Base(source))
			//}

			//gui.EntryManager.SetEntries(gui.InputPath.GetText())

		// edit file with $EDITOR
		case event.Rune() == 'e':
			editor := os.Getenv("EDITOR")
			if editor == "" {
				log.Println("please set your editor to $EDITOR")
				return event
			}

			entry := gui.EntryManager.GetSelectEntry()

			gui.App.Suspend(func() {
				if err := gui.ExecCmd(true, editor, entry.PathName); err != nil {
					log.Printf("cannot edit: %s", err)
				}
			})
		case event.Rune() == 'q':
			gui.Stop()
		}

		return event
	})

	gui.EntryManager.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 {
			f := gui.EntryManager.Entries()[row-1]
			gui.Preview.UpdateView(gui, f)
		}
	})

	gui.InputPath.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			gui.App.Stop()
		}

		if key == tcell.KeyEnter {
			path := gui.InputPath.GetText()
			row, _ := gui.EntryManager.GetSelection()
			gui.HistoryManager.Save(row, path)
			gui.EntryManager.SetEntries(path)

			gui.App.SetFocus(gui.EntryManager)
		}

	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			gui.App.SetFocus(gui.EntryManager)
		}

		return event
	})

	grid := tview.NewGrid().SetRows(1, 0).SetColumns(0, 0)
	grid.AddItem(gui.InputPath, 0, 0, 1, 2, 0, 0, true).
		AddItem(gui.EntryManager, 1, 0, 1, 1, 0, 0, true).
		AddItem(gui.Preview, 1, 1, 1, 1, 0, 0, true)

	if err := gui.App.SetRoot(grid, true).SetFocus(gui.EntryManager).Run(); err != nil {
		gui.App.Stop()
		return 1, err
	}

	return 0, nil
}
