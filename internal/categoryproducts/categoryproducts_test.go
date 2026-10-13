package categoryproducts_test

import (
	"fmt"
	"strings"
	"testing"

	// list "charm.land/bubbles/v2/list" // Not directly used for assertions
	tea "charm.land/bubbletea/v2"
	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/internal/categoryproducts"
	"github.com/matt-riley/hopcli/internal/commands"
)

func TestNewCategoryProductsModel_InitialValues(t *testing.T) {
	is := is.New(t)
	model := categoryproducts.NewModel("Test Category", 123)

	// Assert exported fields
	is.Equal(model.CurrentPage, 1)
	is.Equal(model.PerPage, 10)
	// is.Equal(model.CategoryName, "Test Category") // CategoryName is unexported
	// is.Equal(model.CategoryID, 123)             // CategoryID is unexported
	// is.Equal(model.ApiEndpoint, "http://fake.api/ep") // apiEndpoint is unexported

	// Check list properties (these are set via methods on list.Model, not direct field access on our Model)
	// We can't directly access Paginator.ShowPagination or ShowStatusBar easily without reflection
	// or if the list.Model itself provided getters.
	// However, we trust that NewModel calls l.SetShowPagination(true) and l.SetShowStatusBar(true).
	// We can check if the list is initialized.
	is.True(model.List.Items() != nil) // Accessing exported List field
}

func TestCategoryProductsModel_Update_ProductsForCategoryResponseMsg(t *testing.T) {
	is := is.New(t)
	model := categoryproducts.NewModel("Test Cat", 1)
	testProducts := &commands.Products{{ID: 1, Title: "Product 1"}}
	msg := commands.ProductsForCategoryResponseMsg{
		Products:     testProducts,
		CategoryID:   1,
		CategoryName: "Test Cat",
		TotalItems:   30,
		TotalPages:   3,
	}
	updatedModel, _ := model.Update(msg)
	cpm, ok := updatedModel.(categoryproducts.Model)
	if !ok {
		t.Fatalf("expected categoryproducts.Model, got %T", updatedModel)
	}

	is.Equal(cpm.TotalItems, 30)
	is.Equal(cpm.TotalPages, 3)
	is.Equal(len(cpm.List.Items()), 1)
	// The list renders from the start of the page it was handed; the global
	// page number lives in PaginatedModel, not in the list's own paginator.
	is.Equal(cpm.List.Paginator.Page, 0)
}

func TestCategoryProductsModel_Update_ProductsForCategoryResponseMsg_Error(t *testing.T) {
	is := is.New(t)
	model := categoryproducts.NewModel("Test Cat", 1)

	msg := commands.ProductsForCategoryResponseMsg{
		Products:     nil,
		CategoryID:   1,
		CategoryName: "Test Cat",
		Err:          fmt.Errorf("API error"),
	}
	updatedModel, _ := model.Update(msg)
	cpm, ok := updatedModel.(categoryproducts.Model)
	if !ok {
		t.Fatalf("expected categoryproducts.Model, got %T", updatedModel)
	}

	is.Equal(cpm.ErrMsg, "API error")

	// View should display the error
	view := cpm.View()
	is.True(view.Content != "")
}

func TestCategoryProductsModel_Update_PageNavigation(t *testing.T) {
	// is := is.New(t) // is created per sub-test

	// Initial setup common for page navigation tests
	initialCategoryName := "Test Cat"
	initialCategoryID := 1
	initialPerPage := 10

	baseModel := categoryproducts.NewModel(initialCategoryName, initialCategoryID)
	baseModel.CurrentPage = 2
	baseModel.TotalPages = 3
	baseModel.PerPage = initialPerPage

	t.Run("next page", func(t *testing.T) {
		is := is.New(t)
		model := baseModel // Use baseModel as starting point for this sub-test
		updatedModel, cmd := model.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
		cpm, ok := updatedModel.(categoryproducts.Model)
		if !ok {
			t.Fatalf("expected categoryproducts.Model, got %T", updatedModel)
		}

		is.Equal(cpm.CurrentPage, 3)
		is.True(cmd != nil)
		pageMsg, ok := cmd().(commands.LoadCategoryProductsPageMsg)
		if !ok {
			t.Fatalf("expected LoadCategoryProductsPageMsg, got %T", cmd())
		}
		is.Equal(pageMsg.Page, 3)
		is.Equal(pageMsg.PerPage, initialPerPage)
		is.Equal(pageMsg.CategoryID, initialCategoryID)
		is.Equal(pageMsg.CategoryName, initialCategoryName)
	})

	t.Run("next page on last page", func(t *testing.T) {
		is := is.New(t)
		lastPageModel := baseModel
		lastPageModel.CurrentPage = 3 // Explicitly set to last page
		_, cmd := lastPageModel.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
		is.True(cmd == nil)
	})

	t.Run("previous page", func(t *testing.T) {
		is := is.New(t)
		model := baseModel // model starts at CurrentPage=2
		updatedModel, cmd := model.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
		cpm, ok := updatedModel.(categoryproducts.Model)
		if !ok {
			t.Fatalf("expected categoryproducts.Model, got %T", updatedModel)
		}

		is.Equal(cpm.CurrentPage, 1)
		is.True(cmd != nil)
		pageMsg, ok := cmd().(commands.LoadCategoryProductsPageMsg)
		if !ok {
			t.Fatalf("expected LoadCategoryProductsPageMsg, got %T", cmd())
		}
		is.Equal(pageMsg.Page, 1)
		is.Equal(pageMsg.PerPage, initialPerPage)
		is.Equal(pageMsg.CategoryID, initialCategoryID)
		is.Equal(pageMsg.CategoryName, initialCategoryName)
	})

	t.Run("previous page on first page", func(t *testing.T) {
		is := is.New(t)
		firstPageModel := baseModel
		firstPageModel.CurrentPage = 1 // Explicitly set to first page
		_, cmd := firstPageModel.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
		is.True(cmd == nil)
	})
}

// TestCategoryProductsModel_View_LaterPage reproduces the crash from the demo
// gif: the view died with "slice bounds out of range [20:10]" because the list
// was told it was on page 3 (Paginator.Page = 2) while holding a single page
// of ten items, so it tried to slice items[20:10].
func TestCategoryProductsModel_View_LaterPage(t *testing.T) {
	is := is.New(t)

	model := categoryproducts.NewModel("Test Cat", 1)
	model.CurrentPage = 3
	model.PerPage = 10
	model.TotalPages = 4

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 80})
	model, ok := updated.(categoryproducts.Model)
	is.True(ok)

	updated, _ = model.Update(commands.ProductsForCategoryResponseMsg{
		Products:     onePageOfProducts(),
		CategoryID:   1,
		CategoryName: "Test Cat",
		TotalItems:   34,
		TotalPages:   4,
	})
	model, ok = updated.(categoryproducts.Model)
	is.True(ok)

	// A panic here is the [20:10] crash: the list sliced a global page window
	// out of a single page of items.
	var view string
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("View() panicked: %v", r)
			}
		}()
		view = model.View().Content
	}()

	// The list holds only this page's items, so it must render all of them.
	is.True(strings.Contains(view, "Product 1"))
	is.True(strings.Contains(view, "Product 10"))
	is.True(strings.Contains(view, "Page 3/4"))
}

// onePageOfProducts returns the ten products a single API page holds.
func onePageOfProducts() *commands.Products {
	products := make(commands.Products, 10)
	for i := range products {
		products[i] = commands.Product{
			ID:     i + 1,
			Title:  fmt.Sprintf("Product %d", i+1),
			Prices: commands.ProductPrices{Price: "420", CurrencyPrefix: "£", CurrencyMinorUnit: 2},
		}
	}
	return &products
}
