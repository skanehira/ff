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
	selectPos  map[string]string
	expandInfo map[string]struct{}
	originRoot *tview.TreeNode
	*tview.TreeView
}

func NewTree() *Tree {
	t := &Tree{
		TreeView:   tview.NewTreeView(),
		selectPos:  make(map[string]string),
		expandInfo: make(map[string]struct{}),
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

func (t *Tree) SetSelectPos(path string) {
	n := t.GetCurrentNode()
	if n != nil {
		f, ok := n.GetReference().(*File)
		if ok {
			t.selectPos[path] = f.PathName
		}
	}
}

func (t *Tree) RestorePos(path string) {
	oldpath, ok := t.selectPos[path]
	if !ok {
		return
	}

	currentlyNode := t.GetCurrentlyNode(oldpath, t.GetRoot())
	if currentlyNode != nil {
		t.SetCurrentNode(currentlyNode)
	}

	return
}

func (t *Tree) GetCurrentlyNode(oldpath string, target *tview.TreeNode) *tview.TreeNode {
	for _, node := range target.GetChildren() {
		f, ok := node.GetReference().(*File)
		if !ok {
			continue
		}

		if oldpath == f.PathName {
			return node
		}

		if len(node.GetChildren()) > 0 {
			n := t.GetCurrentlyNode(oldpath, node)
			if n != nil {
				return n
			}
		}
	}

	return nil
}

func (t *Tree) UpdateView() {
	current, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}

	t.SetSelectPos(current)
	t.SetEntries(current)
	t.RestorePos(current)
}

func (t *Tree) GetSelectEntry() *File {
	n := t.GetCurrentNode()
	if n == nil {
		return nil
	}
	f, ok := n.GetReference().(*File)
	if !ok {
		return nil
	}
	return f
}

func (t *Tree) ChangeDir(gui *Gui, current string, target string) error {
	t.searchWord = ""
	if gui.Config.Bookmark.Enable {
		gui.Bookmark.SetSearchWord("")
	}
	t.SetSelectPos(current)

	root := tview.NewTreeNode(filepath.Base(target)).
		SetReference(&File{PathName: current}).SetSelectable(false)

	t.SetRoot(root).SetCurrentNode(root)
	originRoot := *root
	t.originRoot = &originRoot

	t.SetEntries(target)

	if err := os.Chdir(target); err != nil {
		log.Println(err)
		return err
	}

	t.RestorePos(target)

	gui.InputPath.SetText(target)
	return nil
}

func (t *Tree) Keybinding(gui *Gui) {
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		gui.commonFileBrowserKeybinding(event)

		switch event.Key() {
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
			e := t.GetSelectEntry()
			if e != nil {
				delete(t.expandInfo, e.PathName)
			}

		case 'l':
			node := t.GetCurrentNode()
			f := t.GetSelectEntry()
			if f != nil && f.IsDir {
				files := GetFiles(f.PathName, t.searchWord, t.ignorecase)
				t.AddNode(node, files)
				node.Expand()
				t.expandInfo[f.PathName] = struct{}{}
			}

		case 'd':
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

				t.UpdateView()
				return nil
			})

		// copy entry
		case 'y':
			entry := t.GetSelectEntry()
			if entry != nil {
				gui.Register.CopySource = entry
				t.GetCurrentNode().SetColor(tcell.ColorYellow)
			}

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

						current := gui.InputPath.GetText()

						e := t.GetSelectEntry()
						if e != nil {
							if e.IsDir {
								current = e.PathName
							} else {
								current = filepath.Dir(e.PathName)
							}
						}
						target := filepath.Join(current, name)
						if err := system.Copy(source.PathName, target); err != nil {
							log.Println(err)
							return err
						}

						gui.Register.CopySource = nil
						t.UpdateView()
						return nil
					})
			}

		case 'm':
			gui.Form(map[string]string{"name": ""}, "create", "new direcotry",
				"create_directory", FilesPanel,
				7, func(values map[string]string) error {
					name := values["name"]
					if name == "" {
						return ErrNoDirName
					}

					current := gui.InputPath.GetText()

					e := t.GetSelectEntry()
					if e != nil {
						if e.IsDir {
							current = e.PathName
						} else {
							current = filepath.Dir(e.PathName)
						}
					}
					target := filepath.Join(current, name)
					if err := system.NewDir(target); err != nil {
						log.Println(err)
						return err
					}

					t.UpdateView()
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

					e := t.GetSelectEntry()
					if e != nil {
						current = filepath.Dir(e.PathName)
					}
					target := filepath.Join(current, name)

					if err := system.Rename(entry.PathName, target); err != nil {
						return err
					}

					t.UpdateView()
					return nil
				})

		case 'n':
			gui.Form(map[string]string{"name": ""}, "create", "new file", "create_file", FilesPanel,
				7, func(values map[string]string) error {
					name := values["name"]
					if name == "" {
						return ErrNoFileOrDirName
					}

					current := gui.InputPath.GetText()

					e := t.GetSelectEntry()
					if e != nil {
						if e.IsDir {
							current = e.PathName
						} else {
							current = filepath.Dir(e.PathName)
						}
					}

					target := filepath.Join(current, name)
					if err := system.NewFile(target); err != nil {
						log.Println(err)
						return err
					}

					t.UpdateView()
					return nil
				})
		case 'f', '/':
			t.SearchFiles(gui)

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

	t.AddNode(t.GetRoot(), files)

	t.files = files
	return files
}

func (t *Tree) AddNode(parent *tview.TreeNode, files []*File) {
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
		if _, ok := t.expandInfo[f.PathName]; ok {
			files := GetFiles(f.PathName, t.searchWord, t.ignorecase)
			if len(files) != 0 {
				t.AddNode(n, files)
			}
		}
		nodes[i] = n
	}

	parent.SetChildren(nodes)
}
