package main

import (
	"io/ioutil"
	"os"
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

// CurrentDir return current dir path
func CurrentDir() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(exe)
}

// Entries return current entries
func Entries(path string) []Entry {
	files, err := ioutil.ReadDir(path)

	if err != nil {
		panic(err)
	}

	var entries []Entry
	var access, change, create, perm, owner, group string

	for _, file := range files {
		t, err := times.Stat(file.Name())
		if err != nil {
			panic(err)
		}

		if stat, ok := file.Sys().(*syscall.Stat_t); ok {
			access = t.AccessTime().Format(dateFmt)
			change = file.ModTime().Format(dateFmt)
			if t.HasBirthTime() {
				create = t.BirthTime().Format(dateFmt)
			}
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

// SetHeader set table header
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
		//"Create",
		//"Access",
		//"Change",
	}

	SetHeader(table, headers)

	for k, entry := range entries {
		if entry.IsDir {
			table.SetCell(k+1, 0, tview.NewTableCell(entry.Name).SetTextColor(tcell.ColorDarkCyan))
			table.SetCell(k+1, 1, tview.NewTableCell(entry.Size).SetTextColor(tcell.ColorDarkCyan))
			table.SetCell(k+1, 2, tview.NewTableCell(entry.Permission).SetTextColor(tcell.ColorDarkCyan))
			table.SetCell(k+1, 3, tview.NewTableCell(entry.Owner).SetTextColor(tcell.ColorDarkCyan))
			table.SetCell(k+1, 4, tview.NewTableCell(entry.Group).SetTextColor(tcell.ColorDarkCyan))
			//table.SetCell(k+1, 5, tview.NewTableCell(entry.Create).SetTextColor(tcell.ColorDarkCyan))
			//table.SetCell(k+1, 6, tview.NewTableCell(entry.Access).SetTextColor(tcell.ColorDarkCyan))
			//table.SetCell(k+1, 7, tview.NewTableCell(entry.Change).SetTextColor(tcell.ColorDarkCyan))
		} else {
			table.SetCell(k+1, 0, tview.NewTableCell(entry.Name))
			table.SetCell(k+1, 1, tview.NewTableCell(entry.Size))
			table.SetCell(k+1, 2, tview.NewTableCell(entry.Permission))
			table.SetCell(k+1, 3, tview.NewTableCell(entry.Owner))
			table.SetCell(k+1, 4, tview.NewTableCell(entry.Group))
			//table.SetCell(k+1, 5, tview.NewTableCell(entry.Create))
			//table.SetCell(k+1, 6, tview.NewTableCell(entry.Access))
			//table.SetCell(k+1, 7, tview.NewTableCell(entry.Change))
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

	grid := tview.NewGrid()
	grid.AddItem(table, 0, 0, 1, 1, 0, 0, true)

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}
