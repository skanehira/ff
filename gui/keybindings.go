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
		row, _ := gui.EntryManager.GetSelection()
		gui.HistoryManager.Save(row, path)
		if path != "" {
			gui.InputPath.SetText(path)
			gui.EntryManager.SetEntries(path)
			gui.EntryManager.Select(0, 0)
			gui.EntryManager.SetOffset(0, 0)
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
			gui.InputPath.SetText(entry.Path)
			gui.EntryManager.SetEntries(entry.Path)
			gui.EntryManager.Select(0, 0)
			gui.EntryManager.SetOffset(0, 0)
		}
	}
}
