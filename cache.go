package httpcache

import (
	"net/http"
)

type Filterer interface {
	Filter(req *http.Request) bool
}

type FiltererFunc func(req *http.Request) bool

func (f FiltererFunc) Filter(req *http.Request) bool {
	return f(req)
}

type Keyer interface {
	Key(req *http.Request) string
}

type KeyerFunc func(req *http.Request) string

func (k KeyerFunc) Key(req *http.Request) string {
	return k(req)
}

type Storer interface {
	Get(key string) (*http.Response, bool)
	Put(key string, resp *http.Response)
	Del(key string)
}
