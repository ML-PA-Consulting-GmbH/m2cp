package message

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	maxDownloadAttempts = 8
	headerTimeout       = 30 * time.Second
)

// idleTimeout: no bytes received for this long aborts the attempt. A variable so tests can shorten it.
var idleTimeout = 30 * time.Second

var downloadClient = &http.Client{
	Transport: func() http.RoundTripper {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.ResponseHeaderTimeout = headerTimeout
		return t
	}(),
}

// createPartFile opens <dir>/<day>_<blob base name>.part for writing and returns it with the final path.
// The final file is only replaced once the download is complete.
func createPartFile(dir, day, blobName string) (*os.File, string, error) {
	name := path.Base(strings.ReplaceAll(blobName, "\\", "/"))
	if name == "" || name == "." || name == "/" {
		name = "data.json"
	}
	final := filepath.Join(dir, day+"_"+name)
	f, err := os.OpenFile(final+".part", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("could not create file: %w", err)
	}
	return f, final, nil
}

type progressWriter struct {
	w        io.Writer
	label    string
	base     int64 // bytes already in the file before this attempt
	total    int64 // 0 if unknown
	written  int64
	start    time.Time
	lastShow time.Time
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	p.written += int64(n)
	if now := time.Now(); now.Sub(p.lastShow) >= 500*time.Millisecond {
		p.lastShow = now
		done := p.base + p.written
		speed := float64(p.written) / 1024 / 1024 / now.Sub(p.start).Seconds()
		if p.total > 0 {
			fmt.Printf("\r%s Downloading... %.2f%% (%.2f / %.2f MB) @ %.2f MB/s",
				p.label, 100*float64(done)/float64(p.total), float64(done)/1024/1024, float64(p.total)/1024/1024, speed)
		} else {
			fmt.Printf("\r%s Downloading... %.2f MB @ %.2f MB/s", p.label, float64(done)/1024/1024, speed)
		}
	}
	return n, err
}

// downloadTo streams uri into f. After a connection error it resumes with a Range request.
// Blobs of the current day are append-only, so a resumed download stays byte-exact.
func downloadTo(f *os.File, uri, label string) error {
	var lastErr error
	for attempt := 1; attempt <= maxDownloadAttempts; attempt++ {
		if attempt > 1 {
			fmt.Printf("\r%-80s\r%s Retrying (%d/%d) after: %v\n", "", label, attempt, maxDownloadAttempts, lastErr)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		retry, err := downloadAttempt(f, uri, label)
		if err == nil {
			return nil
		}
		if !retry {
			return err
		}
		lastErr = err
	}
	return fmt.Errorf("failed to download blob after %d attempts: %w", maxDownloadAttempts, lastErr)
}

// idleReader resets timer whenever bytes arrive.
type idleReader struct {
	r     io.Reader
	timer *time.Timer
}

func (i *idleReader) Read(b []byte) (int, error) {
	n, err := i.r.Read(b)
	if n > 0 {
		i.timer.Reset(idleTimeout)
	}
	return n, err
}

// downloadAttempt makes one request and appends the body to f. The bool tells if a failure is worth a retry.
func downloadAttempt(f *os.File, uri, label string) (retry bool, err error) {
	offset, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return false, fmt.Errorf("could not seek output file: %w", err)
	}

	// A stalled connection cancels the request when no bytes arrive for idleTimeout.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	idle := time.AfterFunc(idleTimeout, cancel)
	defer idle.Stop()
	stalled := func(err error) error {
		if ctx.Err() != nil {
			return fmt.Errorf("no data received for %s", idleTimeout)
		}
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create download request: %w", err)
	}
	if offset > 0 {
		req.Header.Set("Range", "bytes="+strconv.FormatInt(offset, 10)+"-")
	}

	resp, err := downloadClient.Do(req)
	if err != nil {
		return true, stalled(err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusRequestedRangeNotSatisfiable:
		return false, nil // everything is already on disk
	case resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests:
		return true, fmt.Errorf("server returned %s", resp.Status)
	case resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent:
		return false, fmt.Errorf("failed to download blob: server returned %s", resp.Status)
	}

	if resp.StatusCode == http.StatusOK && offset > 0 {
		// Server ignored the Range header: start again.
		if err := f.Truncate(0); err != nil {
			return false, fmt.Errorf("could not truncate output file: %w", err)
		}
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return false, fmt.Errorf("could not seek output file: %w", err)
		}
		offset = 0
	}

	var total int64
	if resp.ContentLength > 0 {
		total = offset + resp.ContentLength
	}
	idle.Reset(idleTimeout)
	pw := &progressWriter{w: f, label: label, base: offset, total: total, start: time.Now()}
	if _, err := io.Copy(pw, &idleReader{r: resp.Body, timer: idle}); err != nil {
		return true, stalled(err)
	}

	size, _ := f.Seek(0, io.SeekEnd)
	fmt.Printf("\r%-80s\r%s Download complete: %.2f MB\n", "", label, float64(size)/1024/1024)
	return false, nil
}
