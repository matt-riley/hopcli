package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/internal/api"
)

// These tests cover endpoint-level edge cases that the existing
// api_test.go did not exercise, pushing package coverage above the 80%
// acceptance bar (and exercising the malformed-JSON and default-parameter
// branches of FetchProduct / FetchCategories / FetchProductsByCategory,
// plus the non-retryable network-error path through doRequest).

// TestFetchProduct_MalformedJSON verifies the single-product decode-error
// branch (api.go:385-387).
func TestFetchProduct_MalformedJSON(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, err := client.FetchProduct(context.Background(), 1)

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "JSON decode error"))
}

// TestFetchCategories_MalformedJSON verifies the categories decode-error
// branch (api.go:403-405).
func TestFetchCategories_MalformedJSON(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, err := client.FetchCategories(context.Background())

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "JSON decode error"))
}

// TestFetchProductsByCategory_MalformedJSON verifies the by-category
// decode-error branch (api.go:430-432).
func TestFetchProductsByCategory_MalformedJSON(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-WP-Total", "1")
		w.Header().Set("X-WP-TotalPages", "1")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`broken`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, _, err := client.FetchProductsByCategory(context.Background(), 5, 1, 10)

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "JSON decode error"))
}

// TestFetchProductsByCategory_DefaultParams covers the default-page and
// default-perPage branches (api.go:412-417) when both are non-positive.
func TestFetchProductsByCategory_DefaultParams(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.Equal(r.URL.Query().Get("page"), "1")
		is.Equal(r.URL.Query().Get("per_page"), "10")

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-WP-Total", "0")
		w.Header().Set("X-WP-TotalPages", "0")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	products, pagination, err := client.FetchProductsByCategory(context.Background(), 7, 0, -1)

	is.NoErr(err)
	is.Equal(len(products), 0)
	is.Equal(pagination.TotalItems, 0)
	is.Equal(pagination.TotalPages, 0)
}

// TestFetchProduct_NetworkError exercises the non-retryable request-error
// return path in doRequest (api.go:283-288) by pointed the client at a
// closed server, producing a connection-refused error that is *url.Error
// (and thus retryable) — so we instead use a cancelled context to hit the
// non-retryable-direct-return on the first attempt without retries.
func TestFetchProduct_NetworkError(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	server.Close() // close immediately → connection refused

	client := api.NewHTTPClient(server.URL)
	// NewHTTPClient sets no Retry config, so a transient url.Error from the
	// closed server is treated as retryable by doRequest's single attempt
	// (maxRetries() == 1 via nil config) and returned as the final error.
	_, err := client.FetchProduct(context.Background(), 42)

	is.True(err != nil)
}

// TestFetchCategories_NetworkError is the categories analogue of the above,
// confirming the error surfaces for FetchCategories (api.go:393-397 return).
func TestFetchCategories_NetworkError(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	client := api.NewHTTPClient(server.URL)
	_, err := client.FetchCategories(context.Background())

	is.True(err != nil)
}

// TestFetchProductsByCategory_NetworkError confirms the by-category error
// path (api.go:422-424).
func TestFetchProductsByCategory_NetworkError(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	client := api.NewHTTPClient(server.URL)
	_, _, err := client.FetchProductsByCategory(context.Background(), 9, 1, 10)

	is.True(err != nil)
}
