package gui

import (
	"path/filepath"

	"github.com/gdamore/tcell"
)

// globalKeybinding
func globalKeybinding(gui *Gui, event *tcell.EventKey) {
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
			gui.HistoryManager.Save(row, filepath.Join(filepath.Dir(gui.InputPath.GetText()), entry.Path))
			gui.InputPath.SetText(entry.PathName)
			gui.EntryManager.SetEntries(entry.PathName)
			gui.EntryManager.Select(1, 0)
			gui.EntryManager.SetOffset(1, 0)
			entry := gui.EntryManager.GetSelectEntry()
			gui.Preview.UpdateView(gui, entry)
		}
	}

}
