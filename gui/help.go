package gui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var (
	helpHeaders    = []string{"KEY", "DESCRIPTION"}
	fileTableHelps = []map[string]string{
		{"tab": "focus to files"},
		{"j": "move next"},
		{"k": "move previous"},
		{"g": "move to top"},
		{"G": "move to bottom"},
		{"ctrl-b": "move previous page"},
		{"ctrl-f": "move next page"},
		{"h": "move to parent path"},
		{"l": "move to specified path"},
		{"y": "copy selected file or directory"},
		{"p": "paste file or directory"},
		{"d": "delete selected file or directory"},
		{"m": "make a new directory"},
		{"n": "make a new file"},
		{"r": "rename a directory or file"},
		{"e": "edit file with $EDITOR"},
		{"o": "open file or dierectory"},
		{"f or /": "search files or directories"},
		{"ctrl-j": "scroll preview panel down"},
		{"ctrl-k": "scroll preview panel up"},
		{".": "edit config.yaml"},
		{"b": "bookmark directory"},
		{"B": "open bookmarks panel"},
	}

	fileTreeHelps = []map[string]string{
		{"tab": "focus to files"},
		{"j": "move next"},
		{"k": "move previous"},
		{"g": "move to top"},
		{"G": "move to bottom"},
		{"h": "collapse specified path"},
		{"l": "expand specified path"},
		{"H": "move to parent path"},
		{"L": "move to specified path"},
		{"y": "copy selected file or directory"},
		{"p": "paste file or directory"},
		{"d": "delete selected file or directory"},
		{"m": "make a new directory"},
		{"n": "make a new file"},
		{"r": "rename a directory or file"},
		{"e": "edit file with $EDITOR"},
		{"o": "open file or dierectory"},
		{"f or /": "search files or directories"},
		{"ctrl-j": "scroll preview panel down"},
		{"ctrl-k": "scroll preview panel up"},
		{".": "edit config.yaml"},
		{"b": "bookmark directory"},
		{"B": "open bookmarks panel"},
	}

	pathHelps = []map[string]string{
		{"enter": "change directory"},
	}

	bookmarkHelps = []map[string]string{
		{"a": "add bookmark"},
		{"d": "delete bookmark"},
		{"q": "close bookmarks panel"},
		{"ctrl-g": "go to bookmark"},
		{"f or /": "search bookmarks"},
	}
)

type Help struct {
	*tview.Table
}

func NewHelp() *Help {
	h := &Help{
		Table: tview.NewTable().SetSelectable(true, false).SetFixed(1, 1),
	}

	h.SetBorder(true).SetTitle("help").
		SetTitleAlign(tview.AlignLeft)

	return h
}

func (h *Help) UpdateView(panel Panel) {
	table := h.Table.Clear()

	for i, h := range helpHeaders {
		table.SetCell(0, i, &tview.TableCell{
			Text:            h,
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorYellow,
			BackgroundColor: tcell.ColorDefault,
		})
	}

	var keybindings []map[string]string

	switch panel {
	case PathPanel:
		keybindings = pathHelps
	case FileTablePanel:
		keybindings = fileTableHelps
	case FileTreePanel:
		keybindings = fileTreeHelps
	case BookmarkPanel:
		keybindings = bookmarkHelps
	}

	for i, keybind := range keybindings {
		for k, d := range keybind {
			table.SetCell(i+1, 0, tview.NewTableCell(k).SetTextColor(tcell.ColorWhite))
			table.SetCell(i+1, 1, tview.NewTableCell(d).SetTextColor(tcell.ColorWhite))
		}
	}
}

func (h *Help) Keybinding(gui *Gui) {
	h.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'l':
			return nil
		case 'q':
			gui.Pages.RemovePage("help")
			gui.FocusPanel(gui.CurrentPanel)
		}
		return event
	})
}
