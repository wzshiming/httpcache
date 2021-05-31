package httpcache

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

type Handler struct {
	option

	base http.Handler
}

func NewHandler(base http.Handler, options ...Option) http.Handler {
	handler := &Handler{
		base: base,
	}
	handler.option.init(options)
	return handler
}

func (h *Handler) unmarshalResponse(rw http.ResponseWriter, r io.Reader) error {
	resp, err := unmarshalResponse(r)
	if err != nil {
		return err
	}
	header := rw.Header()
	for key, values := range resp.Header {
		header[key] = values
	}
	rw.WriteHeader(resp.StatusCode)
	buf := getBytes()
	io.CopyBuffer(rw, resp.Body, buf)
	putBytes(buf)
	return nil
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if !h.filterer.Filter(r) {
		h.base.ServeHTTP(rw, r)
		return
	}
	key := h.keyer.Key(r)

	data, ok := h.storer.Get(key)
	if ok {
		err := h.unmarshalResponse(rw, data)
		if err == nil {
			return
		}
	}

	var mutex sync.RWMutex
	mut, ok := h.muts.LoadOrStore(key, &mutex)
	rmut := mut.(*sync.RWMutex)
	if ok {
		rmut.RLock()
		defer rmut.RUnlock()
		data, ok := h.storer.Get(key)
		if ok {
			err := h.unmarshalResponse(rw, data)
			if err == nil {
				return
			}
		}
		h.base.ServeHTTP(rw, r)
		return
	} else {
		rmut.Lock()
		defer func() {
			rmut.Unlock()
			h.muts.Delete(key)
		}()
	}

	w := newResponseWriter(rw)

	h.base.ServeHTTP(w, r)

	if buf, ok := h.storer.Put(key); ok {
		marshalResponse(&w.response, buf)
		buf.Close()
	}
}

type responseWriter struct {
	responseWriter http.ResponseWriter
	response       http.Response
	buf            *bytes.Buffer
	io.Writer
}

func newResponseWriter(rw http.ResponseWriter) *responseWriter {
	r := &responseWriter{
		responseWriter: rw,
		buf:            getBuffer(),
	}
	r.response.StatusCode = http.StatusOK
	r.response.Header = rw.Header()
	r.response.Body = &readerWithClose{
		Reader: r.buf,
		close: func() error {
			putBuffer(r.buf)
			return nil
		},
	}
	r.Writer = io.MultiWriter(r.responseWriter, r.buf)
	return r
}

func (r *responseWriter) Header() http.Header {
	return r.response.Header
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.responseWriter.WriteHeader(statusCode)
	r.response.StatusCode = statusCode
}
