package message

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDownloadToResumesAfterConnectionReset(t *testing.T) {
	payload := bytes.Repeat([]byte("0123456789abcdef\n"), 100000)
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		start := 0
		rng := r.Header.Get("Range")
		if rng != "" {
			start, _ = strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(rng, "bytes="), "-"))
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)-start))
		if rng != "" {
			w.WriteHeader(http.StatusPartialContent)
		}
		if requests == 1 {
			// Promise the full body, send half, then reset the connection.
			w.Write(payload[start : start+len(payload)/2])
			w.(http.Flusher).Flush()
			panic(http.ErrAbortHandler)
		}
		w.Write(payload[start:])
	}))
	defer srv.Close()

	dir := t.TempDir()
	f, _, err := createPartFile(dir, "2026-10-12", "a/b/data.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := downloadTo(f, srv.URL, "[1/1]"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(f.Name())
	if !bytes.Equal(got, payload) {
		t.Fatalf("content differs: got %d bytes, want %d", len(got), len(payload))
	}
	if requests < 2 {
		t.Fatalf("expected a resumed request, got %d request(s)", requests)
	}
}

func TestDownloadToRetriesAfterStalledBody(t *testing.T) {
	old := idleTimeout
	idleTimeout = 300 * time.Millisecond
	defer func() { idleTimeout = old }()

	payload := bytes.Repeat([]byte("0123456789abcdef\n"), 1000)
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		start := 0
		rng := r.Header.Get("Range")
		if rng != "" {
			start, _ = strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(rng, "bytes="), "-"))
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)-start))
		if rng != "" {
			w.WriteHeader(http.StatusPartialContent)
		}
		if requests == 1 {
			// Send half, then stall without closing the connection.
			w.Write(payload[start : start+len(payload)/2])
			w.(http.Flusher).Flush()
			<-r.Context().Done()
			return
		}
		w.Write(payload[start:])
	}))
	defer srv.Close()

	f, _, err := createPartFile(t.TempDir(), "2026-10-12", "data.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := downloadTo(f, srv.URL, "[1/1]"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(f.Name())
	if !bytes.Equal(got, payload) {
		t.Fatalf("content differs: got %d bytes, want %d", len(got), len(payload))
	}
	if requests != 2 {
		t.Fatalf("expected one retry after the stall, got %d request(s)", requests)
	}
}

func TestDownloadAndParseBlobReplacesOldFile(t *testing.T) {
	line := `{"event":"e","time":"2026-10-12T10:00:00Z","data":{"topic":"data/t","origin":"o","format":"signal","fields":{"a":{"value":"1","type":"int"}}}}` + "\n"
	payload := []byte(line + line)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(payload)
	}))
	defer srv.Close()

	dir := t.TempDir()
	final := filepath.Join(dir, "2026-10-12_part.json")
	if err := os.WriteFile(final, []byte("old content that is longer than the new one\n"), 0644); err != nil {
		t.Fatal(err)
	}

	file, lines, err := downloadAndParseBlob("2026-10-12", "x/y/part.json", srv.URL, dir, nil, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if file != final || lines != 2 {
		t.Fatalf("got %s with %d lines", file, lines)
	}
	got, _ := os.ReadFile(final)
	if !bytes.Equal(got, payload) {
		t.Fatalf("file not replaced: %q", got)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 { // the json file and signal.csv, no .part file
		t.Fatalf("expected the json and csv file, found %d entries", len(entries))
	}
}
