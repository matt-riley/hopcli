package hopt_test

import (
	"errors"
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/cmd/hopt"
	"github.com/matt-riley/hopcli/internal/commands"
	"github.com/matt-riley/hopcli/internal/default"
	// Import other necessary internal packages if their types are directly used in MainModel fields
	// "github.com/matt-riley/hopcli/internal/latest"
	// "github.com/matt-riley/hopcli/internal/categories"
	// "github.com/matt-riley/hopcli/internal/categoryproducts"
	// "github.com/matt-riley/hopcli/internal/product"
)

func TestInitialModel(t *testing.T) {
	is := is.New(t)
	model := hopt.InitialModel()

	is.Equal(model.State, hopt.DefaultView)
	is.True(model.CurrentView != nil) // Check that CurrentView is not nil
	is.Equal(model.Loading, false)
	is.Equal(model.ErrMsg, "")
}

func TestMainModelUpdate_StateTransitions(t *testing.T) {
	// is := is.New(t) // is is created per sub-test

	t.Run("should transition to loading on StartLoadingLatestMsg", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		updatedModel, cmd := m.Update(defaultview.StartLoadingLatestMsg{})
		m = updatedModel.(hopt.MainModel) // Apply type assertion

		is.Equal(m.Loading, true)
		is.True(cmd != nil) // Should return a command
	})

	t.Run("should transition to latestView on LatestResponseMsg success", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{}) // Start loading
		m = updatedModel.(hopt.MainModel)

		products := &commands.Products{{ID: 1, Title: struct {
			Rendered string `json:"rendered"`
		}{Rendered: "Test Beer"}}}
		msg := commands.LatestResponseMsg{Products: products, Width: 80, Height: 24}
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.Loading, false)
		is.Equal(m.State, hopt.LatestView)
		is.Equal(m.ErrMsg, "")
		is.True(m.CurrentView != nil)
	})

	t.Run("should set error message on LatestResponseMsg failure", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		originalState := m.State                                         // Store original state before loading
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{}) // Start loading
		m = updatedModel.(hopt.MainModel)

		errMsgContent := "network error"
		msg := commands.LatestResponseMsg{Err: fmt.Errorf(errMsgContent)}
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.Loading, false)
		is.Equal(m.ErrMsg, errMsgContent)
		is.Equal(m.State, originalState) // Should remain in the state it was in before loading started
	})

	t.Run("should transition to loading on StartLoadingCategoriesMsg", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		updatedModel, cmd := m.Update(defaultview.StartLoadingCategoriesMsg{})
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.Loading, true)
		is.True(cmd != nil)
	})

	t.Run("should transition to categoriesView on CategoriesResponseMsg success", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		updatedModel, _ := m.Update(defaultview.StartLoadingCategoriesMsg{}) // Start loading
		m = updatedModel.(hopt.MainModel)

		categoriesData := &commands.Categories{{ID: 1, Name: "Test Category"}}
		msg := commands.CategoriesResponseMsg{Categories: categoriesData, Width: 80, Height: 24}
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.Loading, false)
		is.Equal(m.State, hopt.CategoriesView)
		is.Equal(m.ErrMsg, "")
		is.True(m.CurrentView != nil)
	})

	t.Run("should set error message on CategoriesResponseMsg failure", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		originalState := m.State
		updatedModel, _ := m.Update(defaultview.StartLoadingCategoriesMsg{}) // Start loading
		m = updatedModel.(hopt.MainModel)

		errMsgContent := "categories fetch error"
		msg := commands.CategoriesResponseMsg{Err: fmt.Errorf(errMsgContent)}
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.Loading, false)
		is.Equal(m.ErrMsg, errMsgContent)
		is.Equal(m.State, originalState)
	})

	t.Run("should transition to loading on StartLoadingProductsForCategoryMsg", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Simulate being in categories view first
		m.State = hopt.CategoriesView // Access exported field
		// m.CurrentView = categories.NewCategoriesModel() // Assuming this setup is covered by CategoriesResponseMsg test

		msg := commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "Test", APIEndpoint: "some/api", Width: 80, Height: 24}
		updatedModel, cmd := m.Update(msg)
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.Loading, true)
		is.True(cmd != nil)
	})

	t.Run("should transition to categoryProductsView on ProductsForCategoryResponseMsg success", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Simulate loading products for a category
		updatedModel, _ := m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "Test", APIEndpoint: "some/api", Width: 80, Height: 24})
		m = updatedModel.(hopt.MainModel)

		products := &commands.Products{{ID: 1, Title: struct {
			Rendered string `json:"rendered"`
		}{Rendered: "Test Beer in Category"}}}
		msg := commands.ProductsForCategoryResponseMsg{Products: products, CategoryID: 1, CategoryName: "Test", Width: 80, Height: 24}
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.Loading, false)
		is.Equal(m.State, hopt.CategoryProductsView)
		is.Equal(m.ErrMsg, "")
		is.True(m.CurrentView != nil)
	})

	t.Run("should set error message on ProductsForCategoryResponseMsg failure", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Simulate we were in CategoriesView before trying to load products
		m.State = hopt.CategoriesView // Manually set the state we expect to return to
		updatedModel, _ := m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "Test", APIEndpoint: "some/api", Width: 80, Height: 24})
		m = updatedModel.(hopt.MainModel)

		errMsgContent := "products for category fetch error"
		msg := commands.ProductsForCategoryResponseMsg{Err: errors.New(errMsgContent), CategoryID: 1, CategoryName: "Test"}
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.Loading, false)
		is.Equal(m.ErrMsg, errMsgContent)
		is.Equal(m.State, hopt.CategoriesView) // Expect to remain in the view that initiated the loading if error
	})

	t.Run("should return Quit command on 'q' key", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		is.Equal(cmd, tea.Quit)
	})

	t.Run("should navigate back on 'h' key when previous view exists", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel() // State is DefaultView

		// 1. Go to LatestView
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{})
		m = updatedModel.(hopt.MainModel)
		updatedModel, _ = m.Update(commands.LatestResponseMsg{Products: &commands.Products{}, Width: 80, Height: 24})
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.State, hopt.LatestView)
		is.Equal(len(m.PreviousViews), 1)

		// 2. Press 'h' to go back
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		m = updatedModel.(hopt.MainModel)
		is.Equal(m.State, hopt.DefaultView)
		is.Equal(len(m.PreviousViews), 0)
		_, ok := m.CurrentView.(defaultview.DefaultModel)
		is.True(ok) // CurrentView is DefaultModel

		// 3. Test another level of back navigation: Default -> Categories -> ProductsForCategory -> Categories
		m = hopt.InitialModel() // Reset to DefaultView
		updatedModel, _ = m.Update(defaultview.StartLoadingCategoriesMsg{})
		m = updatedModel.(hopt.MainModel)
		updatedModel, _ = m.Update(commands.CategoriesResponseMsg{Categories: &commands.Categories{{ID: 1, Name: "cat1"}}, Width: 80, Height: 24})
		m = updatedModel.(hopt.MainModel)
		is.Equal(m.State, hopt.CategoriesView)
		is.Equal(len(m.PreviousViews), 1)

		updatedModel, _ = m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "cat1", APIEndpoint: "test/ep", Width: 80, Height: 24})
		m = updatedModel.(hopt.MainModel)
		updatedModel, _ = m.Update(commands.ProductsForCategoryResponseMsg{Products: &commands.Products{}, CategoryID: 1, CategoryName: "cat1", Width: 80, Height: 24})
		m = updatedModel.(hopt.MainModel)
		is.Equal(m.State, hopt.CategoryProductsView)
		is.Equal(len(m.PreviousViews), 2)

		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}) // Back to CategoriesView
		m = updatedModel.(hopt.MainModel)
		is.Equal(m.State, hopt.CategoriesView)
		is.Equal(len(m.PreviousViews), 1)

		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}) // Back to DefaultView
		m = updatedModel.(hopt.MainModel)
		is.Equal(m.State, hopt.DefaultView)
		is.Equal(len(m.PreviousViews), 0)
	})

	t.Run("should not navigate back on 'h' key when no previous view exists", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		is.Equal(m.State, hopt.DefaultView) // Pre-condition
		is.Equal(len(m.PreviousViews), 0)   // Pre-condition

		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		m = updatedModel.(hopt.MainModel)
		is.Equal(m.State, hopt.DefaultView) // Should remain in DefaultView
		is.Equal(len(m.PreviousViews), 0)   // PreviousViews stack should still be empty
	})
}
