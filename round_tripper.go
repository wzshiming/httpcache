package httpcache

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
)

type RoundTripper struct {
	option

	base http.RoundTripper
}

func NewRoundTripper(base http.RoundTripper, options ...Option) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	crt := &RoundTripper{
		base: base,
	}
	crt.option.init(options)
	return crt
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !r.filterer.Filter(req) {
		return r.base.RoundTrip(req)
	}
	key := r.keyer.Key(req)
	if !r.noSync {
		var mutex sync.RWMutex
		mut, _ := r.muts.LoadOrStore(key, &mutex)
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
