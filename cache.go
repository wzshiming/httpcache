package httpcache

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
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

type Option func(c *RoundTripper)

func WithStorer(storer Storer) func(c *RoundTripper) {
	return func(c *RoundTripper) {
		c.storer = storer
	}
}

func WithKeyer(keyer Keyer) func(c *RoundTripper) {
	return func(c *RoundTripper) {
		c.keyer = keyer
	}
}

func WithFilterer(filterer Filterer) func(c *RoundTripper) {
	return func(c *RoundTripper) {
		c.filterer = filterer
	}
}

func NewRoundTripper(base http.RoundTripper, options ...Option) *RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	crt := &RoundTripper{
		base: base,
	}
	for _, option := range options {
		option(crt)
	}
	if crt.keyer == nil {
		crt.keyer = PathKeyer()
	}
	if crt.storer == nil {
		crt.storer = MemoryStorer()
	}
	if crt.filterer == nil {
		crt.filterer = MethodFilterer(http.MethodGet)
	}
	return crt
}

type RoundTripper struct {
	filterer Filterer
	keyer    Keyer
	storer   Storer
	base     http.RoundTripper

	muts   sync.Map
	noSync bool
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !r.filterer.Filter(req) {
		return r.base.RoundTrip(req)
	}
	key := r.keyer.Key(req)
	if !r.noSync {
		mut, _ := r.muts.LoadOrStore(key, &sync.RWMutex{})
		rmut := mut.(*sync.RWMutex)
		rmut.Lock()
		defer func() {
			rmut.Unlock()
			r.muts.Delete(key)
		}()
	}

	resp, ok := r.storer.Get(key)
	if ok {
		return resp, nil
	}

	resp, err := r.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}
	resp.Body.Close()

	buf := bytes.NewReader(content)
	resp.Body = ioutil.NopCloser(buf)
	r.storer.Put(key, resp)
	buf.Seek(0, io.SeekStart)
	return resp, nil
}
