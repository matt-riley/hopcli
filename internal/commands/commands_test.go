package commands_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/internal/commands"
)

func TestHandleGetLatest_Pagination(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.Equal(r.URL.Path, "/wp-json/wc/store/v1/products")
		is.Equal(r.URL.Query().Get("page"), "2")
		is.Equal(r.URL.Query().Get("per_page"), "5")
		is.Equal(r.URL.Query().Get("orderby"), "date")
		is.Equal(r.URL.Query().Get("order"), "desc")

		w.Header().Set("X-WP-Total", "50")
		w.Header().Set("X-WP-TotalPages", "10")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`[{"id":1,"name":"Test Beer 1","description":"Desc 1","short_description":"Short 1","permalink":"https://example.com/test-beer-1","prices":{"price":"420","regular_price":"420","sale_price":"","currency_code":"GBP","currency_symbol":"£","currency_minor_unit":2,"currency_prefix":"£","currency_suffix":""}}]`))
		is.NoErr(err)
	}))
	defer server.Close()

	originalBaseURL := commands.TheHoptimistBaseURL
	commands.TheHoptimistBaseURL = server.URL
	defer func() { commands.TheHoptimistBaseURL = originalBaseURL }()

	cmd := commands.HandleGetLatest(80, 24, 2, 5, 1)
	msg := cmd()

	is.True(msg != nil)
	latestMsg, ok := msg.(commands.LatestResponseMsg)
	is.True(ok)

	is.True(latestMsg.Err == nil)
	is.Equal(latestMsg.RequestID, 1)
	is.Equal(latestMsg.TotalItems, 50)
	is.Equal(latestMsg.TotalPages, 10)
	is.True(latestMsg.Products != nil)
	is.Equal(len(*latestMsg.Products), 1)
	if len(*latestMsg.Products) == 1 {
		is.Equal((*latestMsg.Products)[0].ID, 1)
		is.Equal((*latestMsg.Products)[0].Title, "Test Beer 1")
	}
}

func TestHandleGetProductsByCategory_Pagination(t *testing.T) {
	is := is.New(t)

	expectedCatID := 123
	expectedPage := "3"
	expectedPerPage := "8"
	expectedTotalItems := 24
	expectedTotalPages := 3

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.Equal(r.URL.Path, "/wp-json/wc/store/v1/products")
		is.Equal(r.URL.Query().Get("category"), strconv.Itoa(expectedCatID))
		is.Equal(r.URL.Query().Get("page"), expectedPage)
		is.Equal(r.URL.Query().Get("per_page"), expectedPerPage)

		w.Header().Set("X-WP-Total", strconv.Itoa(expectedTotalItems))
		w.Header().Set("X-WP-TotalPages", strconv.Itoa(expectedTotalPages))
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`[{"id":2,"name":"Category Beer 2","description":"Desc 2","short_description":"Short 2","permalink":"https://example.com/category-beer-2","prices":{"price":"599","regular_price":"599","sale_price":"","currency_code":"GBP","currency_symbol":"£","currency_minor_unit":2,"currency_prefix":"£","currency_suffix":""}}]`))
		is.NoErr(err)
	}))
	defer server.Close()

	originalBaseURL := commands.TheHoptimistBaseURL
	commands.TheHoptimistBaseURL = server.URL
	defer func() { commands.TheHoptimistBaseURL = originalBaseURL }()

	cmd := commands.HandleGetProductsByCategory(expectedCatID, "Test Cat", 3, 8, 1)
	msg := cmd()

	is.True(msg != nil)
	catMsg, ok := msg.(commands.ProductsForCategoryResponseMsg)
	is.True(ok)

	is.True(catMsg.Err == nil)
	is.Equal(catMsg.RequestID, 1)
	is.Equal(catMsg.TotalItems, expectedTotalItems)
	is.Equal(catMsg.TotalPages, expectedTotalPages)
	is.Equal(catMsg.CategoryID, expectedCatID)
	is.Equal(catMsg.CategoryName, "Test Cat")
	is.True(catMsg.Products != nil)
	is.Equal(len(*catMsg.Products), 1)
	if len(*catMsg.Products) == 1 {
		is.Equal((*catMsg.Products)[0].ID, 2)
		is.Equal((*catMsg.Products)[0].Title, "Category Beer 2")
	}
}

func TestHandleGetCategories_UsesBaseURL(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		is.Equal(r.URL.Path, "/wp-json/wc/store/v1/products/categories")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`[{"id":1,"name":"Category 1","slug":"category-1","parent":0,"count":5}]`))
		is.NoErr(err)
	}))
	defer server.Close()

	originalBaseURL := commands.TheHoptimistBaseURL
	commands.TheHoptimistBaseURL = server.URL
	defer func() { commands.TheHoptimistBaseURL = originalBaseURL }()

	cmd := commands.HandleGetCategories(1)
	msg := cmd()

	is.True(msg != nil)
	catMsg, ok := msg.(commands.CategoriesResponseMsg)
	is.True(ok)
	is.True(catMsg.Err == nil)
	is.Equal(catMsg.RequestID, 1)
	is.True(catMsg.Categories != nil && len(*catMsg.Categories) == 1)
	if len(*catMsg.Categories) == 1 {
		is.Equal((*catMsg.Categories)[0].Name, "Category 1")
	}
}

func TestHandleGetLatest_HTTPErrorStatus(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	originalBaseURL := commands.TheHoptimistBaseURL
	commands.TheHoptimistBaseURL = server.URL
	defer func() { commands.TheHoptimistBaseURL = originalBaseURL }()

	cmd := commands.HandleGetLatest(80, 24, 1, 10, 1)
	msg := cmd()

	latestMsg, ok := msg.(commands.LatestResponseMsg)
	is.True(ok)
	is.True(latestMsg.Err != nil)
}

func TestHandleGetLatest_ErrorResponseCarriesRequestID(t *testing.T) {
	is := is.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	originalBaseURL := commands.TheHoptimistBaseURL
	commands.TheHoptimistBaseURL = server.URL
	defer func() { commands.TheHoptimistBaseURL = originalBaseURL }()

	cmd := commands.HandleGetLatest(80, 24, 1, 10, 42)
	msg := cmd()

	latestMsg, ok := msg.(commands.LatestResponseMsg)
	is.True(ok)
	is.True(latestMsg.Err != nil)
	is.Equal(latestMsg.RequestID, 42)
}

func TestFormatPrice(t *testing.T) {
	is := is.New(t)

	// Normal GBP price
	is.Equal(commands.FormatPrice("420", "£", "", 2), "£4.20")

	// Zero price
	is.Equal(commands.FormatPrice("0", "£", "", 2), "£0.00")

	// Zero minor unit (e.g. Japanese yen)
	is.Equal(commands.FormatPrice("100", "¥", "", 0), "¥100")

	// Empty price
	is.Equal(commands.FormatPrice("", "£", "", 2), "")

	// Non-numeric price fallback
	is.Equal(commands.FormatPrice("free", "£", "", 2), "£free")

	// Negative minorUnit fallback
	is.Equal(commands.FormatPrice("420", "£", "", -1), "£420")
}

func TestExtractSummary(t *testing.T) {
	is := is.New(t)

	// Normal 4-part summary with HTML entities for en-dash
	is.Equal(
		commands.ExtractSummary(`<p>Lager &#8211; Helles &#8211; Bottle 500ml &#8211; 5.0%</p><p>Full description here.</p>`),
		"Lager – Helles – Bottle 500ml – 5.0%",
	)

	// Can format
	is.Equal(
		commands.ExtractSummary(`<p>IPA &#8211; New England / Hazy &#8211; Can 440ml &#8211; 6.5%</p>`),
		"IPA – New England / Hazy – Can 440ml – 6.5%",
	)

	// 3-part format (no substyle)
	is.Equal(
		commands.ExtractSummary(`<p>Pale Ale &#8211; Can 440ml &#8211; 5.4%</p>`),
		"Pale Ale – Can 440ml – 5.4%",
	)

	// First <p> has class/id attributes (matches p[^>]*)
	is.Equal(
		commands.ExtractSummary(`<p class="summary">Stout &#8211; Imperial &#8211; Can 330ml &#8211; 10.0%</p>`),
		"Stout – Imperial – Can 330ml – 10.0%",
	)

	// First <p> has inner HTML tags (stripped)
	is.Equal(
		commands.ExtractSummary(`<p><strong>Wheat Beer</strong> &#8211; Hefeweizen &#8211; Bottle 500ml &#8211; 5.3%</p>`),
		"Wheat Beer – Hefeweizen – Bottle 500ml – 5.3%",
	)

	// No <p> tag — returns ""
	is.Equal(commands.ExtractSummary(`Just plain text, no tags`), "")

	// Empty description — returns ""
	is.Equal(commands.ExtractSummary(""), "")

	// Whitespace trimmed
	is.Equal(
		commands.ExtractSummary(`<p>  Lager &#8211; Helles &#8211; Bottle 500ml &#8211; 5.0%  </p>`),
		"Lager – Helles – Bottle 500ml – 5.0%",
	)
}
