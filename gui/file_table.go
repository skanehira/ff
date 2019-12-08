package gui

import (
	"os"
	"path/filepath"

	"log"

	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/skanehira/ff/system"
)

var (
	dateFmt = "2006-01-02 15:04:05"
)

type selectPos struct {
	row int
	col int
}

// FileTable file list
type FileTable struct {
	enableIgnorecase bool
	files            []*File
	selectPos        map[string]selectPos
	searchWord       string
	*tview.Table
}

// NewFileTable new entry list
func NewFileTable(enableIgnorecase bool) *FileTable {
	e := &FileTable{
		enableIgnorecase: enableIgnorecase,
		Table:            tview.NewTable().Select(0, 0).SetFixed(1, 1).SetSelectable(true, false),
		selectPos:        make(map[string]selectPos),
	}

	e.SetBorder(true).SetTitle("files").SetTitleAlign(tview.AlignLeft)

	return e
}

// Entries get entries
func (e *FileTable) Entries() []*File {
	return e.files
}

func (e *FileTable) SetSearchWord(word string) {
	e.searchWord = word
}

func (e *FileTable) GetSearchWord() string {
	return e.searchWord
}

// SetSelectPos save select position
func (e *FileTable) SetSelectPos(path string) {
	row, col := e.GetSelection()
	e.selectPos[path] = selectPos{row, col}
}

// RestorePos restore select position
func (e *FileTable) RestorePos(path string) {
	pos, ok := e.selectPos[path]
	if !ok {
		pos = selectPos{1, 0}
	}

	e.Select(pos.row, pos.col)
}

// SetEntries set entries
func (e *FileTable) SetEntries(path string) []*File {
	files := GetFiles(path, e.searchWord, e.enableIgnorecase)

	if len(files) == 0 {
		e.files = nil
		e.SetColumns()
		return nil
	}

	e.files = files
	e.SetColumns()
	return files
}

func (e *FileTable) RefreshView() {
	e.SetColumns()
}

func (e *FileTable) SetViewable(viewable bool) {
	entry := e.GetSelectEntry()
	entry.Viewable = viewable
	e.RefreshView()
}

// SetHeader set table header
func (e *FileTable) SetHeader() {
	headers := []string{
		"Name",
		"Size",
		"Permission",
		"Owner",
		"Group",
	}
	for k, v := range headers {
		e.Table.SetCell(0, k, &tview.TableCell{
			Text:            v,
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorYellow,
			BackgroundColor: tcell.ColorDefault,
		})
	}
}

// SetColumns set entries
func (e *FileTable) SetColumns() {
	table := e.Clear()
	e.SetHeader()
	var i int
	for _, entry := range e.files {
		size := humanize.Bytes(uint64(entry.Size))
		table.SetCell(i+1, 0, tview.NewTableCell(entry.Name))
		table.SetCell(i+1, 1, tview.NewTableCell(size))
		table.SetCell(i+1, 2, tview.NewTableCell(entry.Permission))
		table.SetCell(i+1, 3, tview.NewTableCell(entry.Owner))
		table.SetCell(i+1, 4, tview.NewTableCell(entry.Group))
		i++
	}

	e.UpdateColor()
}

// GetSelectEntry get selected entry
func (e *FileTable) GetSelectEntry() *File {
	row, _ := e.GetSelection()
	if len(e.files) == 0 {
		return nil
	}
	if row < 1 {
		return nil
	}

	if row > len(e.files) {
		return nil
	}
	return e.files[row-1]
}

func (e *FileTable) UpdateColor() {
	rowNum := e.GetRowCount()

	e.GetSelection()
	for i := 1; i < rowNum; i++ {
		color := tcell.ColorWhite
		if e.files[i-1].IsDir {
			color = tcell.ColorDarkCyan
		}

		for j := 0; j < 5; j++ {
			e.GetCell(i, j).SetTextColor(color)
		}
	}

}

func (e *FileTable) UpdateView() {
	current, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}

	e.SetEntries(current)
}

func (e *FileTable) ChangeDir(gui *Gui, current, target string) error {
	e.searchWord = ""
	if gui.Config.Bookmark.Enable {
		gui.Bookmark.SetSearchWord("")
	}

	// save select position
	e.SetSelectPos(current)

	// update files
	e.SetEntries(target)

	// if current postion is over than bottom entry position
	row, _ := e.GetSelection()
	count := e.GetRowCount()
	if row > count {
		e.Select(count-1, 0)
	}

	if gui.Config.Preview.Enable {
		entry := e.GetSelectEntry()
		gui.Preview.UpdateView(gui, entry)
	}

	if err := os.Chdir(target); err != nil {
		log.Println(err)
		return err
	}

	// restore select position
	e.RestorePos(target)

	gui.InputPath.SetText(target)

	return nil
}

func (e *FileTable) Keybinding(gui *Gui) {
	e.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		gui.commonFileBrowserKeybinding(event)

		switch event.Key() {
		case tcell.KeyF1:
			gui.Help.UpdateView(FileTablePanel)
			gui.Pages.AddAndSwitchToPage("help", gui.Modal(gui.Help, 0, 0), true).ShowPage("main")
		}

		switch event.Rune() {
		case '?':
			gui.Help.UpdateView(FileTablePanel)
			gui.Pages.AddAndSwitchToPage("help", gui.Modal(gui.Help, 0, 0), true).ShowPage("main")

		case 'h':
			current := gui.InputPath.GetText()
			parent := filepath.Dir(current)

			if parent != "" {
				if err := e.ChangeDir(gui, current, parent); err != nil {
					gui.Message(err.Error(), FileTablePanel)
				}
			}

		// go to selected dir
		case 'l':
			entry := e.GetSelectEntry()

			if entry != nil && entry.IsDir {
				current := gui.InputPath.GetText()
				if err := e.ChangeDir(gui, current, entry.PathName); err != nil {
					gui.Message(err.Error(), FileTablePanel)
				}
			}

		case 'd':
			if len(e.files) == 0 {
				return event
			}

			gui.Confirm("do you want to remove this?", "yes", FileTablePanel, func() error {
				entry := e.GetSelectEntry()
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
				e.SetEntries(path)
				return nil
			})

		// copy entry
		case 'y':
			if len(e.files) == 0 {
				return event
			}

			e.UpdateColor()
			entry := e.GetSelectEntry()
			gui.Register.CopySource = entry

			row, _ := e.GetSelection()
			for i := 0; i < 5; i++ {
				e.GetCell(row, i).SetTextColor(tcell.ColorYellow)
			}

		// paste entry
		case 'p':
			if gui.Register.CopySource != nil {
				source := gui.Register.CopySource

				gui.Form(map[string]string{"name": source.Name}, "paste", "new name", "new_name", FileTablePanel,
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
						e.SetEntries(gui.InputPath.GetText())
						return nil
					})
			}

		case 'm':
			gui.Form(map[string]string{"name": ""}, "create", "new direcotry",
				"create_directory", FileTablePanel,
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

					e.SetEntries(gui.InputPath.GetText())
					return nil
				})

		case 'r':
			entry := e.GetSelectEntry()
			if entry == nil {
				return event
			}

			gui.Form(map[string]string{"new name": entry.Name}, "rename", "new name", "rename", FileTablePanel,
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

					e.SetEntries(gui.InputPath.GetText())
					return nil
				})

		case 'n':
			gui.Form(map[string]string{"name": ""}, "create", "new file", "create_file", FileTablePanel,
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

					e.SetEntries(gui.InputPath.GetText())
					return nil
				})

		case 'f', '/':
			e.SearchFiles(gui)

		}
		return event
	})

	e.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 {
			if gui.Config.Preview.Enable {
				if len(e.files) > 1 {
					gui.Preview.UpdateView(gui, e.files[row-1])
				}
			}
		}
	})

}

func (e *FileTable) SearchFiles(gui *Gui) {
	pageName := "search"
	if gui.Pages.HasPage(pageName) {
		searchFiles.SetText(gui.FileBrowser.GetSearchWord())
		gui.Pages.ShowPage(pageName)
	} else {
		searchFiles = tview.NewInputField()
		searchFiles.SetBorder(true).SetTitle("search").SetTitleAlign(tview.AlignLeft)
		searchFiles.SetChangedFunc(func(text string) {
			gui.FileBrowser.SetSearchWord(text)
			current := gui.InputPath.GetText()
			gui.FileBrowser.SetEntries(current)

			if gui.Config.Preview.Enable {
				gui.Preview.UpdateView(gui, gui.FileBrowser.GetSelectEntry())
			}
		})
		searchFiles.SetLabel("word").SetLabelWidth(5).SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				gui.Pages.HidePage(pageName)
				gui.FocusPanel(FileTablePanel)
			}

		})

		gui.Pages.AddAndSwitchToPage(pageName, gui.Modal(searchFiles, 0, 3), true).ShowPage("main")
	}
}
