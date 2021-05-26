package httpcache

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httputil"
)

func MarshalRequest(req *http.Request) ([]byte, error) {
	return httputil.DumpRequest(req, true)
}

func UnmarshalRequest(data []byte) (req *http.Request, err error) {
	return http.ReadRequest(bufio.NewReader(bytes.NewBuffer(data)))
}

func MarshalResponse(resp *http.Response) ([]byte, error) {
	return httputil.DumpResponse(resp, true)
}

func UnmarshalResponse(data []byte) (resp *http.Response, err error) {
	return http.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), nil)
}
