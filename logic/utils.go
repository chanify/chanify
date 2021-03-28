package logic

import (
	"io/ioutil"
	"os"
)

func FixPath(path string) error {
	s, err := os.Stat(path)
	if err == nil && s.IsDir() {
		return nil
	}
	return os.MkdirAll(path, os.ModePerm)
}

func LoadFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
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
