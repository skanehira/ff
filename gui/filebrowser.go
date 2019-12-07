package gui

import "github.com/rivo/tview"

type FileBrowser interface {
	tview.Primitive
	GetSearchWord() string
	SetSearchWord(word string)
	SearchFiles(gui *Gui)
	UpdateView()
	GetSelectEntry() *File
	SetEntries(path string) []*File
	ChangeDir(gui *Gui, current, target string) error
	Keybinding(gui *Gui)
}
