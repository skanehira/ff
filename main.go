package main

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
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

// CurrentDir return current dir path
func CurrentDir() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(exe)
}

// CurrentEntries return current entries
func Entries(path string) []Entry {
	files, err := ioutil.ReadDir(path)

	if err != nil {
		panic(err)
	}

	var entries []Entry
	var access, change, create, perm, owner, group string

	for _, file := range files {
		if stat, ok := file.Sys().(*syscall.Stat_t); ok {
			access = time.Unix(stat.Atimespec.Unix()).Format(dateFmt)
			change = time.Unix(stat.Ctimespec.Unix()).Format(dateFmt)
			create = time.Unix(stat.Birthtimespec.Unix()).Format(dateFmt)
			perm = file.Mode().String()
			u, err := user.LookupId(strconv.Itoa(int(stat.Uid)))
			if err != nil {
				panic(err)
			}
			owner = u.Username
			g, err := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))
			if err != nil {
				panic(err)
			}
			group = g.Name
		}

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

	return entries
}

func SetHeader(table *tview.Table, headers []string) {
	for k, v := range headers {
		table.SetCell(0, k, &tview.TableCell{
			Text:            v,
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorYellow,
			BackgroundColor: tcell.ColorDefault,
		})
	}
}

func main() {
	app := tview.NewApplication()
	table := tview.NewTable().SetSelectable(true, false)
	//table.SetBorder(true)

	entries := Entries(".")

	headers := []string{"Name",
		"Size",
		"Permission",
		"Owner",
		"Group",
	}

	SetHeader(table, headers)

	for k, entry := range entries {
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

	table.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
	}).SetSelectedFunc(func(row int, column int) {
		// TODO
		table.GetCell(row, column).SetText("gorilla")
	})

	grid := tview.NewGrid().SetColumns(0, 0)
	grid.AddItem(table, 0, 0, 1, 1, 0, 0, true)

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}
