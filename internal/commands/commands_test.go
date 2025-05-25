package commands_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	// "strings" // Not strictly needed for these specific tests based on final structure
	"testing"

	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/internal/commands"
)

func TestHandleGetLatest_Pagination(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.Equal(r.URL.Query().Get("page"), "2")
		is.Equal(r.URL.Query().Get("per_page"), "5")
		is.Equal(r.URL.Query().Get("orderby"), "date")

		w.Header().Set("X-WP-Total", "50")
		w.Header().Set("X-WP-TotalPages", "10") // 50 items, 5 per page = 10 pages
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`[{"id":1,"title":{"rendered":"Test Beer 1"}, "excerpt":{"rendered":"Desc 1"}}]`))
		is.NoErr(err)
	}))
	defer server.Close()

	originalBaseURL := commands.TheHoptimistBaseURL
	commands.TheHoptimistBaseURL = server.URL // Override package-level variable
	defer func() { commands.TheHoptimistBaseURL = originalBaseURL }()

	cmd := commands.HandleGetLatest(80, 24, 2, 5) // Test page 2, 5 per page
	msg := cmd()                                  // Execute the command

	is.True(msg != nil)
	latestMsg, ok := msg.(commands.LatestResponseMsg)
	is.True(ok) // Message is of expected type

	is.True(latestMsg.Err == nil) // Should be no error
	is.Equal(latestMsg.TotalItems, 50)
	is.Equal(latestMsg.TotalPages, 10)
	is.True(latestMsg.Products != nil) // Products should not be nil
	is.Equal(len(*latestMsg.Products), 1)
	if len(*latestMsg.Products) == 1 {
		is.Equal((*latestMsg.Products)[0].ID, 1)
		is.Equal((*latestMsg.Products)[0].Title.Rendered, "Test Beer 1")
	}
}

func TestHandleGetProductsByCategory_Pagination(t *testing.T) {
	is := is.New(t)

	expectedCatIDStr := "123"
	expectedPage := "3"
	expectedPerPage := "8"
	expectedTotalItems := 24
	expectedTotalPages := 3 // 24 items, 8 per page = 3 pages

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Example: endpoint might be /wp-json/wp/v2/product?product_cat=123
		// The test will construct testAPIEndpoint as server.URL + "/wp-json/wp/v2/product?product_cat=123"
		// So, r.URL.Path will be "/wp-json/wp/v2/product"
		// and r.URL.Query() will contain "product_cat", "page", "per_page"

		is.Equal(r.URL.Query().Get("product_cat"), expectedCatIDStr) // Assuming this is part of the base apiEndpoint
		is.Equal(r.URL.Query().Get("page"), expectedPage)
		is.Equal(r.URL.Query().Get("per_page"), expectedPerPage)

		w.Header().Set("X-WP-Total", strconv.Itoa(expectedTotalItems))
		w.Header().Set("X-WP-TotalPages", strconv.Itoa(expectedTotalPages))
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`[{"id":2,"title":{"rendered":"Category Beer 2"}, "excerpt":{"rendered":"Desc 2"}}]`))
		is.NoErr(err)
	}))
	defer server.Close()

	// Construct the testAPIEndpoint using the server's URL
	// The original apiEndpoint that HandleGetProductsByCategory receives is a full URL.
	// For the test, we make this full URL point to our test server.
	// The HandleGetProductsByCategory function will then append &page=...&per_page=... to it.
	testAPIEndpoint := fmt.Sprintf("%s/wp-json/wp/v2/product?product_cat=%s", server.URL, expectedCatIDStr)

	cmd := commands.HandleGetProductsByCategory(80, 24, 123, "Test Cat", testAPIEndpoint, 3, 8)
	msg := cmd()

	is.True(msg != nil)
	catMsg, ok := msg.(commands.ProductsForCategoryResponseMsg)
	is.True(ok) // Message is of expected type

	is.True(catMsg.Err == nil)
	is.Equal(catMsg.TotalItems, expectedTotalItems)
	is.Equal(catMsg.TotalPages, expectedTotalPages)
	is.Equal(catMsg.CategoryID, 123)
	is.Equal(catMsg.CategoryName, "Test Cat")
	is.Equal(catMsg.APIEndpoint, testAPIEndpoint) // Check if APIEndpoint is passed through
	is.True(catMsg.Products != nil)
	is.Equal(len(*catMsg.Products), 1)
	if len(*catMsg.Products) == 1 {
		is.Equal((*catMsg.Products)[0].ID, 2)
		is.Equal((*catMsg.Products)[0].Title.Rendered, "Category Beer 2")
	}
}

func TestHandleGetCategories_UsesBaseURL(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.Equal(r.URL.Path, "/wp-json/wp/v2/product_cat") // Check path
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`[{"id":1,"name":"Category 1"}]`)) // Minimal valid JSON
		is.NoErr(err)
	}))
	defer server.Close()

	originalBaseURL := commands.TheHoptimistBaseURL
	commands.TheHoptimistBaseURL = server.URL // Override
	defer func() { commands.TheHoptimistBaseURL = originalBaseURL }()

	cmd := commands.HandleGetCategories(80, 24)
	msg := cmd()

	is.True(msg != nil)
	catMsg, ok := msg.(commands.CategoriesResponseMsg)
	is.True(ok)
	is.True(catMsg.Err == nil)
	is.True(catMsg.Categories != nil && len(*catMsg.Categories) == 1)
	if len(*catMsg.Categories) == 1 {
		is.Equal((*catMsg.Categories)[0].Name, "Category 1")
	}
}
