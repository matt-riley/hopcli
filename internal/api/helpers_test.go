package api

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
)

// fakeNetError is a minimal net.Error implementation used to exercise
// isRetryableError's net.Error and *url.Error branches without depending on
// real network conditions. It satisfies the net.Error interface
// (Error/Timeout/Temporary) by value, so errors.As succeeds.
type fakeNetError struct {
	timeout   bool
	temporary bool
	msg       string
}

func (e fakeNetError) Error() string   { return e.msg }
func (e fakeNetError) Timeout() bool    { return e.timeout }
func (e fakeNetError) Temporary() bool  { return e.temporary }

// TestIsRetryableError exercises every branch of isRetryableError directly:
// nil input, non-retryable context errors, retryable net.Errors, plain
// errors, and *url.Error wrapping (which recurses into its inner Err).
func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"context canceled", context.Canceled, false},
		{"context deadline exceeded", context.DeadlineExceeded, false},
		{"net timeout error", fakeNetError{timeout: true, temporary: false, msg: "i/o timeout"}, true},
		{"net temporary error", fakeNetError{timeout: false, temporary: true, msg: "connection reset"}, true},
		{"plain non-network error", errors.New("something broke"), false},
		{"url.Error wrapping retryable net error",
			&url.Error{Op: "Get", URL: "http://x", Err: fakeNetError{timeout: true, msg: "timeout"}}, true},
		{"url.Error wrapping plain error",
			&url.Error{Op: "Get", URL: "http://x", Err: errors.New("connection refused")}, false},
		{"url.Error wrapping context canceled",
			&url.Error{Op: "Get", URL: "http://x", Err: context.Canceled}, false},
		{"nested url.Error wrapping retryable net error",
			&url.Error{Op: "Get", URL: "http://x",
				Err: &url.Error{Op: "Get", URL: "http://y", Err: fakeNetError{temporary: true, msg: "temp"}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			is.Equal(isRetryableError(tt.err), tt.want)
		})
	}
}

// TestRetryConfig_maxRetries covers all three branches: nil config (1),
// non-positive MaxRetries (defaults to 3), and an explicit positive value.
func TestRetryConfig_maxRetries(t *testing.T) {
	is := is.New(t)
	is.Equal((*RetryConfig)(nil).maxRetries(), 1)
	is.Equal((&RetryConfig{}).maxRetries(), 3)
	is.Equal((&RetryConfig{MaxRetries: 5}).maxRetries(), 5)
}

// TestRetryConfig_backoffBase covers nil config, zero/unset value (default 1s),
// and an explicit value.
func TestRetryConfig_backoffBase(t *testing.T) {
	is := is.New(t)
	is.Equal((*RetryConfig)(nil).backoffBase(), time.Second)
	is.Equal((&RetryConfig{}).backoffBase(), time.Second)
	is.Equal((&RetryConfig{BackoffBase: 100 * time.Millisecond}).backoffBase(), 100*time.Millisecond)
}

// TestRetryConfig_maxDuration covers nil config, zero/unset value (default 15s),
// and an explicit value.
func TestRetryConfig_maxDuration(t *testing.T) {
	is := is.New(t)
	is.Equal((*RetryConfig)(nil).maxDuration(), 15*time.Second)
	is.Equal((&RetryConfig{}).maxDuration(), 15*time.Second)
	is.Equal((&RetryConfig{MaxDuration: 30 * time.Second}).maxDuration(), 30*time.Second)
}

// TestRetryConfig_isRetryableStatusCode covers the default path (nil config and
// nil RetryableStatusCodes → 5xx + 429) and the custom-list path (match,
// no-match, and a code that would be retryable by default but is excluded by
// the custom list).
func TestRetryConfig_isRetryableStatusCode(t *testing.T) {
	is := is.New(t)
	// nil rc → default behaviour.
	is.Equal((*RetryConfig)(nil).isRetryableStatusCode(500), true)
	is.Equal((*RetryConfig)(nil).isRetryableStatusCode(503), true)
	is.Equal((*RetryConfig)(nil).isRetryableStatusCode(429), true)
	is.Equal((*RetryConfig)(nil).isRetryableStatusCode(404), false)
	is.Equal((*RetryConfig)(nil).isRetryableStatusCode(400), false)

	// Empty RetryConfig (nil list) → default behaviour.
	is.Equal((&RetryConfig{}).isRetryableStatusCode(502), true)
	is.Equal((&RetryConfig{}).isRetryableStatusCode(200), false)

	// Custom list: only listed codes are retryable.
	custom := &RetryConfig{RetryableStatusCodes: []int{418, 502}}
	is.Equal(custom.isRetryableStatusCode(418), true)
	is.Equal(custom.isRetryableStatusCode(502), true)
	is.Equal(custom.isRetryableStatusCode(500), false) // 500 is retryable by default but excluded by the custom list
	is.Equal(custom.isRetryableStatusCode(429), false)  // same — excluded by custom list
	is.Equal(custom.isRetryableStatusCode(404), false)
}

// TestBodyPreview covers the short-body, empty-body, exact-boundary, and
// over-boundary (truncation) branches.
func TestBodyPreview(t *testing.T) {
	is := is.New(t)
	is.Equal(bodyPreview(nil), "")
	is.Equal(bodyPreview([]byte("short")), "short")
	is.Equal(bodyPreview([]byte(strings.Repeat("x", 200))), strings.Repeat("x", 200)) // exact boundary → no truncation

	long := strings.Repeat("y", 250)
	got := bodyPreview([]byte(long))
	is.Equal(len(got), 203) // 200 bytes + "..."
	is.True(strings.HasSuffix(got, "..."))
	is.Equal(got[:200], strings.Repeat("y", 200))
}

// TestHTTPClient_timeout covers the zero-Timeout (default) and positive branches.
func TestHTTPClient_timeout(t *testing.T) {
	is := is.New(t)
	is.Equal((&HTTPClient{}).timeout(), DefaultHTTPTimeout)
	is.Equal((&HTTPClient{Timeout: 7 * time.Second}).timeout(), 7*time.Second)
}

// TestHTTPClient_httpClient covers the nil (→ http.DefaultClient) and set branches.
func TestHTTPClient_httpClient(t *testing.T) {
	is := is.New(t)
	is.Equal((&HTTPClient{}).httpClient(), http.DefaultClient)
	custom := &http.Client{Timeout: 3 * time.Second}
	is.Equal((&HTTPClient{HTTPClient: custom}).httpClient(), custom)
}
