package httpcache

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
)

func marshalResponse(resp *http.Response, w io.Writer) error {
	return resp.Write(w)
}

func unmarshalResponse(r io.Reader) (*http.Response, error) {
	br := getReader(r)
	tp := textproto.NewReader(br)
	resp := &http.Response{}

	// Parse the first line of the response.
	line, err := tp.ReadLine()
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	if i := strings.IndexByte(line, ' '); i == -1 {
		return nil, fmt.Errorf("malformed HTTP response %q", line)
	} else {
		resp.Proto = line[:i]
		resp.Status = strings.TrimLeft(line[i+1:], " ")
	}
	statusCode := resp.Status
	if i := strings.IndexByte(resp.Status, ' '); i != -1 {
		statusCode = resp.Status[:i]
	}
	if len(statusCode) != 3 {
		return nil, fmt.Errorf("malformed HTTP status code %s", statusCode)
	}
	resp.StatusCode, err = strconv.Atoi(statusCode)
	if err != nil || resp.StatusCode < 0 {
		return nil, fmt.Errorf("malformed HTTP status code %s", statusCode)
	}
	var ok bool
	if resp.ProtoMajor, resp.ProtoMinor, ok = http.ParseHTTPVersion(resp.Proto); !ok {
		return nil, fmt.Errorf("malformed HTTP version %s", resp.Proto)
	}

	// Parse the response headers.
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	resp.Header = http.Header(mimeHeader)
	resp.Body = &readerWithClose{
		Reader: br,
		close: func() error {
			putReader(br)
			return nil
		},
	}
	return resp, nil
}

var poolReader = sync.Pool{
	New: func() interface{} {
		return bufio.NewReader(nil)
	},
}

func putReader(r *bufio.Reader) {
	r.Reset(nil)
	poolReader.Put(r)
}

func getReader(reader io.Reader) *bufio.Reader {
	r := poolReader.Get().(*bufio.Reader)
	r.Reset(reader)
	return r
}

var poolBytes = sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024)
	},
}

func putBytes(buf []byte) {
	poolBytes.Put(buf)
}

func getBytes() []byte {
	return poolBytes.Get().([]byte)
}

var poolBuffer = sync.Pool{
	New: func() interface{} {
		buf := bytes.NewBuffer(getBytes())
		buf.Reset()
		return buf
	},
}

func putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	poolBuffer.Put(buf)
}

func getBuffer() *bytes.Buffer {
	return poolBuffer.Get().(*bytes.Buffer)
}

type writeWithClose struct {
	io.Writer
	close func() error
}

func (f *writeWithClose) Close() error {
	return f.close()
}

type readerWithClose struct {
	io.Reader
	close func() error
}

func (r *readerWithClose) Close() error {
	return r.close()
}

type autoCloser struct {
	auto    io.ReadCloser
	isClose bool
}

func (a *autoCloser) Close() error {
	a.isClose = true
	return a.auto.Close()
}

func (a *autoCloser) Read(p []byte) (int, error) {
	if a.isClose {
		return 0, io.EOF
	}
	n, err := a.auto.Read(p)
	if err == io.EOF {
		a.isClose = true
		a.auto.Close()
	}
	return n, err
}
