package gui

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell"
	"github.com/skanehira/ff/system"
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
		if !hasEntry(gui) {
			return
		}

		entry := gui.EntryManager.GetSelectEntry()
		row, _ := gui.EntryManager.GetSelection()

		if entry.IsDir {
			if len(gui.EntryManager.SetEntries(entry.PathName)) == 0 {
				return
			}
			gui.HistoryManager.Save(row, filepath.Join(filepath.Dir(gui.InputPath.GetText()), entry.Path))
			gui.InputPath.SetText(entry.PathName)
			gui.EntryManager.Select(1, 0)
			gui.EntryManager.SetOffset(1, 0)
			entry := gui.EntryManager.GetSelectEntry()
			gui.Preview.UpdateView(gui, entry)
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
		switch {
		// cut entry
		case event.Rune() == 'd':
			if !hasEntry(gui) {
				return event
			}

			gui.Confirm("do you want to remove this?", "yes", gui.EntryManager, func() error {
				entry := gui.EntryManager.GetSelectEntry()
				if entry == nil {
					return nil
				}

				if err := gui.RemoveFile(entry); err != nil {
					return err
				}

				path := gui.InputPath.GetText()
				gui.EntryManager.SetEntries(path)
				return nil
			})

		// copy entry
		case event.Rune() == 'y':
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
		case event.Rune() == 'p':
			source := gui.Register.CopySource
			target := filepath.Join(gui.InputPath.GetText(), source.Name)

			if err := system.CopyFile(source.PathName, target); err != nil {
				gui.Message(err.Error(), gui.EntryManager)
				return event
			}

			gui.EntryManager.SetEntries(gui.InputPath.GetText())

		// edit file with $EDITOR
		case event.Rune() == 'e':
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

}

func (gui *Gui) RemoveFile(entry *Entry) error {
	if entry.IsDir {
		return nil
	}

	_, err := os.Stat(entry.PathName)
	if os.IsNotExist(err) {
		log.Println(err)
		return err
	}

	if err := os.Remove(entry.PathName); err != nil {
		log.Println(err)
		return err
	}
	return nil
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
