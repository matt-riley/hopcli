package commands

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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
		Products *Products
		Width    int
		Height   int
		Err      error
	}
	ProductsMsg struct {
		Product *Product
		Width   int
		Height  int
		Err     error
	}
)

// Category models a product category from the WordPress API.
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
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
	Width        int
	Height       int
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
}

func HandleGetLatest(w int, h int) tea.Cmd {
	return func() tea.Msg {
		url := "https://thehoptimist.co.uk/wp-json/wp/v2/product?order_by=date"
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

		var products Products
		err = json.NewDecoder(res.Body).Decode(&products)
		if err != nil {
			return LatestResponseMsg{
				Err: err,
			}
		}

		return LatestResponseMsg{
			Products: &products,
			Width:    w,
			Height:   h,
		}
	}
}

// HandleGetCategories fetches product categories from the API.
func HandleGetCategories(w int, h int) tea.Cmd {
	return func() tea.Msg {
		url := "https://thehoptimist.co.uk/wp-json/wp/v2/product_cat"
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
func HandleGetProductsByCategory(w int, h int, categoryID int, categoryName string, apiEndpoint string) tea.Cmd {
	return func() tea.Msg {
		if apiEndpoint == "" {
			return ProductsForCategoryResponseMsg{
				Err:          NewError("API endpoint for category is empty"),
				CategoryName: categoryName,
				CategoryID:   categoryID,
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiEndpoint, nil)
		if err != nil {
			return ProductsForCategoryResponseMsg{Err: err, CategoryName: categoryName, CategoryID: categoryID}
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return ProductsForCategoryResponseMsg{Err: err, CategoryName: categoryName, CategoryID: categoryID}
		}
		defer res.Body.Close()

		var products Products
		if err := json.NewDecoder(res.Body).Decode(&products); err != nil {
			return ProductsForCategoryResponseMsg{Err: err, CategoryName: categoryName, CategoryID: categoryID}
		}

		return ProductsForCategoryResponseMsg{
			Products:     &products,
			CategoryName: categoryName,
			CategoryID:   categoryID,
			Width:        w,
			Height:       h,
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
