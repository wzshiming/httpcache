package httpcache

import (
	"bytes"
	"io"
	"sync"
)

func MemoryStorer() Storer {
	return &Memory{}
}

type Memory struct {
	m sync.Map
}

func (m *Memory) Get(key string) (io.ReadCloser, bool) {
	val, ok := m.m.Load(key)
	if !ok {
		return nil, false
	}
	return io.NopCloser(bytes.NewBuffer(val.(*bytes.Buffer).Bytes())), true
}

func (m *Memory) Put(key string) (io.WriteCloser, bool) {
	buffer := getBuffer()
	return &writeWithClose{
		Writer: buffer,
		close: func() error {
			val, ok := m.m.LoadOrStore(key, buffer)
			if ok {
				putBuffer(val.(*bytes.Buffer))
			}
			return nil
		},
	}, true
}

func (m *Memory) Del(key string) bool {
	val, ok := m.m.LoadAndDelete(key)
	if ok {
		putBuffer(val.(*bytes.Buffer))
	}
	return true
}
