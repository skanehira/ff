package gui

import "github.com/rivo/tview"

type Tree struct {
	*tview.TreeView
}

func NewTree() *Tree {
	return &Tree{
		TreeView: tview.NewTreeView(),
	}
}

func (t *Tree) GetSearchWord() string {
	panic("not implemented")
}

func (t *Tree) SetSearchWord(word string) {
	panic("not implemented")
}

func (t *Tree) SearchFiles(gui *Gui) {
	panic("not implemented")
}

func (t *Tree) UpdateView() {
	panic("not implemented")
}

func (t *Tree) GetSelectEntry() *File {
	panic("not implemented")
}

func (t *Tree) ChangeDir(gui *Gui, current string, target string) error {
	panic("not implemented")
}

func (t *Tree) Keybinding(gui *Gui) {
	panic("not implemented")
}

func (t *Tree) SetEntries(path string) []*File {
	panic("not implemented")
}
