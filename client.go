package httpcache

import (
	"net/http"
)

func NewClient(base *http.Client, options ...Option) *http.Client {
	var cli http.Client
	if base != nil {
		cli = *base
	}
	cli.Transport = NewRoundTripper(cli.Transport, options...)
	return &cli
}
