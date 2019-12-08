package gui

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/skanehira/ff/system"
)

var (
	ErrNoDirName       = errors.New("no directory name")
	ErrNoFileName      = errors.New("no file name")
	ErrNoFileOrDirName = errors.New("no file or directory name")
	ErrNoFileOrDir     = errors.New("no file or directory")
	ErrNoNewName       = errors.New("no new name")
)

func (gui *Gui) commonFileBrowserKeybinding(event *tcell.EventKey) {
	if gui.Config.Preview.Enable {
		switch event.Key() {
		case tcell.KeyCtrlJ:
			gui.Preview.ScrollDown()
		case tcell.KeyCtrlK:
			gui.Preview.ScrollUp()
		}
	}

	switch event.Key() {
	case tcell.KeyTab:
		gui.App.SetFocus(gui.InputPath)
	}

	switch event.Rune() {
	case 'e':
		entry := gui.FileBrowser.GetSelectEntry()
		if entry == nil {
			log.Println("cannot get entry")
			return
		}

		if err := gui.EditFile(entry.PathName); err != nil {
			gui.Message(err.Error(), FileTablePanel)
		}

	case 'q':
		gui.Stop()

	case 'o':
		entry := gui.FileBrowser.GetSelectEntry()
		if entry == nil {
			return
		}
		if err := system.Open(entry.PathName); err != nil {
			gui.Message(err.Error(), FileTablePanel)
		}

	case ':', 'c':
		gui.FocusPanel(CmdLinePanel)

	case '.':
		if err := gui.EditFile(gui.Config.ConfigFile); err != nil {
			gui.Message(err.Error(), FileTablePanel)
		}

	case 'b':
		if gui.Config.Bookmark.Enable {
			entry := gui.FileBrowser.GetSelectEntry()
			if entry != nil && entry.IsDir {
				if err := gui.Bookmark.Add(entry.PathName); err != nil {
					gui.Message(err.Error(), FileTablePanel)
				}
			}
		}

	case 'B':
		if gui.Config.Bookmark.Enable {
			if err := gui.Bookmark.Update(); err != nil {
				gui.Message(err.Error(), FileTablePanel)
				return
			}
			gui.CurrentPanel = BookmarkPanel
			gui.Pages.AddAndSwitchToPage("bookmark", gui.Bookmark, true).ShowPage("main")
		}

	}
}

func (gui *Gui) SetKeybindings() {
	gui.FileBrowser.Keybinding(gui)
	gui.InputPathKeybinding()
	gui.CmdLine.Keybinding(gui)
	gui.Help.Keybinding(gui)

	if gui.Config.Bookmark.Enable {
		gui.Bookmark.BookmarkKeybinding(gui)
	}
}

func (gui *Gui) EditFile(file string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return ErrNoEditor
	}

	// if `ff` running in vim terminal, use running vim
	if os.Getenv("VIM_TERMINAL") != "" && editor == "vim" {
		cmd := exec.Command("sh", "-c", fmt.Sprintf(`echo -e '\x1b]51;["drop","%s"]\x07'`, file))
		cmd.Stdout = os.Stdout
		return cmd.Run()
	}

	gui.App.Suspend(func() {
		if err := gui.ExecCmd(true, editor, file); err != nil {
			log.Printf("%s: %s\n", ErrEdit, err)
		}
	})

	if gui.Config.Preview.Enable {
		entry := gui.FileBrowser.GetSelectEntry()
		gui.Preview.UpdateView(gui, entry)
	}

	return nil
}

func (gui *Gui) InputPathKeybinding() {
	gui.InputPath.SetAutocompleteFunc(func(text string) []string {
		var entries []string

		dir := filepath.Dir(text)
		i, err := os.Lstat(dir)
		if err != nil || !i.IsDir() {
			log.Println(err)
			return entries
		}

		var fileName string
		if !strings.HasSuffix(text, "/") {
			fileName = filepath.Base(text)
		}

		parent, _ := filepath.Split(text)

		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Println(err)
			return entries
		}

		for _, f := range files {
			target := f.Name()
			if gui.Config.IgnoreCase {
				target = strings.ToLower(f.Name())
				fileName = strings.ToLower(fileName)
			}
			if f.IsDir() && strings.Index(target, fileName) != -1 {
				entries = append(entries, filepath.Join(parent, f.Name()))
			}
		}

		return entries
	})

	gui.InputPath.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			path := os.ExpandEnv(gui.InputPath.GetText())
			file, err := os.Lstat(path)
			if err != nil {
				log.Println(err)
				return
			}

			parent := filepath.Dir(path)
			if parent != "" && file.IsDir() {
				if err := gui.FileBrowser.ChangeDir(gui, parent, path); err != nil {
					gui.Message(err.Error(), FileTablePanel)
					return
				}
				gui.FocusPanel(FileTablePanel)
			}
		}
	})

	gui.InputPath.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			gui.CurrentPanel = PathPanel
			gui.Help.UpdateView(gui.CurrentPanel)
			gui.Pages.AddAndSwitchToPage("help", gui.Modal(gui.Help, 0, 0), true).ShowPage("main")

		}
		return event
	})
}
