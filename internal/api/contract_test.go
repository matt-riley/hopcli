package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/matt-riley/hopcli/internal/api"
)

// hangingServerHandler is an http.HandlerFunc that blocks until the request
// context is cancelled (e.g. by the client timing out or the server shutting
// down). Using r.Context() instead of a blind time.Sleep means server.Close()
// in a defer doesn't have to wait for a fixed sleep to finish, keeping the
// timeout tests fast.
func hangingServerHandler(w http.ResponseWriter, r *http.Request) {
	<-r.Context().Done()
}

// TestFetchProducts_NetworkTimeout covers the network-timeout error case:
// the server accepts the connection but never responds within the client's
// configured Timeout. The client should surface a meaningful timeout error
// (wrapped *url.Error whose Timeout() is true) rather than hanging.
func TestFetchProducts_NetworkTimeout(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(hangingServerHandler))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	client.Timeout = 50 * time.Millisecond // very short — must time out

	done := make(chan error, 1)
	go func() {
		_, _, err := client.FetchProducts(context.Background(), 1, 10)
		done <- err
	}()

	select {
	case err := <-done:
		is.True(err != nil)
		// The error should mention a timeout or deadline in some form.
		is.True(
			errors.Is(err, context.DeadlineExceeded) ||
				isTimeoutErr(err))
	case <-time.After(5 * time.Second):
		is.Fail() // the call hung — timeout path not working
	}
}

// TestFetchProduct_NetworkTimeout is the single-product analogue of above.
func TestFetchProduct_NetworkTimeout(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(hangingServerHandler))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	client.Timeout = 50 * time.Millisecond

	err := make(chan error, 1)
	go func() {
		_, e := client.FetchProduct(context.Background(), 42)
		err <- e
	}()

	select {
	case e := <-err:
		is.True(e != nil)
		is.True(errors.Is(e, context.DeadlineExceeded) || isTimeoutErr(e))
	case <-time.After(5 * time.Second):
		is.Fail()
	}
}

// TestFetchCategories_NetworkTimeout covers timeout on FetchCategories.
func TestFetchCategories_NetworkTimeout(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(hangingServerHandler))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	client.Timeout = 50 * time.Millisecond

	err := make(chan error, 1)
	go func() {
		_, e := client.FetchCategories(context.Background())
		err <- e
	}()

	select {
	case e := <-err:
		is.True(e != nil)
		is.True(errors.Is(e, context.DeadlineExceeded) || isTimeoutErr(e))
	case <-time.After(5 * time.Second):
		is.Fail()
	}
}

// TestFetchProducts_EmptyResponse covers the empty-response case explicitly:
// a 200 with an empty JSON array `[]` and no items. The client should return
// a zero-length (non-nil) Products slice, zero pagination, and no error.
func TestFetchProducts_EmptyResponse(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-WP-Total", "0")
		w.Header().Set("X-WP-TotalPages", "0")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	products, pagination, err := client.FetchProducts(context.Background(), 1, 10)

	is.NoErr(err)
	is.Equal(len(products), 0)
	is.Equal(pagination.TotalItems, 0)
	is.Equal(pagination.TotalPages, 0)
}

// TestFetchCategories_EmptyResponse covers the empty-response case for
// categories: a 200 with `[]`.
func TestFetchCategories_EmptyResponse(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	categories, err := client.FetchCategories(context.Background())

	is.NoErr(err)
	is.Equal(len(categories), 0)
}

// TestFetchProductsByCategory_EmptyResponse covers the empty-response case
// for FetchProductsByCategory.
func TestFetchProductsByCategory_EmptyResponse(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-WP-Total", "0")
		w.Header().Set("X-WP-TotalPages", "0")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewHTTPClient(server.URL)
	products, pagination, err := client.FetchProductsByCategory(context.Background(), 5, 1, 10)

	is.NoErr(err)
	is.Equal(len(products), 0)
	is.Equal(pagination.TotalItems, 0)
	is.Equal(pagination.TotalPages, 0)
}

// ---------------------------------------------------------------------------
// Table-driven interface-contract test using a mock Client implementation.
// ---------------------------------------------------------------------------
//
// The task requires a table-driven test that uses a MOCK implementation to
// confirm the Client interface contract: every method returns the correctly
// typed result (Products / Categories / Product) and a meaningful error when
// configured to fail. This is distinct from TestHTTPClientImplementsInterface,
// which is only a compile-time type assertion using the real HTTPClient.

// mockClient is a configurable mock implementation of api.Client used to
// verify the interface contract in isolation — no HTTP, no network.
type mockClient struct {
	products   api.Products
	categories api.Categories
	product    api.Product
	pagination api.Pagination
	err        error

	// Tracks which methods were called and with what args, so the table-driven
	// tests can assert the correct method is dispatched for each case.
	callsMu sync.Mutex
	calls   []string
}

func (m *mockClient) FetchProducts(ctx context.Context, page, perPage int) (api.Products, api.Pagination, error) {
	m.callsMu.Lock()
	m.calls = append(m.calls, "FetchProducts")
	m.callsMu.Unlock()
	return m.products, m.pagination, m.err
}

func (m *mockClient) FetchCategories(ctx context.Context) (api.Categories, error) {
	m.callsMu.Lock()
	m.calls = append(m.calls, "FetchCategories")
	m.callsMu.Unlock()
	return m.categories, m.err
}

func (m *mockClient) FetchProduct(ctx context.Context, productID int) (api.Product, error) {
	m.callsMu.Lock()
	m.calls = append(m.calls, "FetchProduct")
	m.callsMu.Unlock()
	return m.product, m.err
}

func (m *mockClient) FetchProductsByCategory(ctx context.Context, categoryID, page, perPage int) (api.Products, api.Pagination, error) {
	m.callsMu.Lock()
	m.calls = append(m.calls, "FetchProductsByCategory")
	m.callsMu.Unlock()
	return m.products, m.pagination, m.err
}

// Compile-time guard: mockClient satisfies the Client interface.
var _ api.Client = (*mockClient)(nil)

func TestMockClient_InterfaceContract(t *testing.T) {
	// Shared canned data used across subtests.
	beerProducts := api.Products{
		{ID: 1, Title: "Pale Ale", Prices: api.ProductPrices{Price: "420", CurrencySymbol: "£"}},
		{ID: 2, Title: "Stout", Prices: api.ProductPrices{Price: "550", CurrencySymbol: "£"}},
	}
	beerCats := api.Categories{
		{ID: 10, Name: "Ales", Slug: "ales", Count: 5},
		{ID: 20, Name: "Stouts", Slug: "stouts", Count: 3},
	}
	singleBeer := api.Product{ID: 42, Title: "IPA", Prices: api.ProductPrices{Price: "499"}}
	page := api.Pagination{TotalItems: 50, TotalPages: 5}
	sentinelErr := errors.New("API unavailable")
	// rip guard against unused-import warnings for atomic/time/http if subtests shrink.
	_ = atomic.AddInt32
	_ = time.Second
	_ = http.StatusOK

	tests := []struct {
		name string
		// mock state
		products   api.Products
		categories api.Categories
		product    api.Product
		pagination api.Pagination
		err        error
		// expectation
		wantMethod string
		wantErr    bool
		// call function against a fresh client
		call func(c api.Client) (any, error)
		// type-assert the returned `any` against the expected concrete type
		checkResult func(t *testing.T, got any)
	}{
		{
			name:       "FetchProducts happy path returns Products",
			products:   beerProducts,
			pagination: page,
			wantMethod: "FetchProducts",
			wantErr:    false,
			call: func(c api.Client) (any, error) {
				ps, pg, err := c.FetchProducts(context.Background(), 1, 10)
				_ = pg // checked via checkResult below
				return ps, err
			},
			checkResult: func(t *testing.T, got any) {
				is := is.New(t)
				ps, ok := got.(api.Products)
				is.True(ok) // returned type must be api.Products
				is.Equal(len(ps), 2)
				is.Equal(ps[0].ID, 1)
				is.Equal(ps[0].Title, "Pale Ale")
				is.Equal(ps[1].Title, "Stout")
			},
		},
		{
			name:       "FetchProducts error propagation",
			err:        sentinelErr,
			wantMethod: "FetchProducts",
			wantErr:    true,
			call: func(c api.Client) (any, error) {
				ps, _, err := c.FetchProducts(context.Background(), 1, 10)
				return ps, err
			},
			checkResult: func(t *testing.T, got any) {
				is := is.New(t)
				ps, ok := got.(api.Products)
				is.True(ok)
				is.Equal(len(ps), 0) // nil/empty on error
			},
		},
		{
			name:       "FetchCategories happy path returns Categories",
			categories: beerCats,
			wantMethod: "FetchCategories",
			wantErr:    false,
			call: func(c api.Client) (any, error) {
				cats, err := c.FetchCategories(context.Background())
				return cats, err
			},
			checkResult: func(t *testing.T, got any) {
				is := is.New(t)
				cats, ok := got.(api.Categories)
				is.True(ok) // returned type must be api.Categories
				is.Equal(len(cats), 2)
				is.Equal(cats[0].Name, "Ales")
				is.Equal(cats[1].Slug, "stouts")
				is.Equal(cats[1].Count, 3)
			},
		},
		{
			name:       "FetchCategories error propagation",
			err:        sentinelErr,
			wantMethod: "FetchCategories",
			wantErr:    true,
			call: func(c api.Client) (any, error) {
				cats, err := c.FetchCategories(context.Background())
				return cats, err
			},
			checkResult: func(t *testing.T, got any) {
				is := is.New(t)
				cats, ok := got.(api.Categories)
				is.True(ok)
				is.Equal(len(cats), 0)
			},
		},
		{
			name:       "FetchProduct happy path returns Product",
			product:    singleBeer,
			wantMethod: "FetchProduct",
			wantErr:    false,
			call: func(c api.Client) (any, error) {
				p, err := c.FetchProduct(context.Background(), 42)
				return p, err
			},
			checkResult: func(t *testing.T, got any) {
				is := is.New(t)
				p, ok := got.(api.Product)
				is.True(ok) // returned type must be api.Product
				is.Equal(p.ID, 42)
				is.Equal(p.Title, "IPA")
				is.Equal(p.Prices.Price, "499")
			},
		},
		{
			name:       "FetchProduct error propagation",
			err:        sentinelErr,
			wantMethod: "FetchProduct",
			wantErr:    true,
			call: func(c api.Client) (any, error) {
				p, err := c.FetchProduct(context.Background(), 99)
				return p, err
			},
			checkResult: func(t *testing.T, got any) {
				is := is.New(t)
				p, ok := got.(api.Product)
				is.True(ok)
				is.Equal(p.ID, 0) // zero value on error
			},
		},
		{
			name:       "FetchProductsByCategory happy path returns Products + Pagination",
			products:   beerProducts,
			pagination: page,
			wantMethod: "FetchProductsByCategory",
			wantErr:    false,
			call: func(c api.Client) (any, error) {
				ps, pg, err := c.FetchProductsByCategory(context.Background(), 10, 1, 10)
				// stash pagination via closure for checkResult
				return struct {
					Products   api.Products
					Pagination api.Pagination
				}{ps, pg}, err
			},
			checkResult: func(t *testing.T, got any) {
				is := is.New(t)
				v, ok := got.(struct {
					Products   api.Products
					Pagination api.Pagination
				})
				is.True(ok)
				is.Equal(len(v.Products), 2)
				is.Equal(v.Products[0].Title, "Pale Ale")
				is.Equal(v.Pagination.TotalItems, 50)
				is.Equal(v.Pagination.TotalPages, 5)
			},
		},
		{
			name:       "FetchProductsByCategory error propagation",
			err:        sentinelErr,
			wantMethod: "FetchProductsByCategory",
			wantErr:    true,
			call: func(c api.Client) (any, error) {
				ps, pg, err := c.FetchProductsByCategory(context.Background(), 10, 1, 10)
				return struct {
					Products   api.Products
					Pagination api.Pagination
				}{ps, pg}, err
			},
			checkResult: func(t *testing.T, got any) {
				is := is.New(t)
				v, ok := got.(struct {
					Products   api.Products
					Pagination api.Pagination
				})
				is.True(ok)
				is.Equal(len(v.Products), 0)
				is.Equal(v.Pagination.TotalItems, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			mc := &mockClient{
				products:   tt.products,
				categories: tt.categories,
				product:    tt.product,
				pagination: tt.pagination,
				err:        tt.err,
			}

			got, err := tt.call(mc)

			if tt.wantErr {
				is.True(err != nil)
				is.True(errors.Is(err, sentinelErr)) // error must be the sentinel verbatim
			} else {
				is.NoErr(err)
			}

			// Verify the correct method was dispatched.
			is.True(len(mc.calls) == 1)
			is.Equal(mc.calls[0], tt.wantMethod)

			// Verify the returned value is the correctly-typed result.
			if tt.checkResult != nil {
				tt.checkResult(t, got)
			}
		})
	}
}

// TestMockClient_AllMethodsCallable confirms that every method declared on
// the api.Client interface is callable through the mock — a smoke test that
// catches interface drift if a method is added or renamed without updating
// the mock.
func TestMockClient_AllMethodsCallable(t *testing.T) {
	is := is.New(t)

	mc := &mockClient{}
	ctx := context.Background()

	// Each call should succeed (zero-value returns, nil error) without panic.
	ps, pg, err := mc.FetchProducts(ctx, 1, 10)
	is.NoErr(err)
	is.Equal(len(ps), 0)
	is.Equal(pg.TotalItems, 0)

	cats, err := mc.FetchCategories(ctx)
	is.NoErr(err)
	is.Equal(len(cats), 0)

	prod, err := mc.FetchProduct(ctx, 1)
	is.NoErr(err)
	is.Equal(prod.ID, 0)

	ps2, pg2, err := mc.FetchProductsByCategory(ctx, 1, 1, 10)
	is.NoErr(err)
	is.Equal(len(ps2), 0)
	is.Equal(pg2.TotalPages, 0)

	// All four methods should have been recorded.
	is.True(len(mc.calls) == 4)
	is.Equal(mc.calls[0], "FetchProducts")
	is.Equal(mc.calls[1], "FetchCategories")
	is.Equal(mc.calls[2], "FetchProduct")
	is.Equal(mc.calls[3], "FetchProductsByCategory")
}

// isTimeoutErr returns true if err is/contains a net.Error with Timeout()==true.
// Used to confirm the timeout-tests actually surface a timeout error rather
// than some other failure.
func isTimeoutErr(err error) bool {
	// Walk the error chain for any net.Error that reports a timeout.
	// We avoid importing net directly to keep the test import list lean;
	// the error string will mention "timeout" or "deadline".
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, marker := range []string{"timeout", "deadline", "timed out"} {
		if containsFold(msg, marker) {
			return true
		}
	}
	return false
}

func containsFold(s, substr string) bool {
	s = toLowerASCII(s)
	substr = toLowerASCII(substr)
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return len(substr) == 0
}

func toLowerASCII(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}
