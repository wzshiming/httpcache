package httpcache

import (
	"bytes"
	"io"
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
			data.Close()
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
				data.Close()
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
		if err != nil {
			r.storer.Del(key)
		}
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

type response struct {
	*http.Response
}

func (w response) Header() http.Header {
	return w.Response.Header
}

func (w response) StatusCode() int {
	return w.Response.StatusCode
}
