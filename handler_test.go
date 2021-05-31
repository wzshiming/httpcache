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

func TestHandler(t *testing.T) {
	want := "OK"
	var count int64
	server := httptest.NewServer(NewHandler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, 1)
		rw.Write([]byte(want))
	})))

	var wg sync.WaitGroup
	for i := 0; i != 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(server.URL)
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

func BenchmarkCacheHandler(b *testing.B) {
	want := "OK"
	server := httptest.NewServer(NewHandler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte(want))
	})))

	for i := 0; i != b.N; i++ {
		resp, err := http.Get(server.URL)
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHandler(b *testing.B) {
	want := "OK"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte(want))
	}))

	for i := 0; i != b.N; i++ {
		resp, err := http.Get(server.URL)
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			b.Fatal(err)
		}
	}
}
