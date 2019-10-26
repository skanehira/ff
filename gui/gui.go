package gui

import (
	"os"

	"log"
	"os/exec"

	"github.com/rivo/tview"
)

// Register memory resources
type Register struct {
	MoveSources []*Entry
	CopySources []*Entry
}

// ClearMoveResources clear resources
func (r *Register) ClearMoveResources() {
	r.MoveSources = []*Entry{}
}

// ClearCopyResources clear resouces
func (r *Register) ClearCopyResources() {
	r.MoveSources = []*Entry{}
}

// Gui gui have some manager
type Gui struct {
	InputPath      *tview.InputField
	Register       *Register
	HistoryManager *HistoryManager
	EntryManager   *EntryManager
	Preview        *Preview
	App            *tview.Application
	Pages          *tview.Pages
}

func hasEntry(gui *Gui) bool {
	if len(gui.EntryManager.Entries()) != 0 {
		return true
	}
	return false
}

// New create new gui
func New() *Gui {
	return &Gui{
		EntryManager:   NewEntryManager(),
		HistoryManager: NewHistoryManager(),
		App:            tview.NewApplication(),
		Preview:        NewPreview(),
		Register:       &Register{},
	}
}

// ExecCmd execute command
func (gui *Gui) ExecCmd(attachStd bool, cmd string, args ...string) error {
	command := exec.Command(cmd, args...)

	if attachStd {
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	return command.Run()
}

// Stop stop ff
func (gui *Gui) Stop() {
	gui.App.Stop()
}

func (gui *Gui) Confirm(message, doneLabel string, primitive tview.Primitive, doneFunc func()) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{doneLabel, "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			gui.CloseAndSwitchPanel("modal", primitive)
			if buttonLabel == doneLabel {
				gui.App.QueueUpdateDraw(func() {
					doneFunc()
				})
			}
		})
	gui.Pages.AddAndSwitchToPage("modal", gui.Modal(modal, 50, 29), true).ShowPage("main")
}

func (gui *Gui) CloseAndSwitchPanel(removePrimitive string, primitive tview.Primitive) {
	gui.Pages.RemovePage(removePrimitive).ShowPage("main")
	gui.SwitchPanel(primitive)
}

func (gui *Gui) SwitchPanel(p tview.Primitive) *tview.Application {
	return gui.App.SetFocus(p)
}

func (gui *Gui) Modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
}

// Run run ff
func (gui *Gui) Run() error {
	// get current path
	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("%s: %s\n", ErrGetCwd, err)
		return err
	}

	gui.InputPath = tview.NewInputField().SetText(currentDir)

	gui.HistoryManager.Save(0, currentDir)
	gui.EntryManager.SetEntries(currentDir)

	gui.SetKeybindings()

	grid := tview.NewGrid().SetRows(1, 0).SetColumns(0, 0)
	grid.AddItem(gui.InputPath, 0, 0, 1, 2, 0, 0, true).
		AddItem(gui.EntryManager, 1, 0, 1, 1, 0, 0, true).
		AddItem(gui.Preview, 1, 1, 1, 1, 0, 0, true)

	gui.Pages = tview.NewPages().
		AddAndSwitchToPage("main", grid, true)

	if err := gui.App.SetRoot(grid, true).SetFocus(gui.EntryManager).Run(); err != nil {
		gui.App.Stop()
		return err
	}

	return nil
}
