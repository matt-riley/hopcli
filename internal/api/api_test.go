package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/internal/api"
)

func TestFetchProducts_HappyPath(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.Equal(r.URL.Path, "/wp-json/wc/store/v1/products")
		is.Equal(r.URL.Query().Get("page"), "2")
		is.Equal(r.URL.Query().Get("per_page"), "5")
		is.Equal(r.URL.Query().Get("orderby"), "date")
		is.Equal(r.URL.Query().Get("order"), "desc")

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-WP-Total", "50")
		w.Header().Set("X-WP-TotalPages", "10")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`[{"id":1,"name":"Test Beer","description":"Desc","short_description":"Short","permalink":"https://example.com/beer","prices":{"price":"420","regular_price":"420","sale_price":"","currency_code":"GBP","currency_symbol":"£","currency_minor_unit":2,"currency_prefix":"£","currency_suffix":""}}]`))
		is.NoErr(err)
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	products, pagination, err := client.FetchProducts(context.Background(), 2, 5)

	is.NoErr(err)
	is.Equal(len(products), 1)
	is.Equal(products[0].ID, 1)
	is.Equal(products[0].Title, "Test Beer")
	is.Equal(pagination.TotalItems, 50)
	is.Equal(pagination.TotalPages, 10)
}

func TestFetchProducts_ErrorStatus(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, _, err := client.FetchProducts(context.Background(), 1, 10)

	is.True(err != nil)
}

func TestFetchProducts_DefaultPageAndPerPage(t *testing.T) {
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
	_, _, err := client.FetchProducts(context.Background(), 0, 0)

	is.NoErr(err)
}

func TestFetchProducts_MissingPaginationHeaders(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, pagination, err := client.FetchProducts(context.Background(), 1, 10)

	is.NoErr(err)
	is.Equal(pagination.TotalItems, 0)
	is.Equal(pagination.TotalPages, 0)
}

func TestFetchProducts_InvalidPaginationHeaders(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-WP-Total", "not-a-number")
		w.Header().Set("X-WP-TotalPages", "also-bad")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, pagination, err := client.FetchProducts(context.Background(), 1, 10)

	is.NoErr(err)
	is.Equal(pagination.TotalItems, 0)
	is.Equal(pagination.TotalPages, 0)
}

func TestFetchProducts_MalformedJSON(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, _, err := client.FetchProducts(context.Background(), 1, 10)

	is.True(err != nil)
}

func TestFetchCategories_HappyPath(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.Equal(r.URL.Path, "/wp-json/wc/store/v1/products/categories")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`[{"id":1,"name":"Category 1","slug":"cat-1","parent":0,"count":5}]`))
		is.NoErr(err)
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	categories, err := client.FetchCategories(context.Background())

	is.NoErr(err)
	is.Equal(len(categories), 1)
	is.Equal(categories[0].ID, 1)
	is.Equal(categories[0].Name, "Category 1")
	is.Equal(categories[0].Slug, "cat-1")
	is.Equal(categories[0].Count, 5)
}

func TestFetchCategories_ErrorStatus(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, err := client.FetchCategories(context.Background())

	is.True(err != nil)
}

func TestFetchProductsByCategory_HappyPath(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.Equal(r.URL.Path, "/wp-json/wc/store/v1/products")
		is.Equal(r.URL.Query().Get("category"), "123")
		is.Equal(r.URL.Query().Get("page"), "3")
		is.Equal(r.URL.Query().Get("per_page"), "8")

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-WP-Total", "24")
		w.Header().Set("X-WP-TotalPages", "3")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`[{"id":42,"name":"Category Beer","description":"Desc","short_description":"Short","permalink":"https://example.com/beer","prices":{"price":"599","regular_price":"599","sale_price":"","currency_code":"GBP","currency_symbol":"£","currency_minor_unit":2,"currency_prefix":"£","currency_suffix":""}}]`))
		is.NoErr(err)
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	products, pagination, err := client.FetchProductsByCategory(context.Background(), 123, 3, 8)

	is.NoErr(err)
	is.Equal(len(products), 1)
	is.Equal(products[0].ID, 42)
	is.Equal(products[0].Title, "Category Beer")
	is.Equal(pagination.TotalItems, 24)
	is.Equal(pagination.TotalPages, 3)
}

func TestFetchProductsByCategory_ErrorStatus(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, _, err := client.FetchProductsByCategory(context.Background(), 1, 1, 10)

	is.True(err != nil)
}

func TestUserAgentHeader(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.Equal(r.Header.Get("User-Agent"), "hopcli")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, _, err := client.FetchProducts(context.Background(), 1, 10)
	is.NoErr(err)
}

func TestEmptyUserAgent(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.True(r.Header.Get("User-Agent") != "hopcli")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	client.UserAgent = ""
	_, _, err := client.FetchProducts(context.Background(), 1, 10)
	is.NoErr(err)
}

func TestCustomHTTPClient(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	client.HTTPClient = &http.Client{}
	_, _, err := client.FetchProducts(context.Background(), 1, 10)
	is.NoErr(err)
}

func TestHTTPClientImplementsInterface(t *testing.T) {
	var _ api.Client = (*api.HTTPClient)(nil)
}

// ---- Robustness tests (moved from commands) ----

func TestContentTypeValidation_NonJSON(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body>Error page</body></html>`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, _, err := client.FetchProducts(context.Background(), 1, 10)

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "Content-Type"))
	is.True(strings.Contains(err.Error(), "text/html"))
}

func TestContentTypeValidation_ApplicationJSON(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, _, err := client.FetchProducts(context.Background(), 1, 10)
	is.NoErr(err)
}

func TestNonRetryable4xx_FailsImmediately(t *testing.T) {
	is := is.New(t)

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, _, err := client.FetchProducts(context.Background(), 1, 10)

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "400"))
	is.True(strings.Contains(err.Error(), "bad request"))
	is.Equal(atomic.LoadInt32(&requestCount), int32(1))
}

func TestNotFound4xx_FailsImmediately(t *testing.T) {
	is := is.New(t)

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, _, err := client.FetchProducts(context.Background(), 1, 10)

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "404"))
	is.Equal(atomic.LoadInt32(&requestCount), int32(1))
}

func TestRetryOn5xx(t *testing.T) {
	is := is.New(t)

	// Use NewHTTPClientWithRetry to get a retry-enabled client.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No counting needed — we just need retries to happen.
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := api.NewHTTPClientWithRetry(server.URL, "hopcli")
	// Speed up for test.
	client.Retry.BackoffBase = 1 * time.Millisecond
	client.Retry.MaxDuration = 200 * time.Millisecond

	_, _, err := client.FetchProducts(context.Background(), 1, 10)

	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "max retries exceeded") || strings.Contains(err.Error(), "retry deadline exceeded"))
}

func TestRetryOn429(t *testing.T) {
	is := is.New(t)

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClientWithRetry(server.URL, "hopcli")
	client.Retry.BackoffBase = 1 * time.Millisecond
	client.Retry.MaxDuration = 5 * time.Second

	_, _, err := client.FetchProducts(context.Background(), 1, 10)
	is.NoErr(err)
	is.Equal(atomic.LoadInt32(&requestCount), int32(3))
}

func TestExhaustsRetriesOn5xx(t *testing.T) {
	is := is.New(t)

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := api.NewHTTPClientWithRetry(server.URL, "hopcli")
	client.Retry.BackoffBase = 1 * time.Millisecond
	client.Retry.MaxDuration = 500 * time.Millisecond

	_, _, err := client.FetchProducts(context.Background(), 1, 10)
	is.True(err != nil)
	is.Equal(atomic.LoadInt32(&requestCount), int32(3))
	is.True(strings.Contains(err.Error(), "max retries exceeded") || strings.Contains(err.Error(), "retry deadline exceeded"))
}

func TestFetchProduct(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.Equal(r.URL.Path, "/wp-json/wc/store/v1/products/42")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"id":42,"name":"Single Beer","description":"A single beer","short_description":"Short","permalink":"https://example.com/beer","prices":{"price":"500","regular_price":"500","sale_price":"","currency_code":"GBP","currency_symbol":"£","currency_minor_unit":2,"currency_prefix":"£","currency_suffix":""}}`))
		is.NoErr(err)
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	product, err := client.FetchProduct(context.Background(), 42)

	is.NoErr(err)
	is.Equal(product.ID, 42)
	is.Equal(product.Title, "Single Beer")
}

// ---- Pagination header tests (moved from commands) ----

func TestPaginationHeaders_BothHeadersPresent(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-WP-Total", "50")
		w.Header().Set("X-WP-TotalPages", "10")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, pagination, err := client.FetchProducts(context.Background(), 1, 10)

	is.NoErr(err)
	is.Equal(pagination.TotalItems, 50)
	is.Equal(pagination.TotalPages, 10)
}

func TestPaginationHeaders_MissingXWPTotal(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-WP-TotalPages", "10")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, pagination, err := client.FetchProducts(context.Background(), 1, 10)

	is.NoErr(err)
	is.Equal(pagination.TotalItems, 0)
	is.Equal(pagination.TotalPages, 10)
}

func TestPaginationHeaders_MissingXWPTotalPages(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-WP-Total", "50")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, pagination, err := client.FetchProducts(context.Background(), 1, 10)

	is.NoErr(err)
	is.Equal(pagination.TotalItems, 50)
	is.Equal(pagination.TotalPages, 0)
}

func TestPaginationHeaders_BothHeadersMissing(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	_, pagination, err := client.FetchProducts(context.Background(), 1, 10)

	is.NoErr(err)
	is.Equal(pagination.TotalItems, 0)
	is.Equal(pagination.TotalPages, 0)
}
