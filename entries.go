package main

import (
	"io/ioutil"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"gopkg.in/djherbis/times.v1"
)

var (
	dateFmt = "2006-01-02 15:04:05"
)

// Entry file or dir info
type Entry struct {
	Name       string
	Path       string
	Access     string
	Create     string
	Change     string
	Size       string
	Permission string
	Owner      string
	Group      string
	IsDir      bool
}

// EntryManager file list
type EntryManager struct {
	*tview.Table
	entries []Entry
}

// NewEntryManager new entry list
func NewEntryManager() *EntryManager {
	return &EntryManager{
		Table: tview.NewTable().SetSelectable(true, false),
	}
}

// Entries get entries
func (e *EntryManager) Entries() []Entry {
	return e.entries
}

// SetEntries set entries
func (e *EntryManager) SetEntries(path string) []Entry {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return []Entry{}
	}

	var entries []Entry
	var access, change, create, perm, owner, group string

	for _, file := range files {
		// get file times
		t, err := times.Stat(filepath.Join(path, file.Name()))
		if err != nil {
			// TODO write logger
			continue
		}

		access = t.AccessTime().Format(dateFmt)
		change = file.ModTime().Format(dateFmt)
		if t.HasBirthTime() {
			create = t.BirthTime().Format(dateFmt)
		}

		// get file permission, owner, group
		if stat, ok := file.Sys().(*syscall.Stat_t); ok {
			perm = file.Mode().String()

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
		entries = append(entries, Entry{
			Name:       file.Name(),
			Access:     access,
			Create:     create,
			Change:     change,
			Size:       strconv.Itoa(int(file.Size())),
			Permission: perm,
			IsDir:      file.IsDir(),
			Owner:      owner,
			Group:      group,
			// TODO add file path
		})
	}

	e.entries = entries
	return entries
}

// SetHeader set table header
func (e *EntryManager) SetHeader() {
	headers := []string{"Name",
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
	table := e.Table.Clear()
	e.SetHeader()
	for k, entry := range e.entries {
		if entry.IsDir {
			table.SetCell(k+1, 0, tview.NewTableCell(entry.Name).SetTextColor(tcell.ColorDarkCyan))
			table.SetCell(k+1, 1, tview.NewTableCell(entry.Size).SetTextColor(tcell.ColorDarkCyan))
			table.SetCell(k+1, 2, tview.NewTableCell(entry.Permission).SetTextColor(tcell.ColorDarkCyan))
			table.SetCell(k+1, 3, tview.NewTableCell(entry.Owner).SetTextColor(tcell.ColorDarkCyan))
			table.SetCell(k+1, 4, tview.NewTableCell(entry.Group).SetTextColor(tcell.ColorDarkCyan))
		} else {
			table.SetCell(k+1, 0, tview.NewTableCell(entry.Name))
			table.SetCell(k+1, 1, tview.NewTableCell(entry.Size))
			table.SetCell(k+1, 2, tview.NewTableCell(entry.Permission))
			table.SetCell(k+1, 3, tview.NewTableCell(entry.Owner))
			table.SetCell(k+1, 4, tview.NewTableCell(entry.Group))
		}
	}
	table.ScrollToBeginning()
}
