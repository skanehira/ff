package gui

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell"
)

var (
	ErrNoDirName       = errors.New("no directory name")
	ErrNoFileName      = errors.New("no file name")
	ErrNoFileOrDirName = errors.New("no file or directory name")
	ErrNoFileOrDir     = errors.New("no file or directory")
	ErrNoNewName       = errors.New("no new name")
)

func (gui *Gui) SetKeybindings() {
	gui.FileBrowser.Keybinding(gui)
	gui.InputPathKeybinding()
	gui.CmdLineKeybinding()
	gui.HelpKeybinding()
	if gui.Config.Bookmark.Enable {
		gui.Bookmark.BookmarkKeybinding(gui)
	}
}

func (gui *Gui) ChangeDir(current, target string) error {
	if gui.Config.Bookmark.Enable {
		gui.Bookmark.SetSearchWord("")
	}
	gui.InputPath.SetText(target)

	return gui.FileBrowser.ChangeDir(gui, current, target)
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
				if err := gui.ChangeDir(parent, path); err != nil {
					gui.Message(err.Error(), FilesPanel)
					return
				}
				gui.FocusPanel(FilesPanel)
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

func (gui *Gui) CmdLineKeybinding() {
	cmdline := gui.CmdLine

	cmdline.SetDoneFunc(func(key tcell.Key) {
		text := cmdline.GetText()
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
			cmdline.SetText("")
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

func (gui *Gui) HelpKeybinding() {
	gui.Help.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			gui.Pages.RemovePage("help")
			gui.FocusPanel(gui.CurrentPanel)
		}
		return event
	})
}
