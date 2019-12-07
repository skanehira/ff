package gui

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"log"

	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/skanehira/ff/system"
	"gopkg.in/djherbis/times.v1"
)

var (
	dateFmt = "2006-01-02 15:04:05"
)

// Entry file or dir info
type Entry struct {
	Name       string // file name
	Path       string // file path
	PathName   string // file's path and name
	Access     string
	Create     string
	Change     string
	Size       int64
	Permission string
	Owner      string
	Group      string
	Viewable   bool
	IsDir      bool
}

type selectPos struct {
	row int
	col int
}

// EntryManager file list
type EntryManager struct {
	enableIgnorecase bool
	*tview.Table
	entries    []*Entry
	selectPos  map[string]selectPos
	searchWord string
}

// NewEntryManager new entry list
func NewEntryManager(enableIgnorecase bool) *EntryManager {
	e := &EntryManager{
		enableIgnorecase: enableIgnorecase,
		Table:            tview.NewTable().Select(0, 0).SetFixed(1, 1).SetSelectable(true, false),
		selectPos:        make(map[string]selectPos),
	}

	e.SetBorder(true).SetTitle("files").SetTitleAlign(tview.AlignLeft)

	return e
}

// Entries get entries
func (e *EntryManager) Entries() []*Entry {
	return e.entries
}

func (e *EntryManager) SetSearchWord(word string) {
	e.searchWord = word
}

func (e *EntryManager) GetSearchWord() string {
	return e.searchWord
}

// SetSelectPos save select position
func (e *EntryManager) SetSelectPos(path string) {
	row, col := e.GetSelection()
	e.selectPos[path] = selectPos{row, col}
}

// RestorePos restore select position
func (e *EntryManager) RestorePos(path string) {
	pos, ok := e.selectPos[path]
	if !ok {
		pos = selectPos{1, 0}
	}

	e.Select(pos.row, pos.col)
}

// SetEntries set entries
func (e *EntryManager) SetEntries(path string) []*Entry {
	var entries []*Entry

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("%s: %s\n", ErrReadDir, err)
		return nil
	}

	if len(files) == 0 {
		e.entries = entries
		e.SetColumns()
		return nil
	}

	var access, change, create, perm, owner, group string

	for _, file := range files {
		var name, word string
		if e.enableIgnorecase {
			name = strings.ToLower(file.Name())
			word = strings.ToLower(e.searchWord)
		} else {
			name = file.Name()
			word = e.searchWord
		}
		if strings.Index(name, word) == -1 {
			continue
		}
		// get file times
		pathName := filepath.Join(path, file.Name())
		t, err := times.Stat(pathName)
		if err != nil {
			log.Printf("%s: %s\n", ErrGetTime, err)
			continue
		}

		access = t.AccessTime().Format(dateFmt)
		change = file.ModTime().Format(dateFmt)
		if t.HasBirthTime() {
			create = t.BirthTime().Format(dateFmt)
		}

		perm = file.Mode().String()

		// get file permission, owner, group
		if stat, ok := file.Sys().(*syscall.Stat_t); ok {
			uid := strconv.Itoa(int(stat.Uid))
			u, err := user.LookupId(uid)
			if err != nil {
				owner = uid
			} else {
				owner = u.Username
			}
			gid := strconv.Itoa(int(stat.Gid))
			g, err := user.LookupGroupId(gid)
			if err != nil {
				group = gid
			} else {
				group = g.Name
			}
		}

		// create entriey
		entries = append(entries, &Entry{
			Name:       file.Name(),
			Access:     access,
			Create:     create,
			Change:     change,
			Size:       file.Size(),
			Permission: perm,
			IsDir:      file.IsDir(),
			Owner:      owner,
			Group:      group,
			PathName:   pathName,
			Path:       path,
			Viewable:   true,
		})
	}

	e.entries = entries
	e.SetColumns()
	return entries
}

func (e *EntryManager) RefreshView() {
	e.SetColumns()
}

func (e *EntryManager) SetViewable(viewable bool) {
	entry := e.GetSelectEntry()
	entry.Viewable = viewable
	e.RefreshView()
}

// SetHeader set table header
func (e *EntryManager) SetHeader() {
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
func (e *EntryManager) SetColumns() {
	table := e.Clear()
	e.SetHeader()
	var i int
	for _, entry := range e.entries {
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
func (e *EntryManager) GetSelectEntry() *Entry {
	row, _ := e.GetSelection()
	if len(e.entries) == 0 {
		return nil
	}
	if row < 1 {
		return nil
	}

	if row > len(e.entries) {
		return nil
	}
	return e.entries[row-1]
}

func (e *EntryManager) UpdateColor() {
	rowNum := e.GetRowCount()

	e.GetSelection()
	for i := 1; i < rowNum; i++ {
		color := tcell.ColorWhite
		if e.Entries()[i-1].IsDir {
			color = tcell.ColorDarkCyan
		}

		for j := 0; j < 5; j++ {
			e.GetCell(i, j).SetTextColor(color)
		}
	}

}

func (e *EntryManager) UpdateView() {
	current, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}

	e.SetEntries(current)
}

func (e *EntryManager) ChangeDir(gui *Gui, current, target string) error {
	e.SetSearchWord("")

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

	return nil
}

func (e *EntryManager) Keybinding(gui *Gui) {
	e.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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

		case 'h':
			current := gui.InputPath.GetText()
			parent := filepath.Dir(current)

			if parent != "" {
				if err := gui.ChangeDir(current, parent); err != nil {
					gui.Message(err.Error(), FilesPanel)
				}
			}

		// go to selected dir
		case 'l':
			entry := e.GetSelectEntry()

			if entry != nil && entry.IsDir {
				current := gui.InputPath.GetText()
				if err := gui.ChangeDir(current, entry.PathName); err != nil {
					gui.Message(err.Error(), FilesPanel)
				}
			}
		case 'd':
			if len(e.Entries()) == 0 {
				return event
			}

			gui.Confirm("do you want to remove this?", "yes", FilesPanel, func() error {
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
			if len(e.Entries()) == 0 {
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
						e.SetEntries(gui.InputPath.GetText())
						return nil
					})
			}

		// edit file with $EDITOR
		case 'e':
			entry := e.GetSelectEntry()
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

					e.SetEntries(gui.InputPath.GetText())
					return nil
				})
		case 'r':
			entry := e.GetSelectEntry()
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

					e.SetEntries(gui.InputPath.GetText())
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

					e.SetEntries(gui.InputPath.GetText())
					return nil
				})
		case 'q':
			gui.Stop()

		case 'o':
			entry := e.GetSelectEntry()
			if entry == nil {
				return event
			}
			if err := system.Open(entry.PathName); err != nil {
				gui.Message(err.Error(), FilesPanel)
			}

		case 'f', '/':
			e.SearchFiles(gui)

		case ':', 'c':
			gui.FocusPanel(CmdLinePanel)

		case '.':
			if err := gui.EditFile(gui.Config.ConfigFile); err != nil {
				gui.Message(err.Error(), FilesPanel)
			}

		case 'b':
			if gui.Config.Bookmark.Enable {
				entry := e.GetSelectEntry()
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

	e.SetSelectionChangedFunc(func(row, col int) {
		if row > 0 {
			if gui.Config.Preview.Enable {
				entries := e.Entries()
				if len(entries) > 1 {
					gui.Preview.UpdateView(gui, entries[row-1])
				}
			}
		}
	})

}

func (e *EntryManager) SearchFiles(gui *Gui) {
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
				gui.FocusPanel(FilesPanel)
			}

		})

		gui.Pages.AddAndSwitchToPage(pageName, gui.Modal(searchFiles, 0, 3), true).ShowPage("main")
	}
}
