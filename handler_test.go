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
	server := httptest.NewTLSServer(NewHandler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, 1)
		rw.Header().Set("want", want)
		rw.WriteHeader(http.StatusAccepted)
		rw.Write([]byte(want))
	})))
	cli := server.Client()

	for i := 0; i != 100; i++ {
		resp, err := cli.Get(server.URL + "/handler")
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusAccepted {
			t.Fatalf("want %q, got %q", http.StatusAccepted, resp.StatusCode)
		}
		if got := resp.Header.Get("want"); !reflect.DeepEqual(want, got) {
			t.Fatalf("want %q, got %q", want, got)
		}
		if !reflect.DeepEqual(want, string(body)) {
			t.Fatalf("want %q, got %q", want, body)
		}
	}

	if count != 1 {
		t.Fatalf("cache breakdown")
	}
}

func TestParallelHandler(t *testing.T) {
	want := "OK"
	var count int64
	server := httptest.NewTLSServer(NewHandler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, 1)
		rw.Header().Set("want", want)
		rw.WriteHeader(http.StatusAccepted)
		rw.Write([]byte(want))
	})))
	cli := server.Client()

	var wg sync.WaitGroup
	for i := 0; i != 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := cli.Get(server.URL + "/handler")
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusAccepted {
				t.Fatalf("want %q, got %q", http.StatusAccepted, resp.StatusCode)
			}
			if got := resp.Header.Get("want"); !reflect.DeepEqual(want, got) {
				t.Fatalf("want %q, got %q", want, got)
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

func BenchmarkCacheMemoryHandler(b *testing.B) {
	want := "OK"
	server := httptest.NewTLSServer(NewHandler(
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte(want))
		}),
		WithStorer(MemoryStorer()),
	))
	cli := server.Client()

	for i := 0; i != b.N; i++ {
		resp, err := cli.Get(server.URL + "/handler")
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

func BenchmarkCacheDirectoryHandler(b *testing.B) {
	want := "OK"
	server := httptest.NewTLSServer(NewHandler(
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte(want))
		}),
		WithStorer(DirectoryStorer("./tmp/")),
	))
	cli := server.Client()

	for i := 0; i != b.N; i++ {
		resp, err := cli.Get(server.URL + "/handler")
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

func BenchmarkHandler(b *testing.B) {
	want := "OK"
	server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte(want))
	}))
	cli := server.Client()

	for i := 0; i != b.N; i++ {
		resp, err := cli.Get(server.URL + "/handler")
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

func BenchmarkParallelCacheMemoryHandler(b *testing.B) {
	want := "OK"
	server := httptest.NewTLSServer(NewHandler(
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte(want))
		}),
		WithStorer(MemoryStorer()),
	))
	cli := server.Client()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := cli.Get(server.URL + "/handler")
			if err != nil {
				b.Fatal(err)
			}
			_, err = io.ReadAll(resp.Body)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}

func BenchmarkParallelCacheDirectoryHandler(b *testing.B) {
	want := "OK"
	server := httptest.NewTLSServer(NewHandler(
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte(want))
		}),
		WithStorer(DirectoryStorer("./tmp/")),
	))
	cli := server.Client()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := cli.Get(server.URL + "/handler")
			if err != nil {
				b.Fatal(err)
			}
			_, err = io.ReadAll(resp.Body)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}

func BenchmarkParallelHandler(b *testing.B) {
	want := "OK"
	server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte(want))
	}))
	cli := server.Client()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := cli.Get(server.URL + "/handler")
			if err != nil {
				b.Fatal(err)
			}
			_, err = io.ReadAll(resp.Body)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
}
