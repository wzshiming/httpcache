package httpcache

import (
	"net/http"
	"os"
	"path/filepath"
)

type Directory string

func DirectoryStorer(dir string) Directory {
	return Directory(dir)
}

func (f Directory) Get(key string) (*http.Response, bool) {
	path := filepath.Join(string(f), key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	resp, err := UnmarshalResponse(data)
	if err != nil {
		return nil, false
	}
	return resp, true
}

func (f Directory) Put(key string, resp *http.Response) {
	path := filepath.Join(string(f), key)
	data, err := MarshalResponse(resp)
	if err != nil {
		return
	}
	os.MkdirAll(filepath.Dir(path), 0777)
	err = writeToCompletion(path, data, 0666)
	if err != nil {
		return
	}
}

func (f Directory) Del(key string) {
	path := filepath.Join(string(f), key)
	os.Remove(path)
}

func writeToCompletion(path string, data []byte, mode os.FileMode) error {
	tmp := path + ".tmp"
	err := os.WriteFile(tmp, data, mode)
	if err != nil {
		return err
	}
	err = os.Rename(tmp, path)
	if err != nil {
		return err
	}
	return nil
}
