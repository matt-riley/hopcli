package commands

import (
	"context"
	"fmt"
	"html"
	"math"
	"regexp"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/matt-riley/hopcli/internal/api"
)

// Version is the current hopcli version. Override at build time with:
//
//	go build -ldflags "-X github.com/matt-riley/hopcli/internal/commands.Version=$(git describe --tags --always --dirty)"
var Version = "dev"

// TheHoptimistBaseURL is the base URL for The Hoptimist API.
var TheHoptimistBaseURL = "https://thehoptimist.co.uk"

// ApiClient is the HTTP client used for all API requests.
// It implements the api.Client interface. Override for testing.
var ApiClient api.Client

// SpinnerColor is the lipgloss color string for the loading spinner.
const SpinnerColor = "205"

// ---- Type aliases (backed by api package, preserved for backward compatibility) ----

type Product = api.Product
type Products = api.Products
type ProductPrices = api.ProductPrices
type Category = api.Category
type Categories = api.Categories
type Pagination = api.Pagination

// ---- Bubble Tea message types ----

type (
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

// CategoriesResponseMsg is the message returned after fetching categories.
type CategoriesResponseMsg struct {
	Categories *Categories
	RequestID  int
	Err        error
}

// ProductsForCategoryResponseMsg is the message returned after fetching products for a category.
type ProductsForCategoryResponseMsg struct {
	Products     *Products
	CategoryName string
	CategoryID   int
	TotalItems   int
	TotalPages   int
	RequestID    int
	Err          error
}

// StartLoadingProductsForCategoryMsg is a message to indicate that products for a category should be loaded.
type StartLoadingProductsForCategoryMsg struct {
	CategoryID   int
	CategoryName string
	Width        int
	Height       int
	Page         int
	PerPage      int
	NavGen       int
}

// ---- Command handlers ----

func HandleGetLatest(w int, h int, page int, perPage int, requestID int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		products, pagination, err := ApiClient.FetchProducts(ctx, page, perPage)
		if err != nil {
			return LatestResponseMsg{
				Err:       err,
				RequestID: requestID,
			}
		}
		return LatestResponseMsg{
			Products:   &products,
			Width:      w,
			Height:     h,
			TotalItems: pagination.TotalItems,
			TotalPages: pagination.TotalPages,
			RequestID:  requestID,
		}
	}
}

// HandleGetCategories fetches product categories from the API.
func HandleGetCategories(requestID int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		categories, err := ApiClient.FetchCategories(ctx)
		if err != nil {
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
		ctx := context.Background()
		products, pagination, err := ApiClient.FetchProductsByCategory(ctx, categoryID, page, perPage)
		if err != nil {
			return ProductsForCategoryResponseMsg{
				Err:          err,
				CategoryName: categoryName,
				CategoryID:   categoryID,
				RequestID:    requestID,
			}
		}
		return ProductsForCategoryResponseMsg{
			Products:     &products,
			CategoryName: categoryName,
			CategoryID:   categoryID,
			TotalItems:   pagination.TotalItems,
			TotalPages:   pagination.TotalPages,
			RequestID:    requestID,
		}
	}
}

// ---- Formatting helpers ----

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

var (
	summaryRe  = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
	innerTagRe = regexp.MustCompile(`<[^>]+>`)
)

// ExtractSummary extracts the plain-text content of the first <p> tag from an
// HTML description string.
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

// ---- Shared composable types ----

// ProductListItem is a shared list.Item implementation for displaying products.
// Used by both the latest and categoryproducts views.
type ProductListItem struct {
	title       string
	desc        string
	productData Product
}

func (i ProductListItem) Title() string       { return i.title }
func (i ProductListItem) Description() string { return i.desc }
func (i ProductListItem) FilterValue() string { return i.title }

// ProductData returns the underlying Product for this list item.
func (i ProductListItem) ProductData() Product { return i.productData }

// NewProductListItem creates a list.Item from a Product with formatted price and description.
func NewProductListItem(p Product) ProductListItem {
	processedDesc := ExtractSummary(p.Description)

	formattedPrice := FormatPrice(
		p.Prices.Price,
		p.Prices.CurrencyPrefix,
		p.Prices.CurrencySuffix,
		p.Prices.CurrencyMinorUnit,
	)
	onSaleMarker := ""
	if p.OnSale {
		onSaleMarker = " 🏷️"
	}

	var desc string
	if formattedPrice != "" {
		desc = fmt.Sprintf("%s%s | %s", formattedPrice, onSaleMarker, processedDesc)
	} else {
		desc = processedDesc
	}

	return ProductListItem{
		title:       html.UnescapeString(p.Title),
		desc:        desc,
		productData: p,
	}
}

// PaginatedModel provides shared pagination state and navigation keys ('n'/'p').
// Embed this struct in models that need paginated list views.
type PaginatedModel struct {
	CurrentPage int
	PerPage     int
	TotalItems  int
	TotalPages  int
}

// UpdatePageNavigation handles 'n' (next) and 'p' (previous) key presses.
// Returns true and the new page number when navigation occurred.
func (pm *PaginatedModel) UpdatePageNavigation(msg tea.KeyMsg) (pageChanged bool, newPage int) {
	switch msg.String() {
	case "n":
		if pm.CurrentPage < pm.TotalPages {
			pm.CurrentPage++
			return true, pm.CurrentPage
		}
	case "p":
		if pm.CurrentPage > 1 {
			pm.CurrentPage--
			return true, pm.CurrentPage
		}
	}
	return false, 0
}

// ResponseError extracts the error from common response message types.
// Returns nil if the message does not carry an error, or if the error field is nil.
func ResponseError(msg tea.Msg) error {
	switch m := msg.(type) {
	case LatestResponseMsg:
		return m.Err
	case CategoriesResponseMsg:
		return m.Err
	case ProductsForCategoryResponseMsg:
		return m.Err
	case ProductsMsg:
		return m.Err
	}
	return nil
}
