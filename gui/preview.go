package gui

import (
	"io/ioutil"

	"github.com/rivo/tview"
)

type Preview struct {
	*tview.TextView
}

func NewPreview() *Preview {
	p := &Preview{
		TextView: tview.NewTextView(),
	}

	p.SetBorder(true).SetTitle("preview").SetTitleAlign(tview.AlignLeft)
	return p
}

func (p *Preview) UpdateView(g *Gui, entry *Entry) error {
	var buf []byte
	var err error

	if !entry.IsDir {
		buf, err = ioutil.ReadFile(entry.Name)
		if err != nil {
			return err
		}
	}

	g.App.QueueUpdateDraw(func() {
		p.SetText(string(buf))
	})

	return nil
}
