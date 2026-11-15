package logic

import (
	"io"
	"io/fs"
	"os"
)

// File operations stay local to a Logic instance so failure tests do not change
// filesystem behavior for other instances or concurrent requests.
type writableFile interface {
	io.Writer
	io.Closer
	Name() string
}
type fileOperations struct {
	Stat       func(string) (fs.FileInfo, error)
	Lstat      func(string) (fs.FileInfo, error)
	Open       func(string) (*os.File, error)
	OpenFile   func(string, int, fs.FileMode) (writableFile, error)
	CreateTemp func(string, string) (writableFile, error)
	MkdirAll   func(string, fs.FileMode) error
	ReadDir    func(string) ([]os.DirEntry, error)
	Remove     func(string) error
	Rename     func(string, string) error
}

var systemFiles = fileOperations{
	Stat: os.Stat, Lstat: os.Lstat, Open: os.Open,
	OpenFile: func(name string, flags int, mode fs.FileMode) (writableFile, error) {
		return os.OpenFile(name, flags, mode)
	},
	CreateTemp: func(dir, pattern string) (writableFile, error) { return os.CreateTemp(dir, pattern) },
	MkdirAll:   os.MkdirAll, ReadDir: os.ReadDir, Remove: os.Remove, Rename: os.Rename,
}

func (l *Logic) files() *fileOperations {
	if l.fileOps != nil {
		return l.fileOps
	}
	return &systemFiles
}
