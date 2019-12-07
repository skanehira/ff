package gui

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
