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
