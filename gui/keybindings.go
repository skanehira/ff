package gui

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell"
	"github.com/skanehira/ff/system"
)

var (
	ErrNoDirName  = errors.New("no directory name")
	ErrNoFileName = errors.New("no file name")
)

func (gui *Gui) SetKeybindings() {
	gui.InputPathKeybinding()
	gui.EntryManagerKeybinding()
}

// globalKeybinding
func (gui *Gui) GlobalKeybinding(event *tcell.EventKey) {
	switch {
	// go to input view
	case event.Key() == tcell.KeyTab:
		gui.App.SetFocus(gui.InputPath)

	// go to previous history
	case event.Key() == tcell.KeyCtrlH:
		history := gui.HistoryManager.Previous()
		if history != nil {
			gui.InputPath.SetText(history.Path)
			gui.EntryManager.SetEntries(history.Path)
			gui.EntryManager.Select(history.RowIdx, 0)
		}

	// go to next history
	case event.Key() == tcell.KeyCtrlL:
		history := gui.HistoryManager.Next()
		if history != nil {
			gui.InputPath.SetText(history.Path)
			gui.EntryManager.SetEntries(history.Path)
			gui.EntryManager.Select(history.RowIdx, 0)
		}

	// go to previous dir
	case event.Rune() == 'h':
		path := filepath.Dir(gui.InputPath.GetText())
		entry := gui.EntryManager.GetSelectEntry()
		if path != "" {
			gui.InputPath.SetText(path)
			gui.EntryManager.SetEntries(path)
			gui.EntryManager.Select(1, 0)
			gui.EntryManager.SetOffset(0, 0)
			entry = gui.EntryManager.GetSelectEntry()
			gui.Preview.UpdateView(gui, entry)
		}

	// go to parent dir
	case event.Rune() == 'l':
		entry := gui.EntryManager.GetSelectEntry()
		if entry != nil {
			row, _ := gui.EntryManager.GetSelection()

			if entry.IsDir {
				gui.EntryManager.SetEntries(entry.PathName)
				gui.HistoryManager.Save(row, filepath.Join(filepath.Dir(gui.InputPath.GetText()), entry.Path))
				gui.InputPath.SetText(entry.PathName)
				gui.EntryManager.Select(1, 0)
				gui.EntryManager.SetOffset(1, 0)
				entry := gui.EntryManager.GetSelectEntry()
				gui.Preview.UpdateView(gui, entry)
			}
		}
	}
}

func (gui *Gui) EntryManagerKeybinding() {
	gui.EntryManager.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			gui.App.Stop()
		}

	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		gui.GlobalKeybinding(event)
		switch event.Rune() {
		// cut entry
		case 'd':
			if !hasEntry(gui) {
				return event
			}

			gui.Confirm("do you want to remove this?", "yes", gui.EntryManager, func() error {
				entry := gui.EntryManager.GetSelectEntry()
				if entry == nil {
					return nil
				}

				if entry.IsDir {
					if err := system.RemoveDirAll(entry.PathName); err != nil {
						log.Println(err)
						return err
					}
				} else {
					if err := system.RemoveFile(entry.PathName); err != nil {
						log.Println(err)
						return err
					}
				}

				path := gui.InputPath.GetText()
				gui.EntryManager.SetEntries(path)
				return nil
			})

		// copy entry
		case 'y':
			if !hasEntry(gui) {
				return event
			}

			m := gui.EntryManager
			m.UpdateColor()
			entry := m.GetSelectEntry()
			gui.Register.CopySource = entry

			row, _ := m.GetSelection()
			for i := 0; i < 5; i++ {
				m.GetCell(row, i).SetTextColor(tcell.ColorYellow)
			}

		// paste entry
		case 'p':
			source := gui.Register.CopySource

			gui.Form(map[string]string{"name": source.Name}, "paste", "new name", "new_name", gui.EntryManager,
				7, func(values map[string]string) error {
					name := values["name"]
					if name == "" {
						return ErrNoFileName
					}

					target := filepath.Join(gui.InputPath.GetText(), name)
					if err := system.CopyFile(source.PathName, target); err != nil {
						log.Println(err)
						return err
					}

					gui.EntryManager.SetEntries(gui.InputPath.GetText())
					return nil
				})

		// edit file with $EDITOR
		case 'e':
			editor := os.Getenv("EDITOR")
			if editor == "" {
				log.Println("$EDITOR is empty, please set $EDITOR")
				return event
			}

			entry := gui.EntryManager.GetSelectEntry()
			if entry == nil {
				log.Println("cannot get entry")
				return event
			}

			gui.App.Suspend(func() {
				if err := gui.ExecCmd(true, editor, entry.PathName); err != nil {
					log.Printf("%s: %s\n", ErrEdit, err)
				}
			})
		case 'm':
			gui.Form(map[string]string{"name": ""}, "create", "new direcotry",
				"create_directory", gui.EntryManager,
				7, func(values map[string]string) error {
					name := values["name"]
					if name == "" {
						return ErrNoDirName
					}

					target := filepath.Join(gui.InputPath.GetText(), name)
					if err := system.NewDir(target); err != nil {
						log.Println(err)
						return err
					}

					gui.EntryManager.SetEntries(gui.InputPath.GetText())
					return nil
				})
		case 'n':
			gui.Form(map[string]string{"name": ""}, "create", "new file", "create_directory", gui.EntryManager,
				7, func(values map[string]string) error {
					name := values["name"]
					if name == "" {
						return ErrNoFileName
					}

					target := filepath.Join(gui.InputPath.GetText(), name)
					if err := system.NewFile(target); err != nil {
						log.Println(err)
						return err
					}

					gui.EntryManager.SetEntries(gui.InputPath.GetText())
					return nil
				})
		case 'q':
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

}

func (gui *Gui) InputPathKeybinding() {
	gui.InputPath.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			gui.App.Stop()
		}

		if key == tcell.KeyEnter {
			path := gui.InputPath.GetText()
			path = os.ExpandEnv(path)
			gui.InputPath.SetText(path)
			row, _ := gui.EntryManager.GetSelection()
			gui.HistoryManager.Save(row, path)
			gui.EntryManager.SetEntries(path)
		}

	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			gui.App.SetFocus(gui.EntryManager)
		}

		return event
	})
}
