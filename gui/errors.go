package gui

import "errors"

var (
	ErrGetCwd       = errors.New("can't get current dir")
	ErrEdit         = errors.New("can't edit")
	ErrReadFile     = errors.New("can't read file")
	ErrReadDir      = errors.New("can't read dir")
	ErrTokenise     = errors.New("can't tokenise")
	ErrGetTime      = errors.New("can't get timespec")
	ErrNoPathName   = errors.New("no path name")
	ErrNotExistPath = errors.New("not exist path")
	ErrNoEditor     = errors.New("$EDITOR is empty")
)
