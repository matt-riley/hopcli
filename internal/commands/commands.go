package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// TheHoptimistBaseURL is the base URL for The Hoptimist API.
// It's a variable so it can be overridden for testing.
var TheHoptimistBaseURL = "https://thehoptimist.co.uk"

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

type (
	Product struct {
		ID               int           `json:"id"`
		Link             string        `json:"permalink"`
		Title            string        `json:"name"`
		Description      string        `json:"description"`
		ShortDescription string        `json:"short_description"`
		Prices           ProductPrices `json:"prices"`
		OnSale           bool          `json:"on_sale"`
		IsInStock        bool          `json:"is_in_stock"`
	}
	Products          []Product
	LatestResponseMsg struct {
		Products   *Products
		Width      int
		Height     int
		TotalItems int
		TotalPages int
		RequestID  int
		Err        error
	}
	ProductsMsg struct {
		Product *Product
		Width   int
		Height  int
		NavGen  int
		Err     error
	}
)

// LoadLatestPageMsg is a message to indicate that a specific page of latest items should be loaded.
type LoadLatestPageMsg struct {
	Page    int
	PerPage int
	NavGen  int
}

// LoadCategoryProductsPageMsg is a message to indicate that a specific page of products for a category should be loaded.
type LoadCategoryProductsPageMsg struct {
	CategoryID   int
	CategoryName string
	Page         int
	PerPage      int
	NavGen       int
}

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

// CategoriesResponseMsg is the message returned after fetching categories.
type CategoriesResponseMsg struct {
	Categories *Categories
	Width      int
	Height     int
	RequestID  int
	Err        error
}

// ProductsForCategoryResponseMsg is the message returned after fetching products for a category.
type ProductsForCategoryResponseMsg struct {
	Products     *Products
	CategoryName string
	CategoryID   int
	Width        int
	Height       int
	TotalItems   int
	TotalPages   int
	RequestID    int
	Err          error
}

// StartLoadingProductsForCategoryMsg is a message to indicate that products for a category should be loaded.
// This is defined here to avoid import cycles if defined in categoriesview, as hopt.go needs it.
type StartLoadingProductsForCategoryMsg struct {
	CategoryID   int
	CategoryName string
	Width        int
	Height       int
	Page         int
	PerPage      int
	NavGen       int
}

func HandleGetLatest(w int, h int, page int, perPage int, requestID int) tea.Cmd {
	return func() tea.Msg {
		if page <= 0 {
			page = 1
		}
		if perPage <= 0 {
			perPage = 10
		}
		url := fmt.Sprintf("%s/wp-json/wc/store/v1/products?orderby=date&order=desc&page=%d&per_page=%d", TheHoptimistBaseURL, page, perPage)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return LatestResponseMsg{
				Err:       err,
				RequestID: requestID,
			}
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return LatestResponseMsg{
				Err:       err,
				RequestID: requestID,
			}
		}
		defer res.Body.Close()

		if res.StatusCode < 200 || res.StatusCode > 299 {
			return LatestResponseMsg{
				Err:       fmt.Errorf("unexpected status %d from API", res.StatusCode),
				RequestID: requestID,
			}
		}

		totalItemsStr := res.Header.Get("X-WP-Total")
		totalPagesStr := res.Header.Get("X-WP-TotalPages")

		var totalItems int
		if totalItemsStr != "" {
			parsedTotalItems, err := strconv.Atoi(totalItemsStr)
			if err != nil {
				// fmt.Fprintf(os.Stderr, "Warning: could not parse X-WP-Total header '%s': %v\n", totalItemsStr, err)
				totalItems = 0 // Default to 0 on error
			} else {
				totalItems = parsedTotalItems
			}
		} else {
			totalItems = 0 // Default if header is empty
		}

		var totalPages int
		if totalPagesStr != "" {
			parsedTotalPages, err := strconv.Atoi(totalPagesStr)
			if err != nil {
				// fmt.Fprintf(os.Stderr, "Warning: could not parse X-WP-TotalPages header '%s': %v\n", totalPagesStr, err)
				totalPages = 0 // Default to 0 on error
			} else {
				totalPages = parsedTotalPages
			}
		} else {
			totalPages = 0 // Default if header is empty
		}

		var products Products
		decodeErr := json.NewDecoder(res.Body).Decode(&products) // Renamed err to decodeErr
		if decodeErr != nil {
			return LatestResponseMsg{
				Err:       decodeErr,
				RequestID: requestID,
			}
		}

		return LatestResponseMsg{
			Products:   &products,
			Width:      w,
			Height:     h,
			TotalItems: totalItems,
			TotalPages: totalPages,
			RequestID:  requestID,
		}
	}
}

// HandleGetCategories fetches product categories from the API.
func HandleGetCategories(requestID int) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("%s/wp-json/wc/store/v1/products/categories", TheHoptimistBaseURL)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return CategoriesResponseMsg{Err: err, RequestID: requestID}
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return CategoriesResponseMsg{Err: err, RequestID: requestID}
		}
		defer res.Body.Close()

		if res.StatusCode < 200 || res.StatusCode > 299 {
			return CategoriesResponseMsg{
				Err:       fmt.Errorf("unexpected status %d from API", res.StatusCode),
				RequestID: requestID,
			}
		}

		var categories Categories
		if err := json.NewDecoder(res.Body).Decode(&categories); err != nil {
			return CategoriesResponseMsg{Err: err, RequestID: requestID}
		}

		return CategoriesResponseMsg{
			Categories: &categories,
			RequestID:  requestID,
		}
	}
}

// HandleGetProductsByCategory fetches products for a given category from the API.
func HandleGetProductsByCategory(categoryID int, categoryName string, page int, perPage int, requestID int) tea.Cmd {
	return func() tea.Msg {
		if page <= 0 {
			page = 1
		}
		if perPage <= 0 {
			perPage = 10
		}

		url := fmt.Sprintf("%s/wp-json/wc/store/v1/products?category=%d&page=%d&per_page=%d",
			TheHoptimistBaseURL, categoryID, page, perPage)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return ProductsForCategoryResponseMsg{Err: err, CategoryName: categoryName, CategoryID: categoryID, RequestID: requestID}
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return ProductsForCategoryResponseMsg{Err: err, CategoryName: categoryName, CategoryID: categoryID, RequestID: requestID}
		}
		defer res.Body.Close()

		if res.StatusCode < 200 || res.StatusCode > 299 {
			return ProductsForCategoryResponseMsg{
				Err:          fmt.Errorf("unexpected status %d from API", res.StatusCode),
				CategoryName: categoryName,
				CategoryID:   categoryID,
				RequestID:    requestID,
			}
		}

		totalItemsStr := res.Header.Get("X-WP-Total")
		totalPagesStr := res.Header.Get("X-WP-TotalPages")

		var totalItems int
		if totalItemsStr != "" {
			parsedTotalItems, err := strconv.Atoi(totalItemsStr)
			if err != nil {
				// fmt.Fprintf(os.Stderr, "Warning: could not parse X-WP-Total header '%s': %v\n", totalItemsStr, err)
				totalItems = 0 // Default to 0 on error
			} else {
				totalItems = parsedTotalItems
			}
		} else {
			totalItems = 0 // Default if header is empty
		}

		var totalPages int
		if totalPagesStr != "" {
			parsedTotalPages, err := strconv.Atoi(totalPagesStr)
			if err != nil {
				// fmt.Fprintf(os.Stderr, "Warning: could not parse X-WP-TotalPages header '%s': %v\n", totalPagesStr, err)
				totalPages = 0 // Default to 0 on error
			} else {
				totalPages = parsedTotalPages
			}
		} else {
			totalPages = 0 // Default if header is empty
		}

		var products Products
		decodeErr := json.NewDecoder(res.Body).Decode(&products)
		if decodeErr != nil {
			return ProductsForCategoryResponseMsg{Err: decodeErr, CategoryName: categoryName, CategoryID: categoryID, RequestID: requestID}
		}

		return ProductsForCategoryResponseMsg{
			Products:     &products,
			CategoryName: categoryName,
			CategoryID:   categoryID,
			TotalItems:   totalItems,
			TotalPages:   totalPages,
			RequestID:    requestID,
		}
	}
}

// FormatPrice converts a Store API minor-unit price string to a display string.
// e.g. FormatPrice("420", "£", "", 2) → "£4.20"
//
// Edge cases:
//   - price == "" → returns "" (callers must suppress empty price in UI)
//   - non-numeric price → returns prefix + price + suffix as-is
//   - minorUnit < 0 → same fallback as non-numeric
//   - minorUnit == 0 → no decimal separator
//   - price == "0" → returns formatted zero (e.g. "£0.00")
func FormatPrice(price string, prefix string, suffix string, minorUnit int) string {
	if price == "" {
		return ""
	}
	n, err := strconv.Atoi(price)
	if err != nil || minorUnit < 0 {
		return prefix + price + suffix
	}
	divisor := math.Pow10(minorUnit)
	return fmt.Sprintf("%s%."+strconv.Itoa(minorUnit)+"f%s", prefix, float64(n)/divisor, suffix)
}

// NewError creates a new error.
// Helper function to create error instances easily.
func NewError(text string) error {
	return &errorString{text}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

var (
	summaryRe  = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
	innerTagRe = regexp.MustCompile(`<[^>]+>`)
)

// ExtractSummary extracts the plain-text content of the first <p> tag from an
// HTML description string. This returns the structured beer summary
// (e.g. "IPA – New England / Hazy – Can 440ml – 6.5%") from product descriptions
// where short_description is unpopulated.
// Returns "" if no <p> tag is found.
func ExtractSummary(description string) string {
	m := summaryRe.FindStringSubmatch(description)
	if len(m) < 2 {
		return ""
	}
	inner := innerTagRe.ReplaceAllString(m[1], "")
	return strings.TrimSpace(html.UnescapeString(inner))
}

func HandleDisplayProduct(w int, h int, prod Product) tea.Cmd {
	return func() tea.Msg {
		return ProductsMsg{
			Product: &prod,
			Width:   w,
			Height:  h,
		}
	}
}
