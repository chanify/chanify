package logic

import (
	"os"
)

func FixPath(path string) error {
	s, err := os.Stat(path)
	if err == nil && s.IsDir() {
		return nil
	}
	return os.MkdirAll(path, os.ModePerm)
}

func SaveFile(path string, data []byte) error {
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
