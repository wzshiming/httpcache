package httpcache

import (
	"io"
	"net/http"
)

type Filterer interface {
	Filter(req *http.Request) bool
}

type FiltererFunc func(req *http.Request) bool

func (f FiltererFunc) Filter(req *http.Request) bool {
	return f(req)
}

type Discarder interface {
	Discard(resp Response) bool
}

type Response interface {
	Header() http.Header
	StatusCode() int
}

type DiscarderFunc func(resp Response) bool

func (f DiscarderFunc) Discard(resp Response) bool {
	return f(resp)
}

type Keyer interface {
	Key(req *http.Request) string
}

type KeyerFunc func(req *http.Request) string

func (k KeyerFunc) Key(req *http.Request) string {
	return k(req)
}

type Storer interface {
	Get(key string) (io.Reader, bool)
	Put(key string) (io.WriteCloser, bool)
	Del(key string)
}
