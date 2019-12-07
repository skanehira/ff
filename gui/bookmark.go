package gui

import (
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/rivo/tview"
	"github.com/skanehira/ff/system"
)

type DBLogger struct{}

func (d DBLogger) Print(v ...interface{}) {
	log.Print(v...)
}

type Bookmark struct {
	ID   int    `gorm:"id; type:integer; primary key; autoincrement"`
	Name string `gorm:"name"`
}

type BookmarkStore struct {
	db *gorm.DB
}

func NewBookmarkStore(file string) (*BookmarkStore, error) {
	file = os.ExpandEnv(file)
	// if db file is not exist, create new db file
	if !system.IsExist(file) {
		if _, err := os.Create(file); err != nil {
			log.Println(err)
			// if can't create new file, use in memory db
			file = ":memory:"
		}
	}

	db, err := gorm.Open("sqlite3", file)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	db.SetLogger(DBLogger{})
	db.LogMode(true)

	if err := db.AutoMigrate(&Bookmark{}).Error; err != nil {
		log.Println(err)
		return nil, err
	}

	return &BookmarkStore{db: db}, nil
}

func (b *BookmarkStore) HasBookmark(name string) bool {
	var count int
	if err := b.db.Table("bookmarks").Where("name = ?", name).Count(&count).Error; err != nil {
		return false
	}

	if count > 0 {
		return true
	}

	return false
}

func (b *BookmarkStore) Save(bookmark Bookmark) error {
	if !b.HasBookmark(bookmark.Name) {
		return b.db.Create(&bookmark).Error
	}
	return nil
}

func (b *BookmarkStore) Load() ([]Bookmark, error) {
	var bookmarks []Bookmark

	if err := b.db.Table("bookmarks").Scan(&bookmarks).Error; err != nil {
		return nil, err
	}

	return bookmarks, nil
}

func (b *BookmarkStore) Delete(id int) error {
	bookmark := &Bookmark{ID: id}
	return b.db.Delete(bookmark).Error
}

type Bookmarks struct {
	store            *BookmarkStore
	entries          []*Bookmark
	searchWord       string
	enableIgnorecase bool
	*tview.Table
}

func NewBookmark(config Config) (*Bookmarks, error) {
	if !system.IsExist(config.Bookmark.File) {
		file, _ := os.OpenFile(os.ExpandEnv(config.Bookmark.File),
			os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		defer file.Close()
	}

	table := tview.NewTable().Select(0, 0).SetFixed(1, 1).SetSelectable(true, false)
	table.SetTitleAlign(tview.AlignLeft).SetTitle("bookmarks").SetBorder(true)

	store, err := NewBookmarkStore(config.Bookmark.File)
	if err != nil {
		return nil, err
	}

	return &Bookmarks{
		store:            store,
		enableIgnorecase: config.IgnoreCase,
		Table:            table,
	}, nil
}

func (b *Bookmarks) SetSearchWord(word string) {
	b.searchWord = word
}

func (b *Bookmarks) GetSearchWord() string {
	return b.searchWord
}

func (b *Bookmarks) Add(name string) error {
	bookmarks := Bookmark{
		Name: name,
	}

	return b.store.Save(bookmarks)
}

func (b *Bookmarks) Delete(id int) error {
	return b.store.Delete(id)
}

func (b *Bookmarks) Update() error {
	entries, err := b.store.Load()
	if err != nil {
		return err
	}

	b.entries = []*Bookmark{}
	for _, e := range entries {
		e := e
		b.entries = append(b.entries, &e)
	}

	return b.UpdateView()
}

func (b *Bookmarks) UpdateView() error {
	table := b.Clear()

	headers := []string{
		"Name",
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

	var entries []*Bookmark
	for _, e := range b.entries {
		var name, word string
		if b.enableIgnorecase {
			name = strings.ToLower(e.Name)
			word = strings.ToLower(b.searchWord)
		} else {
			name = e.Name
			word = b.searchWord
		}

		if strings.Index(name, word) == -1 {
			continue
		}

		entries = append(entries, &Bookmark{Name: name})
	}

	i := 1
	for _, e := range entries {
		table.SetCell(i, 0, tview.NewTableCell(e.Name))
		i++
	}

	return nil
}

func (b *Bookmarks) GetSelectEntry() *Bookmark {
	row, _ := b.GetSelection()
	if len(b.entries) == 0 {
		return nil
	}
	if row < 1 {
		return nil
	}

	if row > len(b.entries) {
		return nil
	}
	return b.entries[row-1]
}

func (e *Bookmarks) SearchBookmark(gui *Gui) {
	pageName := "search_bookmark"
	if gui.Pages.HasPage(pageName) {
		searchBookmarks.SetText(gui.Bookmark.GetSearchWord())
		gui.Pages.SendToFront(pageName).ShowPage(pageName)
	} else {
		searchBookmarks = tview.NewInputField()
		searchBookmarks.SetBorder(true).SetTitle("search bookmark").SetTitleAlign(tview.AlignLeft)
		searchBookmarks.SetChangedFunc(func(text string) {
			gui.Bookmark.SetSearchWord(text)
			gui.Bookmark.UpdateView()
		})
		searchBookmarks.SetLabel("word").SetLabelWidth(5).SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				gui.Pages.HidePage(pageName)
				gui.FocusPanel(BookmarkPanel)
			}

		})

		gui.Pages.AddAndSwitchToPage(pageName, gui.Modal(searchBookmarks, 0, 3), true).ShowPage("bookmark").ShowPage("main")
	}
}

func (b *Bookmarks) CloseBookmark(gui *Gui) {
	gui.Pages.RemovePage("bookmark").ShowPage("main")
	gui.FocusPanel(FilesPanel)
}

func (b *Bookmarks) BookmarkKeybinding(gui *Gui) {
	gui.Bookmark.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			b.CloseBookmark(gui)
		case 'd':
			entry := gui.Bookmark.GetSelectEntry()
			if entry == nil {
				return event
			}
			b.Delete(entry.ID)
			b.Update()
		case 'f', '/':
			b.SearchBookmark(gui)
		case 'a':
			b.AddBookmark(gui)
		case '?':
			gui.Help.UpdateView(BookmarkPanel)
			gui.Pages.AddAndSwitchToPage("help", gui.Modal(gui.Help, 0, 0), true).ShowPage("bookmark")
		}

		switch event.Key() {
		case tcell.KeyF1:
			gui.Help.UpdateView(BookmarkPanel)
			gui.Pages.AddAndSwitchToPage("help", gui.Modal(gui.Help, 0, 0), true).ShowPage("bookmark")
		case tcell.KeyCtrlG:
			entry := gui.Bookmark.GetSelectEntry()
			if entry == nil {
				return event
			}

			if err := gui.FileBrowser.ChangeDir(gui, gui.InputPath.GetText(), entry.Name); err != nil {
				gui.Message(err.Error(), BookmarkPanel)
				return event
			}
			b.CloseBookmark(gui)
		}

		return event
	})
}

func (b *Bookmarks) AddBookmark(gui *Gui) {
	gui.Form(map[string]string{"path": ""}, "add", "new bookmark", "new_bookmark", BookmarkPanel,
		7, func(values map[string]string) error {
			name := values["path"]
			if name == "" {
				return ErrNoPathName
			}
			name = os.ExpandEnv(name)

			if !system.IsExist(name) {
				return ErrNotExistPath
			}

			if err := b.Add(name); err != nil {
				return err
			}

			if err := b.Update(); err != nil {
				return err
			}

			return nil
		})

	gui.Pages.ShowPage("bookmark")
}
