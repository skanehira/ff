package gui

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/skanehira/ff/system"
)

type Tree struct {
	files      []*File
	ignorecase bool
	searchWord string
	originRoot *tview.TreeNode
	*tview.TreeView
}

func NewTree() *Tree {
	t := &Tree{
		TreeView: tview.NewTreeView(),
	}

	t.SetBorder(true).SetTitle("files").SetTitleAlign(tview.AlignLeft)
	return t
}

func (t *Tree) GetSearchWord() string {
	return t.searchWord
}

func (t *Tree) SetSearchWord(word string) {
	t.searchWord = word
}

func (t *Tree) SearchFiles(gui *Gui) {
	// file search
}

func (t *Tree) UpdateView() {
	// TODO restore cursor position
	//current, err := os.Getwd()
	//if err != nil {
	//	log.Println(err)
	//	return
	//}

	//t.SetEntries(current)
}

func (t *Tree) GetSelectEntry() *File {
	f, ok := t.GetCurrentNode().GetReference().(*File)
	if !ok {
		return nil
	}
	return f
}

func (t *Tree) ChangeDir(gui *Gui, current string, target string) error {
	t.searchWord = ""

	root := tview.NewTreeNode(".")
	t.SetRoot(root).SetCurrentNode(root)
	originRoot := *root
	t.originRoot = &originRoot

	t.SetEntries(target)

	if err := os.Chdir(target); err != nil {
		log.Println(err)
		return err
	}

	gui.InputPath.SetText(target)
	return nil
}

func (t *Tree) Keybinding(gui *Gui) {
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
		case tcell.KeyF1:
			gui.Help.UpdateView(FilesPanel)
			gui.Pages.AddAndSwitchToPage("help", gui.Modal(gui.Help, 0, 0), true).ShowPage("main")
		}

		switch event.Rune() {
		case '?':
			gui.Help.UpdateView(FilesPanel)
			gui.Pages.AddAndSwitchToPage("help", gui.Modal(gui.Help, 0, 0), true).ShowPage("main")

		case 'H':
			current := gui.InputPath.GetText()
			t.ChangeDir(gui, current, filepath.Dir(current))

		case 'L':
			f := t.GetSelectEntry()
			if f != nil && f.IsDir {
				t.ChangeDir(gui, gui.InputPath.GetText(), f.PathName)
			}

		case 'h':
			t.GetCurrentNode().Collapse()

		case 'l':
			node := t.GetCurrentNode()
			node.Expand()
			f := t.GetSelectEntry()
			if f != nil && f.IsDir {
				t.AddNode(node, f.PathName)
			}

		case 'd':
			if len(t.files) == 0 {
				return event
			}

			gui.Confirm("do you want to remove this?", "yes", FilesPanel, func() error {
				entry := t.GetSelectEntry()
				if entry == nil {
					return nil
				}

				if entry.IsDir {
					if err := system.RemoveDirAll(entry.PathName); err != nil {
						log.Println(err)
						return err
					}
				} else {
					if err := system.RemoveFile(entry.PathName); err != nil {
						log.Println(err)
						return err
					}
				}

				path := gui.InputPath.GetText()
				t.SetEntries(path)
				return nil
			})

		// copy entry
		case 'y':
			if len(t.files) == 0 {
				return event
			}

			//entry := t.GetSelectEntry()
			//gui.Register.CopySource = entry

			//row, _ := t.GetSelection()
			//for i := 0; i < 5; i++ {
			//	t.GetCell(row, i).SetTextColor(tcell.ColorYellow)
			//}

		// paste entry
		case 'p':
			if gui.Register.CopySource != nil {
				source := gui.Register.CopySource

				gui.Form(map[string]string{"name": source.Name}, "paste", "new name", "new_name", FilesPanel,
					7, func(values map[string]string) error {
						name := values["name"]
						if name == "" {
							return ErrNoNewName
						}

						target := filepath.Join(gui.InputPath.GetText(), name)
						if err := system.Copy(source.PathName, target); err != nil {
							log.Println(err)
							return err
						}

						gui.Register.CopySource = nil
						t.SetEntries(gui.InputPath.GetText())
						return nil
					})
			}

		// edit file with $EDITOR
		case 'e':
			entry := t.GetSelectEntry()
			if entry == nil {
				log.Println("cannot get entry")
				return event
			}

			if err := gui.EditFile(entry.PathName); err != nil {
				gui.Message(err.Error(), FilesPanel)
			}

		case 'm':
			gui.Form(map[string]string{"name": ""}, "create", "new direcotry",
				"create_directory", FilesPanel,
				7, func(values map[string]string) error {
					name := values["name"]
					if name == "" {
						return ErrNoDirName
					}

					target := filepath.Join(gui.InputPath.GetText(), name)
					if err := system.NewDir(target); err != nil {
						log.Println(err)
						return err
					}

					t.SetEntries(gui.InputPath.GetText())
					return nil
				})
		case 'r':
			entry := t.GetSelectEntry()
			if entry == nil {
				return event
			}

			gui.Form(map[string]string{"new name": entry.Name}, "rename", "new name", "rename", FilesPanel,
				7, func(values map[string]string) error {
					name := values["new name"]
					if name == "" {
						return ErrNoFileName
					}

					current := gui.InputPath.GetText()

					target := filepath.Join(current, name)
					if err := system.Rename(entry.PathName, target); err != nil {
						return err
					}

					t.SetEntries(gui.InputPath.GetText())
					return nil
				})

		case 'n':
			gui.Form(map[string]string{"name": ""}, "create", "new file", "create_file", FilesPanel,
				7, func(values map[string]string) error {
					name := values["name"]
					if name == "" {
						return ErrNoFileOrDirName
					}

					target := filepath.Join(gui.InputPath.GetText(), name)
					if err := system.NewFile(target); err != nil {
						log.Println(err)
						return err
					}

					t.SetEntries(gui.InputPath.GetText())
					return nil
				})
		case 'q':
			gui.Stop()

		case 'o':
			entry := t.GetSelectEntry()
			if entry == nil {
				return event
			}
			if err := system.Open(entry.PathName); err != nil {
				gui.Message(err.Error(), FilesPanel)
			}

		case 'f', '/':
			t.SearchFiles(gui)

		case ':', 'c':
			gui.FocusPanel(CmdLinePanel)

		case '.':
			if err := gui.EditFile(gui.Config.ConfigFile); err != nil {
				gui.Message(err.Error(), FilesPanel)
			}

		case 'b':
			if gui.Config.Bookmark.Enable {
				entry := t.GetSelectEntry()
				if entry != nil && entry.IsDir {
					if err := gui.Bookmark.Add(entry.PathName); err != nil {
						gui.Message(err.Error(), FilesPanel)
					}
				}
			}

		case 'B':
			if gui.Config.Bookmark.Enable {
				if err := gui.Bookmark.Update(); err != nil {
					gui.Message(err.Error(), FilesPanel)
					return event
				}
				gui.CurrentPanel = BookmarkPanel
				gui.Pages.AddAndSwitchToPage("bookmark", gui.Bookmark, true).ShowPage("main")
			}
		}

		return event
	})

	t.SetChangedFunc(func(node *tview.TreeNode) {
		if node != nil {
			file, ok := node.GetReference().(*File)
			if !ok {
				return
			}

			if gui.Config.Preview.Enable {
				gui.Preview.UpdateView(gui, file)
			}
		}
	})
}

func (t *Tree) SetEntries(path string) []*File {
	files := GetFiles(path, t.searchWord, t.ignorecase)

	if len(files) == 0 {
		t.files = nil
		return nil
	}

	t.AddNode(t.GetRoot(), path)

	t.files = files
	return files
}

func (t *Tree) AddNode(parent *tview.TreeNode, path string) {
	files := GetFiles(path, t.searchWord, t.ignorecase)

	filesLen := len(files)
	if filesLen == 0 {
		return
	}

	nodes := make([]*tview.TreeNode, filesLen)
	for i, f := range files {
		n := tview.NewTreeNode(f.Name).SetReference(f)
		if f.IsDir {
			n.SetColor(tcell.ColorDarkCyan)
		}
		nodes[i] = n
	}

	parent.SetChildren(nodes)
}
