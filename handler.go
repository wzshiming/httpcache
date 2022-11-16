package httpcache

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

type Handler struct {
	option

	http.Handler
}

func NewHandler(base http.Handler, options ...Option) http.Handler {
	handler := &Handler{
		Handler: base,
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
	defer putBytes(buf)
	_, err = io.CopyBuffer(rw, resp.Body, buf)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if !h.filterer.Filter(r) {
		h.Handler.ServeHTTP(rw, r)
		return
	}
	key := h.keyer.Key(r)

	data, ok := h.storer.Get(key)
	if ok {
		err := h.unmarshalResponse(rw, data)
		if err == nil {
			data.Close()
			return
		}
		data.Close()
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
				data.Close()
				return
			}
			data.Close()
		}
		h.Handler.ServeHTTP(rw, r)
		return
	}

	rmut.Lock()
	defer func() {
		rmut.Unlock()
		h.muts.Delete(key)
	}()

	w := newResponseWriter(rw)

	h.Handler.ServeHTTP(w, r)
	if h.discarder.Discard(w) {
		return
	}

	if buf, ok := h.storer.Put(key); ok {
		err := marshalResponse(&w.response, buf)
		buf.Close()
		if err != nil {
			h.storer.Del(key)
		}
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

func (r *responseWriter) StatusCode() int {
	return r.response.StatusCode
}
