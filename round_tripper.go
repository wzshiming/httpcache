package httpcache

import (
	"bytes"
	"io"
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

	data, ok := r.storer.Get(key)
	if ok {
		resp, err := unmarshalResponse(data)
		if err == nil {
			return resp, nil
		}
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
		}
		return r.base.RoundTrip(req)
	} else {
		rmut.Lock()
		defer func() {
			rmut.Unlock()
			r.muts.Delete(key)
		}()
	}

	resp, err := r.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	buffer := getBuffer()
	_, err = buffer.ReadFrom(resp.Body)
	if err != nil {
		putBuffer(buffer)
		return resp, err
	}
	resp.Body.Close()

	if buf, ok := r.storer.Put(key); ok {
		resp.Body = io.NopCloser(bytes.NewBuffer(buffer.Bytes()))
		err = marshalResponse(resp, buf)
		buf.Close()
	}
	resp.Body = &readerWithClose{
		Reader: buffer,
		close: func() error {
			putBuffer(buffer)
			return nil
		},
	}
	return resp, nil
}
