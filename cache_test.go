package httpcache

import (
	"io"
	"testing"
)

func TestStorer(t *testing.T) {

	tests := []struct {
		name   string
		storer Storer
	}{
		{
			name:   "MemoryStorer",
			storer: MemoryStorer(),
		},
		{
			name:   "DirectoryStorer",
			storer: DirectoryStorer("./tmp/"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "key"
			want := "Hello"
			tt.storer.Del(key)
			_, ok := tt.storer.Get(key)
			if ok {
				t.Error("expected to be unavailable")
				return
			}
			w, ok := tt.storer.Put(key)
			if !ok {
				t.Error("Expected to get the writer")
				return
			}
			w.Write([]byte(want))
			w.Close()

			r, ok := tt.storer.Get(key)
			if !ok {
				t.Error("expected to be available")
				return
			}
			data, err := io.ReadAll(r)
			if err != nil {
				t.Errorf("expected to be available: %s", err)
				return
			}
			if string(data) != want {
				t.Errorf("want %q, got %q", want, data)
				return
			}
			tt.storer.Del(key)
			_, ok = tt.storer.Get(key)
			if ok {
				t.Error("expected to be unavailable")
				return
			}
		})
	}
}
