package commands

import (
	"context"
	"encoding/json"
	"fmt" // Added fmt import
	"net/http"
	"strconv" // Added strconv import
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TheHoptimistBaseURL is the base URL for The Hoptimist API.
// It's a variable so it can be overridden for testing.
var TheHoptimistBaseURL = "https://thehoptimist.co.uk"

type (
	Product struct {
		ID    int    `json:"id"`
		Link  string `json:"link"`
		Title struct {
			Rendered string `json:"rendered"`
		} `json:"title"`
		Description struct {
			Rendered string `json:"rendered"`
		} `json:"excerpt"`
	}
	Products          []Product
	LatestResponseMsg struct {
		Products   *Products
		Width      int
		Height     int
		TotalItems int // New
		TotalPages int // New
		Err        error
	}
	ProductsMsg struct {
		Product *Product
		Width   int
		Height  int
		Err     error
	}
)

// LoadLatestPageMsg is a message to indicate that a specific page of latest items should be loaded.
type LoadLatestPageMsg struct {
	Page    int
	PerPage int
}

// LoadCategoryProductsPageMsg is a message to indicate that a specific page of products for a category should be loaded.
type LoadCategoryProductsPageMsg struct {
	CategoryID   int
	CategoryName string
	APIEndpoint  string // The specific API endpoint for fetching this category's products
	Page         int
	PerPage      int
}

// Category models a product category from the WordPress API.
type Category struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Links struct {
		WpPostType []struct {
			Href string `json:"href"`
		} `json:"wp:post_type"`
	} `json:"_links"`
}

// Categories is a slice of Category.
type Categories []Category

// CategoriesResponseMsg is the message returned after fetching categories.
type CategoriesResponseMsg struct {
	Categories *Categories
	Width      int
	Height     int
	Err        error
}

// ProductsForCategoryResponseMsg is the message returned after fetching products for a category.
type ProductsForCategoryResponseMsg struct {
	Products     *Products
	CategoryName string
	CategoryID   int
	APIEndpoint  string // New: To carry over the API endpoint
	Width        int
	Height       int
	TotalItems   int // New
	TotalPages   int // New
	Err          error
}

// StartLoadingProductsForCategoryMsg is a message to indicate that products for a category should be loaded.
// This is defined here to avoid import cycles if defined in categoriesview, as hopt.go needs it.
type StartLoadingProductsForCategoryMsg struct {
	CategoryID   int
	CategoryName string
	APIEndpoint  string
	Width        int
	Height       int
	Page         int // New for pagination
	PerPage      int // New for pagination
}

func HandleGetLatest(w int, h int, page int, perPage int) tea.Cmd {
	return func() tea.Msg {
		if page <= 0 {
			page = 1
		}
		if perPage <= 0 {
			perPage = 10
		}
		url := fmt.Sprintf("%s/wp-json/wp/v2/product?orderby=date&page=%d&per_page=%d", TheHoptimistBaseURL, page, perPage)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return LatestResponseMsg{
				Err: err,
			}
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return LatestResponseMsg{
				Err: err,
			}
		}
		defer res.Body.Close()

		totalItemsStr := res.Header.Get("X-WP-Total")
		totalPagesStr := res.Header.Get("X-WP-TotalPages")
		totalItems, _ := strconv.Atoi(totalItemsStr)
		totalPages, _ := strconv.Atoi(totalPagesStr)

		var products Products
		decodeErr := json.NewDecoder(res.Body).Decode(&products) // Renamed err to decodeErr
		if decodeErr != nil {
			return LatestResponseMsg{
				Err: decodeErr, // Use decodeErr here
			}
		}

		return LatestResponseMsg{
			Products:   &products,
			Width:      w,
			Height:     h,
			TotalItems: totalItems,
			TotalPages: totalPages,
			// Err should be nil if decodeErr was nil
		}
	}
}

// HandleGetCategories fetches product categories from the API.
func HandleGetCategories(w int, h int) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("%s/wp-json/wp/v2/product_cat", TheHoptimistBaseURL)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return CategoriesResponseMsg{Err: err}
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return CategoriesResponseMsg{Err: err}
		}
		defer res.Body.Close()

		var categories Categories
		if err := json.NewDecoder(res.Body).Decode(&categories); err != nil {
			return CategoriesResponseMsg{Err: err}
		}

		return CategoriesResponseMsg{
			Categories: &categories,
			Width:      w,
			Height:     h,
		}
	}
}

// HandleGetProductsByCategory fetches products for a given category from the API.
func HandleGetProductsByCategory(w int, h int, categoryID int, categoryName string, apiEndpoint string, page int, perPage int) tea.Cmd {
	return func() tea.Msg {
		if page <= 0 {
			page = 1
		}
		if perPage <= 0 {
			perPage = 10
		}

		if apiEndpoint == "" {
			return ProductsForCategoryResponseMsg{
				Err:          NewError("API endpoint for category is empty"),
				CategoryName: categoryName,
				CategoryID:   categoryID,
			}
		}

		url := fmt.Sprintf("%s&page=%d&per_page=%d", apiEndpoint, page, perPage)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return ProductsForCategoryResponseMsg{Err: err, CategoryName: categoryName, CategoryID: categoryID}
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return ProductsForCategoryResponseMsg{Err: err, CategoryName: categoryName, CategoryID: categoryID}
		}
		defer res.Body.Close()

		totalItemsStr := res.Header.Get("X-WP-Total")
		totalPagesStr := res.Header.Get("X-WP-TotalPages")
		totalItems, _ := strconv.Atoi(totalItemsStr)
		totalPages, _ := strconv.Atoi(totalPagesStr)

		var products Products
		decodeErr := json.NewDecoder(res.Body).Decode(&products)
		if decodeErr != nil {
			return ProductsForCategoryResponseMsg{Err: decodeErr, CategoryName: categoryName, CategoryID: categoryID}
		}

		return ProductsForCategoryResponseMsg{
			Products:     &products,
			CategoryName: categoryName,
			CategoryID:   categoryID,
			APIEndpoint:  apiEndpoint, // Populate the APIEndpoint
			Width:        w,
			Height:       h,
			TotalItems:   totalItems,
			TotalPages:   totalPages,
		}
	}
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

func HandleDisplayProduct(w int, h int, prod Product) tea.Cmd {
	return func() tea.Msg {
		return ProductsMsg{
			Product: &prod,
			Width:   w,
			Height:  h,
		}
	}
}
