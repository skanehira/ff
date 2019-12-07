package gui

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type CmdLine struct {
	*tview.InputField
}

func NewCmdLine() *CmdLine {
	c := &CmdLine{
		InputField: tview.NewInputField(),
	}

	c.SetLabel("cmd:")
	c.SetFieldBackgroundColor(tcell.ColorBlack)
	return c
}

func (c *CmdLine) Keybinding(gui *Gui) {
	c.SetDoneFunc(func(key tcell.Key) {
		text := c.GetText()
		if text == "" {
			return
		}

		cmdText := strings.Split(text, " ")

		// expand environments
		for i, c := range cmdText[1:] {
			cmdText[i+1] = os.ExpandEnv(c)
		}

		cmd := exec.Command(cmdText[0], cmdText[1:]...)

		buf := bytes.Buffer{}
		cmd.Stderr = &buf
		cmd.Stdout = &buf
		if err := cmd.Run(); err == nil {
			c.SetText("")
		}

		result := strings.TrimRight(buf.String(), "\n")
		if result != "" {
			gui.Message(result, CmdLinePanel)
		}

		gui.FileBrowser.SetEntries(gui.InputPath.GetText())
	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyEsc:
			gui.App.SetFocus(gui.FileBrowser)
			return event
		case tcell.KeyF1:
			gui.CurrentPanel = CmdLinePanel
			gui.Help.UpdateView(gui.CurrentPanel)
			gui.Pages.AddAndSwitchToPage("help", gui.Modal(gui.Help, 0, 0), true).ShowPage("main")
		}

		return event
	})
}
