package gui

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/gdamore/tcell"
	"github.com/mitchellh/go-homedir"
	"github.com/rivo/tview"
)

// Register memory resources
type Register struct {
	MoveSources []string
	CopySources []string
}

// ClearMoveResources clear resources
func (r *Register) ClearMoveResources() {
	r.MoveSources = []string{}
}

// ClearCopyResources clear resouces
func (r *Register) ClearCopyResources() {
	r.MoveSources = []string{}
}

// Gui gui have some manager
type Gui struct {
	InputPath      *tview.InputField
	Register       *Register
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
		Register:       &Register{},
	}
}

// ExecCmd exec specified command
func (gui *Gui) ExecCmd(attachStd bool, cmd string, args ...string) error {
	command := exec.Command(cmd, args...)

	if attachStd {
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	return command.Run()
}

// Run run ff
func (gui *Gui) Run() (int, error) {
	// get current path
	currentDir, err := os.Getwd()
	if err != nil {
		return 1, err
	}

	gui.HistoryManager.Save(currentDir)

	gui.InputPath = tview.NewInputField().SetText(currentDir)

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
			gui.App.SetFocus(gui.InputPath)
		// go to prev history
		case event.Key() == tcell.KeyCtrlH:
			path := gui.HistoryManager.Previous()
			if path != "" {
				gui.InputPath.SetText(path)
				gui.EntryManager.SetEntries(path)
				gui.EntryManager.SetColumns()
			}
		// go to next history
		case event.Key() == tcell.KeyCtrlL:
			path := gui.HistoryManager.Next()
			if path != "" {
				gui.InputPath.SetText(path)
				gui.EntryManager.SetEntries(path)
				gui.EntryManager.SetColumns()
			}
		// go to specified dir
		// TODO save position info
		case event.Rune() == 'h':
			path := filepath.Dir(gui.InputPath.GetText())
			if path != "" {
				gui.InputPath.SetText(path)
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
				path := path.Join(gui.InputPath.GetText(), gui.EntryManager.GetCell(row, column).Text)
				gui.HistoryManager.Save(path)
				gui.InputPath.SetText(path)
				gui.EntryManager.SetEntries(path)
				gui.EntryManager.SetColumns()
			}
		// cut entry
		case event.Rune() == 'd':
			source := filepath.Join(gui.InputPath.GetText(), gui.EntryManager.GetCell(gui.EntryManager.GetSelection()).Text)
			gui.Register.MoveSources = append(gui.Register.MoveSources, source)
		case event.Rune() == 'y':
			source := filepath.Join(gui.InputPath.GetText(), gui.EntryManager.GetCell(gui.EntryManager.GetSelection()).Text)
			gui.Register.CopySources = append(gui.Register.CopySources, source)
		case event.Rune() == 'p':
			for _, source := range gui.Register.MoveSources {
				dest := filepath.Join(gui.InputPath.GetText(), filepath.Base(source))
				if err := os.Rename(source, dest); err != nil {
					log.Printf("cannot copy or move the file: %s", err)
				}
			}

			// TODO implement file copy
			//for _, source := range gui.Register.CopyResources {
			//dest := filepath.Join(gui.InputPath.GetText(), filepath.Base(source))
			//}

			gui.EntryManager.SetEntries(gui.InputPath.GetText())
			gui.EntryManager.SetColumns()
		// edit file with $EDITOR
		case event.Rune() == 'e':
			editor := os.Getenv("EDITOR")
			if editor == "" {
				log.Println("please set your editor to $EDITOR")
				return event
			}

			entry := gui.EntryManager.GetCell(gui.EntryManager.GetSelection()).Text

			gui.App.Suspend(func() {
				if err := gui.ExecCmd(true, editor, entry); err != nil {
					log.Printf("cannot edit: %s", err)
				}
			})
		}

		return event
	})

	gui.InputPath.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			gui.App.Stop()
		}
		if key == tcell.KeyEnter {
			path := gui.InputPath.GetText()
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
	grid.AddItem(gui.InputPath, 0, 0, 1, 1, 0, 0, true)
	grid.AddItem(gui.EntryManager, 1, 0, 2, 2, 0, 0, true)

	if err := gui.App.SetRoot(grid, true).SetFocus(gui.EntryManager).Run(); err != nil {
		gui.App.Stop()
		return 1, err
	}

	return 0, nil
}
