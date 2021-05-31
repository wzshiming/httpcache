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

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if !h.filterer.Filter(r) {
		h.base.ServeHTTP(rw, r)
		return
	}
	key := h.keyer.Key(r)
	if !h.noSync {
		var mutex sync.RWMutex
		mut, _ := h.muts.LoadOrStore(key, &mutex)
		rmut := mut.(*sync.RWMutex)
		rmut.Lock()
		defer func() {
			rmut.Unlock()
			h.muts.Delete(key)
		}()
	}

	resp, ok := h.storer.Get(key)
	if ok {
		header := rw.Header()
		for key, values := range resp.Header {
			header[key] = values
		}
		rw.WriteHeader(resp.StatusCode)
		io.Copy(rw, resp.Body)
		return
	}

	w := newResponseWriter(rw)
	defer func() {
		h.storer.Put(key, &w.response)
	}()

	h.base.ServeHTTP(w, r)
}

type responseWriter struct {
	responseWriter http.ResponseWriter
	response       http.Response
	buf            bytes.Buffer
	io.Writer
}

func newResponseWriter(rw http.ResponseWriter) *responseWriter {
	r := &responseWriter{
		responseWriter: rw,
	}
	r.response.StatusCode = http.StatusOK
	r.response.Header = rw.Header()
	r.response.Body = io.NopCloser(&r.buf)
	r.Writer = io.MultiWriter(r.responseWriter, &r.buf)
	return r
}

func (r *responseWriter) Header() http.Header {
	return r.response.Header
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.WriteHeader(statusCode)
	r.response.StatusCode = statusCode
}
