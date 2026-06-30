// Package api provides a pure HTTP client for The Hoptimist WooCommerce Store API.
// It has no dependency on Bubble Tea or any TUI framework, making it independently
// testable and reusable.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ProductPrices holds pricing information from the WooCommerce Store API.
type ProductPrices struct {
	Price             string `json:"price"`
	RegularPrice      string `json:"regular_price"`
	SalePrice         string `json:"sale_price"`
	CurrencyCode      string `json:"currency_code"`
	CurrencySymbol    string `json:"currency_symbol"`
	CurrencyMinorUnit int    `json:"currency_minor_unit"`
	CurrencyPrefix    string `json:"currency_prefix"`
	CurrencySuffix    string `json:"currency_suffix"`
}

// Product represents a product from the Store API.
//
// Known limitation: the upstream WooCommerce Store API does not expose brewery
// or brand data for products. There is no Brewery() method — the product title
// (Name field) is the only identifying text available. If brewery/brand data
// becomes available through the API in the future, add a Brewery field here and
// thread it through the product.ProductModel for display.
type Product struct {
	ID               int           `json:"id"`
	Link             string        `json:"permalink"`
	Title            string        `json:"name"`
	Description      string        `json:"description"`
	ShortDescription string        `json:"short_description"`
	Prices           ProductPrices `json:"prices"`
	OnSale           bool          `json:"on_sale"`
	IsInStock        bool          `json:"is_in_stock"`
}

// Products is a slice of Product.
type Products []Product

// Category models a product category from the WooCommerce Store API.
type Category struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Parent int    `json:"parent"`
	Count  int    `json:"count"`
}

// Categories is a slice of Category.
type Categories []Category

// Pagination holds the pagination metadata returned by API response headers.
type Pagination struct {
	TotalItems int
	TotalPages int
}

// Client defines the interface for fetching data from The Hoptimist API.
// This interface exists to enable easy mocking in tests and to decouple
// HTTP logic from TUI concerns.
type Client interface {
	// FetchProducts retrieves products ordered by date descending, with pagination.
	FetchProducts(ctx context.Context, page, perPage int) (Products, Pagination, error)

	// FetchCategories retrieves all product categories.
	FetchCategories(ctx context.Context) (Categories, error)

	// FetchProduct retrieves a single product by its ID.
	FetchProduct(ctx context.Context, productID int) (Product, error)

	// FetchProductsByCategory retrieves products for a specific category, with pagination.
	FetchProductsByCategory(ctx context.Context, categoryID, page, perPage int) (Products, Pagination, error)
}

// RetryConfig controls HTTP request retry behaviour.
// A nil RetryConfig means no retries are attempted.
type RetryConfig struct {
	// MaxRetries is the maximum number of HTTP request attempts (including the initial try).
	// When zero, defaults to 3.
	MaxRetries int

	// BackoffBase is the base duration for exponential backoff.
	// When zero, defaults to 1 second.
	BackoffBase time.Duration

	// MaxDuration is the maximum total wall-clock time spent retrying before giving up.
	// When zero, defaults to 15 seconds.
	MaxDuration time.Duration

	// RetryableStatusCodes contains HTTP status codes that warrant a retry.
	// When nil, defaults to 5xx and 429.
	RetryableStatusCodes []int
}

func (rc *RetryConfig) maxRetries() int {
	if rc == nil {
		return 1 // no retries when RetryConfig is nil
	}
	if rc.MaxRetries <= 0 {
		return 3
	}
	return rc.MaxRetries
}

func (rc *RetryConfig) backoffBase() time.Duration {
	if rc == nil || rc.BackoffBase <= 0 {
		return time.Second
	}
	return rc.BackoffBase
}

func (rc *RetryConfig) maxDuration() time.Duration {
	if rc == nil || rc.MaxDuration <= 0 {
		return 15 * time.Second
	}
	return rc.MaxDuration
}

func (rc *RetryConfig) isRetryableStatusCode(code int) bool {
	if rc != nil && rc.RetryableStatusCodes != nil {
		for _, s := range rc.RetryableStatusCodes {
			if s == code {
				return true
			}
		}
		return false
	}
	// Default: 5xx and 429.
	return code >= 500 || code == 429
}

// isRetryableError returns true for errors that are likely transient and worth retrying.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	// Context cancellation / deadline exceeded are not retryable.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	// net.Error covers timeouts and temporary network failures.
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	// url.Error wraps lower-level errors (DNS, connection refused, etc.).
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return isRetryableError(urlErr.Err)
	}
	return false
}

// bodyPreview returns a truncated preview of body bytes for error messages.
func bodyPreview(body []byte) string {
	const maxLen = 200
	if len(body) <= maxLen {
		return string(body)
	}
	return string(body[:maxLen]) + "..."
}

// DefaultHTTPTimeout is the default timeout for HTTP requests.
const DefaultHTTPTimeout = 5 * time.Second

// DefaultPerPage is the default number of items per page.
const DefaultPerPage = 10

// DefaultUserAgent is the User-Agent header sent with API requests.
const DefaultUserAgent = "hopcli"

// HTTPClient is the default implementation of Client that makes real HTTP calls.
type HTTPClient struct {
	// BaseURL is the root URL of the API (e.g. "https://thehoptimist.co.uk").
	BaseURL string

	// HTTPClient is the underlying HTTP client. When nil, http.DefaultClient is used.
	HTTPClient *http.Client

	// UserAgent is the User-Agent header sent with requests.
	UserAgent string

	// Timeout is the per-request timeout. When zero, DefaultHTTPTimeout is used.
	Timeout time.Duration

	// Retry is the retry configuration. When nil, no retries are attempted.
	Retry *RetryConfig
}

// NewHTTPClient creates a new HTTPClient with sensible defaults (no retry).
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL:   baseURL,
		UserAgent: DefaultUserAgent,
		Timeout:   DefaultHTTPTimeout,
	}
}

// NewHTTPClientWithRetry creates a new HTTPClient with default retry settings
// (3 retries, 1s backoff, 15s max duration, 5xx+429 retryable).
func NewHTTPClientWithRetry(baseURL, userAgent string) *HTTPClient {
	return &HTTPClient{
		BaseURL:   baseURL,
		UserAgent: userAgent,
		Timeout:   DefaultHTTPTimeout,
		Retry:     &RetryConfig{},
	}
}

// NewClient creates a new HTTP API client that satisfies the Client interface
// with sensible defaults (no retry). It is the canonical constructor for
// consumers that only depend on the Client interface.
func NewClient(baseURL string) Client {
	return NewHTTPClient(baseURL)
}

func (c *HTTPClient) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *HTTPClient) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return DefaultHTTPTimeout
}

// doRequest executes an HTTP GET request with:
//   - User-Agent header set
//   - Exponential backoff retry on transient errors (network, 5xx, 429)
//   - Non-retryable 4xx (except 429) fails immediately
//   - Content-Type validated as application/json before returning
//   - Total retry time bounded per RetryConfig
//
// Returns the response (with Body already read and closed — use headers only),
// the full body bytes, and any error.
func (c *HTTPClient) doRequest(ctx context.Context, urlStr string) (*http.Response, []byte, error) {
	rc := c.Retry
	start := time.Now()
	var lastErr error

	maxAttempts := rc.maxRetries()
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Apply backoff before retries (not before the first attempt).
		if attempt > 0 {
			delay := time.Duration(1<<uint(attempt-1)) * rc.backoffBase()
			elapsed := time.Since(start)
			if elapsed+delay > rc.maxDuration() {
				return nil, nil, fmt.Errorf("retry deadline exceeded after %v (last error: %w)", elapsed, lastErr)
			}
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			}
		}

		reqCtx, cancel := context.WithTimeout(ctx, c.timeout())
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, urlStr, nil)
		if err != nil {
			cancel()
			return nil, nil, err
		}
		if c.UserAgent != "" {
			req.Header.Set("User-Agent", c.UserAgent)
		}

		resp, err := c.httpClient().Do(req)
		cancel()
		if err != nil {
			if isRetryableError(err) {
				lastErr = err
				continue
			}
			return nil, nil, err
		}

		// Retry on transient status codes.
		if rc.isRetryableStatusCode(resp.StatusCode) {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("server returned status %d", resp.StatusCode)
			continue
		}

		// Read the full body.
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			if isRetryableError(err) {
				lastErr = err
				continue
			}
			return nil, nil, err
		}

		// Non-retryable 4xx: fail immediately with a clear error.
		if resp.StatusCode >= 400 {
			return nil, body, fmt.Errorf("HTTP %d: %s", resp.StatusCode, bodyPreview(body))
		}

		// Validate Content-Type is JSON for 2xx responses.
		ct := resp.Header.Get("Content-Type")
		if ct != "" && !strings.HasPrefix(ct, "application/json") {
			return nil, body, fmt.Errorf("unexpected Content-Type %q (status %d): %s", ct, resp.StatusCode, bodyPreview(body))
		}

		return resp, body, nil
	}

	return nil, nil, fmt.Errorf("max retries exceeded (%d attempts): %w", maxAttempts, lastErr)
}

// parsePaginationHeaders extracts X-WP-Total and X-WP-TotalPages from response headers.
func parsePaginationHeaders(resp *http.Response) Pagination {
	var p Pagination

	totalItemsStr := resp.Header.Get("X-WP-Total")
	if totalItemsStr != "" {
		if n, err := strconv.Atoi(totalItemsStr); err == nil {
			p.TotalItems = n
		}
	}

	totalPagesStr := resp.Header.Get("X-WP-TotalPages")
	if totalPagesStr != "" {
		if n, err := strconv.Atoi(totalPagesStr); err == nil {
			p.TotalPages = n
		}
	}

	return p
}

// FetchProducts retrieves products ordered by date descending, with pagination.
func (c *HTTPClient) FetchProducts(ctx context.Context, page, perPage int) (Products, Pagination, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = DefaultPerPage
	}

	urlStr := fmt.Sprintf("%s/wp-json/wc/store/v1/products?orderby=date&order=desc&page=%d&per_page=%d",
		c.BaseURL, page, perPage)

	resp, body, err := c.doRequest(ctx, urlStr)
	if err != nil {
		return nil, Pagination{}, err
	}

	pagination := parsePaginationHeaders(resp)

	var products Products
	if err := json.Unmarshal(body, &products); err != nil {
		return nil, Pagination{}, fmt.Errorf("JSON decode error: %w (body: %s)", err, bodyPreview(body))
	}

	return products, pagination, nil
}

// FetchProduct retrieves a single product by its ID.
func (c *HTTPClient) FetchProduct(ctx context.Context, productID int) (Product, error) {
	urlStr := fmt.Sprintf("%s/wp-json/wc/store/v1/products/%d", c.BaseURL, productID)

	resp, body, err := c.doRequest(ctx, urlStr)
	if err != nil {
		return Product{}, err
	}
	_ = resp // response headers available if needed

	var product Product
	if err := json.Unmarshal(body, &product); err != nil {
		return Product{}, fmt.Errorf("JSON decode error: %w (body: %s)", err, bodyPreview(body))
	}

	return product, nil
}

// FetchCategories retrieves all product categories.
func (c *HTTPClient) FetchCategories(ctx context.Context) (Categories, error) {
	urlStr := fmt.Sprintf("%s/wp-json/wc/store/v1/products/categories", c.BaseURL)

	resp, body, err := c.doRequest(ctx, urlStr)
	if err != nil {
		return nil, err
	}
	_ = resp

	var categories Categories
	if err := json.Unmarshal(body, &categories); err != nil {
		return nil, fmt.Errorf("JSON decode error: %w (body: %s)", err, bodyPreview(body))
	}

	return categories, nil
}

// FetchProductsByCategory retrieves products for a specific category, with pagination.
func (c *HTTPClient) FetchProductsByCategory(ctx context.Context, categoryID, page, perPage int) (Products, Pagination, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = DefaultPerPage
	}

	urlStr := fmt.Sprintf("%s/wp-json/wc/store/v1/products?category=%d&page=%d&per_page=%d",
		c.BaseURL, categoryID, page, perPage)

	resp, body, err := c.doRequest(ctx, urlStr)
	if err != nil {
		return nil, Pagination{}, err
	}

	pagination := parsePaginationHeaders(resp)

	var products Products
	if err := json.Unmarshal(body, &products); err != nil {
		return nil, Pagination{}, fmt.Errorf("JSON decode error: %w (body: %s)", err, bodyPreview(body))
	}

	return products, pagination, nil
}
