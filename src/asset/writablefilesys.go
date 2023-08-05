package asset

import (
	"os"
	"path/filepath"
)

type WriteableFileSystem interface {
	WriteFile(path Path, data []byte) error
}

type writableFS struct {
	base Path
}

func (f *writableFS) WriteFile(path Path, data []byte) error {
	return os.WriteFile(filepath.Join(string(f.base), string(path)), data, 0777)
}

func NewWritableFS(basepath Path) WriteableFileSystem {
	return &writableFS{
		base: basepath,
	}
}
