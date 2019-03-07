package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"syscall"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"

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

// MoveHistory have the move history
type MoveHistory struct {
	idx       int
	histories []string
}

// Save save the move history
func (p *MoveHistory) Save(path string) {
	count := len(p.histories)

	// if not have history
	if count == 0 {
		p.histories = append(p.histories, path)
	} else if p.idx == count-1 {
		p.histories = append(p.histories, path)
		p.idx++
	} else {
		p.histories = append(p.histories[:p.idx+1], path)
		p.idx = len(p.histories) - 1
	}
}

// Previous return the previous history
func (p *MoveHistory) Previous() string {
	count := len(p.histories)
	if count == 0 {
		return ""
	}

	p.idx--
	if p.idx < 0 {
		p.idx = 0
	}
	return p.histories[p.idx]
}

// Next return the next history
func (p *MoveHistory) Next() string {
	count := len(p.histories)
	if count == 0 {
		return ""
	}

	p.idx++
	if p.idx >= count {
		p.idx = count - 1
	}
	return p.histories[p.idx]
}

// Entries return current entries
func Entries(path string) []Entry {
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

	return entries
}

// SetHeader set table header
func SetHeader(table *tview.Table) {
	headers := []string{"Name",
		"Size",
		"Permission",
		"Owner",
		"Group",
		//"Create",
		//"Access",
		//"Change",
	}
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

// SetEntries set table entries
func SetEntries(table *tview.Table, entries []Entry) {
	table = table.Clear()
	SetHeader(table)
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
	table.ScrollToBeginning()
}

func initialize() (int, error) {
	home, err := homedir.Dir()
	if err != nil {
		log.Println("cannot get home dir, cause:", err)
		return 1, err
	}
	logFile, err := os.OpenFile(filepath.Join(home, "filemanager.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Println("cannot open log file, cause:", err)
		return 1, err
	}

	log.SetOutput(logFile)
	return 0, nil
}

func run() (int, error) {
	// initialize application
	exitCode, err := initialize()
	if err != nil {
		return exitCode, err
	}

	app := tview.NewApplication()
	// get current path
	currentDir, err := os.Getwd()
	if err != nil {
		return 1, err
	}

	history := &MoveHistory{}
	history.Save(currentDir)

	table := tview.NewTable().SetSelectable(true, false)
	inputPath := tview.NewInputField().SetText(currentDir)

	entries := Entries(currentDir)
	SetEntries(table, entries)

	table.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}

	}).SetSelectedFunc(func(row int, column int) {
		if len(entries) == 0 {
			return
		}

		entry := entries[row-1]

		if entry.IsDir {
			path := path.Join(inputPath.GetText(), table.GetCell(row, column).Text)
			history.Save(path)
			inputPath.SetText(path)
			entries = Entries(path)
			SetEntries(table, entries)
		}
	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(inputPath)
		}

		if event.Rune() == 'h' {
			path := history.Previous()
			if path != "" {
				inputPath.SetText(path)
				entries = Entries(path)
				SetEntries(table, entries)
			}
		}

		if event.Rune() == 'l' {
			path := history.Next()
			if path != "" {
				inputPath.SetText(path)
				entries = Entries(path)
				SetEntries(table, entries)
			}
		}
		return event
	})

	inputPath.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
		if key == tcell.KeyEnter {
			path := inputPath.GetText()
			history.Save(path)
			entries = Entries(path)
			SetEntries(table, entries)

			app.SetFocus(table)
		}

	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(table)
		}
		return event
	})

	grid := tview.NewGrid().SetRows(1, 0)
	grid.AddItem(inputPath, 0, 0, 1, 1, 0, 0, true)
	grid.AddItem(table, 1, 0, 2, 2, 0, 0, true)

	if err := app.SetRoot(grid, true).SetFocus(table).Run(); err != nil {
		app.Stop()
		return 1, err
	}

	return 0, nil
}

func main() {
	exitCode, err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(exitCode)
}
