package httpcache

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

func TestRoundTripper(t *testing.T) {
	want := "OK"
	var count int64
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, 1)
		rw.Write([]byte(want))
	}))
	cli := server.Client()
	cli.Transport = NewRoundTripper(cli.Transport)

	var wg sync.WaitGroup
	for i := 0; i != 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := cli.Get(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(want, string(body)) {
				t.Fatalf("want %q, got %q", want, body)
			}
		}()
	}
	wg.Wait()

	if count != 1 {
		t.Fatalf("cache breakdown")
	}
}

func BenchmarkCacheRoundTripper(b *testing.B) {
	want := "OK"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte(want))
	}))
	cli := server.Client()
	cli.Transport = NewRoundTripper(cli.Transport)

	for i := 0; i != b.N; i++ {
		resp, err := cli.Get(server.URL)
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundTripper(b *testing.B) {
	want := "OK"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte(want))
	}))
	cli := server.Client()

	for i := 0; i != b.N; i++ {
		resp, err := cli.Get(server.URL)
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			b.Fatal(err)
		}
	}
}
