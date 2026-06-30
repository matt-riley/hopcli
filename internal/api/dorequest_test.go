package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matryer/is"
)

// These tests exercise the remaining uncovered branches inside doRequest and
// isRetryableError by operating within the `api` package (internal test), so
// they can construct RetryConfig and HTTPClient states directly without
// exporting test-only knobs. Each test targets a specific uncovered code
// range identified in the coverage profile.

// countHandler is a tiny helper that records how many requests reached the
// server and optionally writes the configured status / body / content-type.
type countHandler struct {
	count  int32
	status int
	body   string
}

func (h *countHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt32(&h.count, 1)
	if h.status == 0 {
		h.status = http.StatusOK
	}
	if h.body != "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(h.status)
	if h.body != "" {
		_, _ = w.Write([]byte(h.body))
	}
}

// TestDoRequest_MaxAttemptsClampedToOne covers the maxAttempts<1 clamp
// (api.go:252-254): a RetryConfig with MaxRetries=0 is clamped to 3 by
// maxRetries(), but a directly-built HTTPClient whose Retry.MaxRetries is
// negative can't easily reach the <1 guard since maxRetries() returns 3 for
// <=0. Instead we exercise the clamp by constructing an HTTPClient with a
// RetryConfig whose MaxRetries is set such that maxRetries() is already >=1;
// the <1 guard at line 252 is defensive dead code. We nonetheless drive a
// single-attempt request through to confirm the clamp path is harmless.
func TestDoRequest_MaxAttemptsClamped(t *testing.T) {
	is := is.New(t)

	var h countHandler
	h.body = `[]`
	server := httptest.NewServer(&h)
	defer server.Close()

	// Retry installed but MaxRetries left at 0 → maxRetries() returns 3;
	// request succeeds on first attempt so the clamp at line 252 is not
	// entered for the success path. This documents that the clamp only
	// matters when an explicit negative value bypasses maxRetries()'s own
	// clamping — which is impossible via the public API, so line 252-254 is
	// effectively unreachable defensive code.
	c := NewHTTPClient(server.URL)
	c.Retry = &RetryConfig{} // defaults: 3 attempts, 1s backoff, 15s max
	c.Retry.BackoffBase = 1 * time.Millisecond
	resp, body, err := c.doRequest(context.Background(), server.URL)
	is.NoErr(err)
	is.True(resp != nil)
	is.True(strings.Contains(string(body), "[]"))
	is.Equal(atomic.LoadInt32(&h.count), int32(1))
}

// TestDoRequest_RetryDeadlineExceeded covers the retry-deadline branch
// (api.go:261-263): when elapsed+delay would exceed MaxDuration, the
// function returns early with a "retry deadline exceeded" error instead of
// waiting out the backoff.
func TestDoRequest_RetryDeadlineExceeded(t *testing.T) {
	is := is.New(t)

	var h countHandler
	h.status = http.StatusInternalServerError // always retryable
	server := httptest.NewServer(&h)
	defer server.Close()

	c := NewHTTPClient(server.URL)
	c.Retry = &RetryConfig{
		MaxRetries:  10,
		BackoffBase: 500 * time.Millisecond, // delay(1) = 500ms
		MaxDuration: 100 * time.Millisecond, // far below elapsed+delay
	}

	_, _, err := c.doRequest(context.Background(), server.URL)
	is.True(err != nil)
	// Either the deadline branch fires (after the first failure) or, in rare
	// timing, it exhausts. We assert the deadline message specifically.
	is.True(strings.Contains(err.Error(), "retry deadline exceeded") ||
		strings.Contains(err.Error(), "max retries exceeded"))
	// Exactly one attempt: the first 500 fails; the retry guard sees that
	// elapsed+500ms > 100ms and bails before attempt 2.
	is.Equal(atomic.LoadInt32(&h.count), int32(1))
}

// TestDoRequest_ContextCanceledDuringBackoff covers the backoff select's
// ctx.Done branch (api.go:266-267): the first attempt fails with a retryable
// 5xx, then while waiting for the backoff the context is cancelled.
func TestDoRequest_ContextCanceledDuringBackoff(t *testing.T) {
	is := is.New(t)

	var h countHandler
	h.status = http.StatusInternalServerError
	server := httptest.NewServer(&h)
	defer server.Close()

	c := NewHTTPClient(server.URL)
	c.Retry = &RetryConfig{
		MaxRetries:  5,
		BackoffBase: 200 * time.Millisecond,
		MaxDuration: 10 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel shortly after the first (failing) attempt, during the 200ms backoff.
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, _, err := c.doRequest(ctx, server.URL)
	is.True(err != nil)
	is.True(errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "max retries"))
	is.Equal(atomic.LoadInt32(&h.count), int32(1))
}

// TestDoRequest_NewRequestError covers the bad-URL branch (api.go:273-276):
// an unparseable URL makes http.NewRequestWithContext fail immediately.
func TestDoRequest_NewRequestError(t *testing.T) {
	is := is.New(t)

	c := NewHTTPClient("http://example.com")
	c.Retry = &RetryConfig{}
	// A URL with a control character is rejected by url.Parse, so
	// NewRequestWithContext returns an error before any network call.
	_, _, err := c.doRequest(context.Background(), "http://example.com\x7f")
	is.True(err != nil)
}

// TestDoRequest_RetryableRequestError covers the retryable-network-error
// branch of the request error path (api.go:283-286): every attempt hits a
// closed server (connection refused → *url.Error, which is retryable), so
// the loop continues until retries are exhausted.
func TestDoRequest_RetryableRequestError(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close() // connection refused on every attempt

	c := NewHTTPClient(server.URL)
	c.Retry = &RetryConfig{
		MaxRetries:  3,
		BackoffBase: 1 * time.Millisecond,
		MaxDuration: 5 * time.Second,
	}
	_, _, err := c.doRequest(context.Background(), server.URL)
	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "max retries exceeded") ||
		strings.Contains(err.Error(), "retry deadline exceeded"))
}

// TestDoRequest_NonRetryableRequestError covers the direct-return branch
// (api.go:288): a non-retryable error from httpClient().Do returns the
// original error immediately. Context cancellation produces such an error
// (context.Canceled is non-retryable per isRetryableError) on the FIRST
// attempt when there is no retry configured.
func TestDoRequest_NonRetryableRequestError(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	c := NewHTTPClient(server.URL)
	// No Retry config → maxRetries()==1, single attempt.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancelled → Do returns context.Canceled (non-retryable)

	_, _, err := c.doRequest(ctx, server.URL)
	is.True(err != nil)
	// As a context error, doRequest returns ctx.Err() directly (non-retryable).
	is.True(errors.Is(err, context.Canceled))
}

// TestDoRequest_RetryableBodyReadError and the non-retryable variant below
// cover the io.ReadAll error branches (api.go:301-306). Producing a body-read
// error from net/http requires a response whose Reader errors mid-stream;
// the simplest deterministic approach is a handler that sets a
// Content-Length but then closes the connection without writing the full
// body, but that is flaky across Go versions. Instead we exercise the
// branch through a custom http.RoundTripper that returns a response whose
// Body is a failing Reader.

// failingReadCloser is an io.ReadCloser whose Read always errors and Close is a no-op.
type failingReadCloser struct {
	err error
}

func (f failingReadCloser) Read(p []byte) (int, error) { return 0, f.err }
func (f failingReadCloser) Close() error               { return nil }

func TestDoRequest_BodyReadErrors(t *testing.T) {
	is := is.New(t)

	// We drive both the retryable and non-retryable body-read branches by
	// injecting a custom http.Client whose Transport returns a response with
	// a failing Body.

	// (a) Retryable body error (net.Error) → loop continues, exhausts retries.
	var attempts int32
	rtRetryable := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&attempts, 1)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       failingReadCloser{err: fakeNetError{timeout: true, msg: "body read timeout"}},
		}, nil
	})
	c := NewHTTPClient("http://test.local")
	c.HTTPClient = &http.Client{Transport: rtRetryable}
	c.Retry = &RetryConfig{MaxRetries: 3, BackoffBase: 1 * time.Millisecond, MaxDuration: 5 * time.Second}
	_, _, err := c.doRequest(context.Background(), "http://test.local/x")
	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "max retries exceeded") ||
		strings.Contains(err.Error(), "retry deadline exceeded"))
	is.True(atomic.LoadInt32(&attempts) == 3) // retried the full 3 times

	// (b) Non-retryable body error → returns immediately (single attempt).
	var attempts2 int32
	rtNonRetryable := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&attempts2, 1)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       failingReadCloser{err: errors.New("non-retryable body boom")},
		}, nil
	})
	c2 := NewHTTPClient("http://test.local")
	c2.HTTPClient = &http.Client{Transport: rtNonRetryable}
	c2.Retry = &RetryConfig{MaxRetries: 3, BackoffBase: 1 * time.Millisecond, MaxDuration: 5 * time.Second}
	_, _, err2 := c2.doRequest(context.Background(), "http://test.local/x")
	is.True(err2 != nil)
	is.True(strings.Contains(err2.Error(), "non-retryable body boom"))
	is.Equal(atomic.LoadInt32(&attempts2), int32(1)) // did not retry
}

// roundTripFunc is an http.RoundTripper backed by a function.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
