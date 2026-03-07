package categoryproducts_test

import (
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
		Width:        80,
		Height:       24,
	}
	updatedModel, _ := model.Update(msg)
	cpm := updatedModel.(categoryproducts.Model)

	is.Equal(cpm.TotalItems, 30)
	is.Equal(cpm.TotalPages, 3)
	is.Equal(len(cpm.List.Items()), 1)
	is.Equal(cpm.List.Paginator.Page, 0) // CurrentPage is 1, Paginator.Page is 0-indexed
	is.Equal(cpm.List.Paginator.PerPage, 10)
	is.Equal(cpm.List.Paginator.TotalPages, 3)
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
		cpm := updatedModel.(categoryproducts.Model)

		is.Equal(cpm.CurrentPage, 3)
		is.True(cmd != nil)
		pageMsg := cmd().(commands.LoadCategoryProductsPageMsg)
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
		cpm := updatedModel.(categoryproducts.Model)

		is.Equal(cpm.CurrentPage, 1)
		is.True(cmd != nil)
		pageMsg := cmd().(commands.LoadCategoryProductsPageMsg)
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
