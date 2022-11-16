package httpcache

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
)

type Directory string

func DirectoryStorer(dir string) Storer {
	return Directory(dir)
}

func (d Directory) Get(key string) (io.ReadCloser, bool) {
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
	os.MkdirAll(filepath.Dir(path), 0755)
	w, err := writeToCompletion(path, 0644)
	if err != nil {
		return nil, false
	}
	return w, true
}

func (d Directory) Del(key string) bool {
	path := filepath.Join(string(d), key)
	os.Remove(path)
	return true
}

func writeToCompletion(path string, mode os.FileMode) (io.WriteCloser, error) {
	tmp := path + "." + strconv.FormatUint(rand.Uint64(), 10) + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return nil, err
	}
	return &writeWithClose{
		Writer: f,
		close: func() error {
			err := f.Sync()
			if err != nil {
				os.Remove(tmp)
				return fmt.Errorf("failed to write file %w", err)
			}
			err = f.Close()
			if err != nil {
				os.Remove(tmp)
				return fmt.Errorf("failed to write file %w", err)
			}
			err = os.Rename(tmp, path)
			if err != nil {
				os.Remove(tmp)
				return fmt.Errorf("failed to write file %w", err)
			}
			return nil
		},
	}, nil
}
