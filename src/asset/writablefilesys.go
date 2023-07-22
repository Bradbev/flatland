package asset

import (
	"io/ioutil"
	"path/filepath"
)

type WriteableFileSystem interface {
	WriteFile(path string, data []byte) error
}

type writableFS struct {
	base string
}

func (f *writableFS) WriteFile(path string, data []byte) error {
	return ioutil.WriteFile(filepath.Join(f.base, path), data, 0777)
}

func NewWritableFS(basepath string) WriteableFileSystem {
	return &writableFS{
		base: basepath,
	}
}
