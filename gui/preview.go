package gui

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
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
	p.SetDynamicColors(true)
	return p
}

func (p *Preview) UpdateView(g *Gui, entry *Entry) {
	var text string
	if !entry.IsDir {
		text = p.Highlight(entry)
	}
	g.App.QueueUpdateDraw(func() {
		p.SetText(text).ScrollToBeginning()
	})
}

func (p *Preview) Highlight(entry *Entry) string {
	// Determine lexer.
	b, err := ioutil.ReadFile(entry.PathName)
	if err != nil {
		return err.Error()
	}

	ext := filepath.Ext(entry.Name)
	l := lexers.Get(ext)
	if l == nil {
		l = lexers.Analyse(string(b))
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	// Determine formatter.
	f := formatters.Get("terminal256")
	if f == nil {
		f = formatters.Fallback
	}

	// Determine style.
	s := styles.Get("monokai")
	if s == nil {
		s = styles.Fallback
	}

	it, err := l.Tokenise(nil, string(b))
	if err != nil {
		return err.Error()
	}

	var buf = bytes.Buffer{}

	if err := f.Format(&buf, s, it); err != nil {
		return err.Error()
	}

	return tview.TranslateANSI(buf.String())
}
