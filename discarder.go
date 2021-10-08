package httpcache

func NormalDiscarder() Discarder {
	return DiscarderFunc(func(resp Response) bool {
		code := resp.StatusCode()
		return code < 200 || 500 <= code
	})
}
