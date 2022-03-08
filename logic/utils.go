package logic

import (
	"bufio"
	"os"

	lua "github.com/yuin/gopher-lua"
	parse "github.com/yuin/gopher-lua/parse"
)

func fixPath(path string) error {
	s, err := os.Stat(path)
	if err == nil && s.IsDir() {
		return nil
	}
	return os.MkdirAll(path, os.ModePerm)
}

func saveFile(path string, data []byte) error {
	if _, err := os.Stat(path); err != nil {
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.Write(data)
		return err
	}
	return nil
}

func compileLua(filePath string) (*lua.FunctionProto, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	chunk, err := parse.Parse(reader, filePath)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, err
	}
	return proto, nil
}
