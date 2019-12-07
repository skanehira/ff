package gui

import (
	"io/ioutil"
	"log"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"gopkg.in/djherbis/times.v1"
)

// File file or dir info
type File struct {
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

func GetFiles(path, searchWord string, ignorecase bool) []*File {
	var files []*File

	entries, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("%s: %s\n", ErrReadDir, err)
		return nil
	}

	if len(entries) == 0 {
		return nil
	}

	var access, change, create,
		perm, owner, group string

	for _, file := range entries {
		var name, word string
		if ignorecase {
			name = strings.ToLower(file.Name())
			word = strings.ToLower(searchWord)
		} else {
			name = file.Name()
			word = searchWord
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
		files = append(files, &File{
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

	return files
}
