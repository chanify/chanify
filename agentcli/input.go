package agentcli

import (
	"io"
	"io/fs"
	"os"

	"github.com/mattn/go-isatty"
)

type inputFile interface {
	io.ReadCloser
	Stat() (fs.FileInfo, error)
}
type inputFiles struct {
	Stat     func(string) (fs.FileInfo, error)
	Open     func(string) (inputFile, error)
	Terminal func(*os.File) bool
}

var systemInputs = inputFiles{
	Stat:     os.Stat,
	Open:     func(path string) (inputFile, error) { return os.Open(path) },
	Terminal: func(file *os.File) bool { return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd()) },
}
