package httpcache

import (
	"net/http"
	"sync"
)

func MemoryStorer() *Memory {
	return &Memory{}
}

type Memory struct {
	m sync.Map
}

func (m *Memory) Get(key string) (*http.Response, bool) {
	val, ok := m.m.Load(key)
	if !ok {
		return nil, false
	}
	resp, err := UnmarshalResponse(val.([]byte))
	if err != nil {
		return nil, false
	}
	return resp, true
}

func (m *Memory) Put(key string, resp *http.Response) {
	data, err := MarshalResponse(resp)
	if err != nil {
		return
	}
	m.m.Store(key, data)
}

func (m *Memory) Del(key string) {
	m.m.Delete(key)
}
