package gui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var (
	helpHeaders     = []string{"KEY", "DESCRIPTION"}
	helpKeybindings = []map[string]string{
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
		{"c or :": "focus cmdline"},
		{".": "edit config.yaml"},
		{"b": "bookmark directory"},
		{"B": "open bookmarks panel"},
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

	h.UpdateView()
	return h
}

func (h *Help) UpdateView() {
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

	for i, keybind := range helpKeybindings {
		for k, d := range keybind {
			table.SetCell(i+1, 0, tview.NewTableCell(k).SetTextColor(tcell.ColorWhite))
			table.SetCell(i+1, 1, tview.NewTableCell(d).SetTextColor(tcell.ColorWhite))
		}
	}
}
