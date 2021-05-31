package httpcache

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"path"
)

func JointKeyer(keyers ...Keyer) Keyer {
	switch len(keyers) {
	case 0:
		return KeyerFunc(noKey)
	case 1:
		return keyers[0]
	}
	return KeyerFunc(func(req *http.Request) string {
		keys := make([]string, 0, len(keyers))
		for _, keyer := range keyers {
			keys = append(keys, keyer.Key(req))
		}
		return path.Join(keys...)
	})
}

func hashKeyer(f func(buf *bytes.Buffer, req *http.Request) []byte) Keyer {
	return KeyerFunc(func(req *http.Request) string {
		buf := getBuffer()
		defer putBuffer(buf)
		hash := md5.New()
		hash.Write(f(buf, req))
		var tmp [md5.Size]byte
		return hex.EncodeToString(hash.Sum(tmp[:0]))
	})
}

func base64Keyer(f func(buf *bytes.Buffer, req *http.Request) []byte) Keyer {
	return KeyerFunc(func(req *http.Request) string {
		buf := getBuffer()
		defer putBuffer(buf)
		return base64.RawStdEncoding.EncodeToString(f(buf, req))
	})
}

func BodyKeyer() Keyer {
	return hashKeyer(func(buf *bytes.Buffer, req *http.Request) []byte {
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(body))
		return body
	})
}

func QueryKeyer(names ...string) Keyer {
	return base64Keyer(func(buf *bytes.Buffer, req *http.Request) []byte {
		query := req.URL.Query()
		for _, name := range names {
			values, ok := query[name]
			if !ok {
				continue
			}
			if len(values) == 0 {
				buf.WriteString(name)
				buf.WriteByte('&')
				continue
			}
			for _, value := range query[name] {
				buf.WriteString(name)
				buf.WriteByte('=')
				buf.WriteString(value)
				buf.WriteByte('&')
			}
		}
		return buf.Bytes()
	})
}

func HeaderKeyer(names ...string) Keyer {
	return base64Keyer(func(buf *bytes.Buffer, req *http.Request) []byte {
		query := req.Header
		for _, name := range names {
			values, ok := query[name]
			if !ok {
				continue
			}
			if len(values) == 0 {
				buf.WriteString(name)
				buf.WriteByte('\n')
				continue
			}
			for _, value := range query[name] {
				buf.WriteString(name)
				buf.WriteByte('=')
				buf.WriteString(value)
				buf.WriteByte('\n')
			}
		}
		return buf.Bytes()
	})
}

func PathKeyer() Keyer {
	return base64Keyer(func(buf *bytes.Buffer, req *http.Request) []byte {
		if req.URL.Path == "" {
			return []byte{'/'}
		}
		return []byte(req.URL.Path)
	})
}

func MethodKeyer() Keyer {
	return KeyerFunc(func(req *http.Request) string {
		return req.Method
	})
}

func noKey(req *http.Request) string {
	return "empty"
}
