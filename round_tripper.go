package httpcache

import (
	"fmt"
	"net/http"
	"sync"
)

type RoundTripper struct {
	option

	http.RoundTripper
}

func NewRoundTripper(base http.RoundTripper, options ...Option) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	crt := &RoundTripper{
		RoundTripper: base,
	}
	crt.option.init(options)
	return crt
}

func (r *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !r.filterer.Filter(req) {
		return r.RoundTripper.RoundTrip(req)
	}
	key := r.keyer.Key(req)

	data, ok := r.storer.Get(key)
	if ok {
		resp, err := unmarshalResponse(data)
		if err == nil {
			return resp, nil
		}
		data.Close()
	}

	var mutex sync.RWMutex
	mut, ok := r.muts.LoadOrStore(key, &mutex)
	rmut := mut.(*sync.RWMutex)
	if ok {
		rmut.RLock()
		defer rmut.RUnlock()
		data, ok := r.storer.Get(key)
		if ok {
			resp, err := unmarshalResponse(data)
			if err == nil {
				return resp, nil
			}
			data.Close()
		}
		return r.RoundTripper.RoundTrip(req)
	}

	rmut.Lock()
	defer func() {
		rmut.Unlock()
		r.muts.Delete(key)
	}()

	resp, err := r.RoundTripper.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if r.discarder.Discard(response{resp}) {
		return resp, nil
	}

	buf, ok := r.storer.Put(key)
	if !ok {
		return resp, nil
	}

	err = marshalResponse(resp, buf)
	buf.Close()
	resp.Body.Close()
	if err != nil {
		r.storer.Del(key)
	}

	data, ok = r.storer.Get(key)
	if !ok {
		return resp, fmt.Errorf("failed to cache data: %q", key)
	}
	return unmarshalResponse(data)
}

type response struct {
	*http.Response
}

func (w response) Header() http.Header {
	return w.Response.Header
}

func (w response) StatusCode() int {
	return w.Response.StatusCode
}
