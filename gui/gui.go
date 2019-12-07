package gui

import (
	"context"
	"os"
	"sync"
	"time"

	"log"
	"os/exec"

	"github.com/rivo/tview"
)

var (
	searchFiles     *tview.InputField
	searchBookmarks *tview.InputField
)

type Panel int

const (
	PathPanel Panel = iota + 1
	FilesPanel
	CmdLinePanel
	BookmarkPanel
)

// Register copy/paste file resource
type Register struct {
	MoveSources []*Entry
	CopySources []*Entry
	CopySource  *Entry
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
	CurrentPanel   Panel
	Config         Config
	InputPath      *tview.InputField
	Register       *Register
	HistoryManager *HistoryManager
	FileBrowser    FileBrowser
	Preview        *Preview
	CmdLine        *CmdLine
	Bookmark       *Bookmarks
	Help           *Help
	App            *tview.Application
	Pages          *tview.Pages
	wg             *sync.WaitGroup
	ctxCancel      context.CancelFunc
}

// New create new gui
func New(config Config) *Gui {
	gui := &Gui{
		Config:         config,
		InputPath:      tview.NewInputField().SetLabel("path").SetLabelWidth(5),
		HistoryManager: NewHistoryManager(),
		CmdLine:        NewCmdLine(),
		Help:           NewHelp(),
		App:            tview.NewApplication(),
		Register:       &Register{},
		Pages:          tview.NewPages(),
		wg:             &sync.WaitGroup{},
	}

	if gui.Config.Preview.Enable {
		gui.Preview = NewPreview(config.Preview.Colorscheme)
	}

	if gui.Config.EnableTree {
		gui.FileBrowser = NewTree()
	} else {
		gui.FileBrowser = NewFileTable(config.IgnoreCase)
	}

	if gui.Config.Bookmark.Enable {
		bookmark, err := NewBookmark(config)
		if err != nil {
			gui.Config.Bookmark.Enable = false
		}
		gui.Bookmark = bookmark
	}

	return gui
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
	gui.ctxCancel()
	gui.wg.Wait()
	gui.App.Stop()
}

func (gui *Gui) Message(message string, panel Panel) {
	doneLabel := "ok"
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{doneLabel}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			gui.Pages.RemovePage("message")
			gui.FocusPanel(panel)
		})

	gui.Pages.AddAndSwitchToPage("message", gui.Modal(modal, 80, 29), true).ShowPage("main")
}

func (gui *Gui) Confirm(message, doneLabel string, panel Panel, doneFunc func() error) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{doneLabel, "cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			gui.Pages.RemovePage("message").SwitchToPage("main")

			if buttonLabel == doneLabel {
				gui.App.QueueUpdateDraw(func() {
					if err := doneFunc(); err != nil {
						log.Println(err)
						gui.Message(err.Error(), panel)
					} else {
						gui.FocusPanel(panel)
					}
				})
			}
			gui.FocusPanel(panel)
		})
	gui.Pages.AddAndSwitchToPage("confirm", gui.Modal(modal, 50, 29), true).ShowPage("main")
}

func (gui *Gui) Modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
}

func (gui *Gui) FocusPanel(panel Panel) {
	var p tview.Primitive
	switch panel {
	case PathPanel:
		p = gui.InputPath
	case FilesPanel:
		p = gui.FileBrowser
	case CmdLinePanel:
		p = gui.CmdLine
	case BookmarkPanel:
		p = gui.Bookmark
	}

	gui.CurrentPanel = panel
	gui.App.SetFocus(p)
}

func (gui *Gui) Form(fieldLabel map[string]string, doneLabel, title, pageName string, panel Panel,
	height int, doneFunc func(values map[string]string) error) {

	form := tview.NewForm()
	for k, v := range fieldLabel {
		form.AddInputField(k, v, 0, nil, nil)
	}

	form.AddButton(doneLabel, func() {
		values := make(map[string]string)

		for label, _ := range fieldLabel {
			item := form.GetFormItemByLabel(label)
			switch item.(type) {
			case *tview.InputField:
				input, ok := item.(*tview.InputField)
				if ok {
					values[label] = os.ExpandEnv(input.GetText())
				}
			}
		}

		if err := doneFunc(values); err != nil {
			log.Println(err)
			gui.Message(err.Error(), panel)
			return
		}

		defer gui.FocusPanel(panel)
		defer gui.Pages.RemovePage(pageName)
	}).
		AddButton("cancel", func() {
			gui.Pages.RemovePage(pageName)
			gui.FocusPanel(panel)
		})

	form.SetBorder(true).SetTitle(title).
		SetTitleAlign(tview.AlignLeft)

	gui.Pages.AddAndSwitchToPage(pageName, gui.Modal(form, 0, height), true).ShowPage("main")
}

// Run run ff
func (gui *Gui) Run() error {
	// get current path
	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("%s: %s\n", ErrGetCwd, err)
		return err
	}

	gui.ChangeDir(currentDir, currentDir)

	grid := tview.NewGrid().SetRows(1, 0, 1).
		AddItem(gui.InputPath, 0, 0, 1, 2, 0, 0, true).
		AddItem(gui.CmdLine, 2, 0, 1, 2, 0, 0, true)

	if gui.Config.Preview.Enable {
		grid.SetColumns(0, 0).
			AddItem(gui.FileBrowser, 1, 0, 1, 1, 0, 0, true).
			AddItem(gui.Preview, 1, 1, 1, 1, 0, 0, true)

		gui.Preview.UpdateView(gui, gui.FileBrowser.GetSelectEntry())
	} else {
		grid.AddItem(gui.FileBrowser, 1, 0, 1, 2, 0, 0, true)
	}

	gui.CurrentPanel = FilesPanel
	gui.SetKeybindings()
	gui.Pages.AddAndSwitchToPage("main", grid, true)

	ctx, cancel := context.WithCancel(context.Background())
	gui.ctxCancel = cancel

	gui.wg.Add(1)
	go func(ctx context.Context) {
		t := time.NewTicker(5 * time.Second)
		defer func() {
			t.Stop()
			gui.wg.Done()
		}()

		for {
			select {
			case <-t.C:
				gui.App.QueueUpdateDraw(func() {
					gui.FileBrowser.UpdateView()
				})
			case <-ctx.Done():
				return
			}
		}

	}(ctx)

	if err := gui.App.SetRoot(gui.Pages, true).SetFocus(gui.FileBrowser).Run(); err != nil {
		gui.Stop()
		return err
	}

	return nil
}
