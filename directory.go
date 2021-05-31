package httpcache

import (
	"io"
	"os"
	"path/filepath"
)

type Directory string

func DirectoryStorer(dir string) Directory {
	return Directory(dir)
}

func (d Directory) Get(key string) (io.Reader, bool) {
	path := filepath.Join(string(d), key)
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, false
	}
	return &autoCloser{
		auto: f,
	}, true
}

func (d Directory) Put(key string) (io.WriteCloser, bool) {
	path := filepath.Join(string(d), key)
	os.MkdirAll(filepath.Dir(path), 0777)
	w, err := writeToCompletion(path, 0666)
	if err != nil {
		return nil, false
	}
	return w, true
}

func (d Directory) Del(key string) {
	path := filepath.Join(string(d), key)
	os.Remove(path)
}

func writeToCompletion(path string, mode os.FileMode) (io.WriteCloser, error) {
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return nil, err
	}
	return &writeWithClose{
		Writer: f,
		close: func() error {
			err := f.Close()
			if err != nil {
				return err
			}
			return os.Rename(tmp, path)
		},
	}, nil
}
