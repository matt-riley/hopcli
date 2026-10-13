package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matryer/is"
)

// slowBodyHandler flushes the response headers and only then writes the body,
// with a delay in between. Real servers (and Cloudflare in front of the Store
// API) do exactly this, so the client observes the headers well before the
// body has been read.
type slowBodyHandler struct {
	body  string
	delay time.Duration
}

func (h *slowBodyHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	time.Sleep(h.delay)
	_, _ = io.WriteString(w, h.body)
}

// TestDoRequest_BodyReadSurvivesRequestCancel is a regression test for the
// demo gif showing "Error: context canceled" for requests that were perfectly
// healthy.
//
// doRequest used to cancel the per-request context as soon as Do returned —
// i.e. before reading the body. Cancelling there aborts the in-flight body
// read, so a slow (but successful) response was reported as
// "context canceled" to the UI.
//
// HTTP/2 is forced because the Store API is served over HTTP/2, where a
// cancelled context reliably fails the in-flight body read.
func TestDoRequest_BodyReadSurvivesRequestCancel(t *testing.T) {
	is := is.New(t)

	const want = `{"name":"hopcli"}`
	server := httptest.NewUnstartedServer(&slowBodyHandler{body: want, delay: 250 * time.Millisecond})
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())

	c := &HTTPClient{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:   &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12},
				ForceAttemptHTTP2: true,
			},
		},
	}

	resp, body, err := c.doRequest(context.Background(), server.URL)
	is.NoErr(err) // a slow body must not be reported as a cancelled request
	is.Equal(resp.StatusCode, http.StatusOK)
	is.Equal(string(body), want)
}
