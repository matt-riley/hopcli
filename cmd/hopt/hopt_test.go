package hopt_test

import (
	"errors"
	"fmt"
	"reflect" // Added reflect import
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/cmd/hopt"
	"github.com/matt-riley/hopcli/internal/categoryproducts" // Added import
	"github.com/matt-riley/hopcli/internal/commands"
	"github.com/matt-riley/hopcli/internal/default"
	"github.com/matt-riley/hopcli/internal/latest" // Added import
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

		products := &commands.Products{{
			ID: 1,
			Title: struct {
				Rendered string `json:"rendered"`
			}{Rendered: "Test Beer"},
			Description: struct {
				Rendered string `json:"rendered"`
			}{Rendered: "%%%A tasty sample beer description."},
		}}
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
		is.True(reflect.ValueOf(cmd).Pointer() == reflect.ValueOf(tea.Quit).Pointer())
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

	t.Run("should handle LoadLatestPageMsg", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Transition to LatestView first to simulate being in that view
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{})
		m = updatedModel.(hopt.MainModel)
		updatedModel, _ = m.Update(commands.LatestResponseMsg{Products: &commands.Products{}, TotalItems: 20, TotalPages: 2})
		m = updatedModel.(hopt.MainModel) // Now in LatestView

		initialPreviousViewsLen := len(m.PreviousViews)

		updatedModel, cmd := m.Update(commands.LoadLatestPageMsg{Page: 2, PerPage: 5})
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.Loading, true)
		is.True(cmd != nil)
		is.Equal(m.State, hopt.LatestView)                      // State should remain LatestView
		is.Equal(len(m.PreviousViews), initialPreviousViewsLen) // PreviousViews stack should not change
	})

	t.Run("should handle LoadCategoryProductsPageMsg", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Transition to CategoryProductsView first
		updatedModel, _ := m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "Test", APIEndpoint: "test/ep"})
		m = updatedModel.(hopt.MainModel)
		updatedModel, _ = m.Update(commands.ProductsForCategoryResponseMsg{CategoryID: 1, CategoryName: "Test", APIEndpoint: "test/ep", Products: &commands.Products{}, TotalItems: 20, TotalPages: 2})
		m = updatedModel.(hopt.MainModel) // Now in CategoryProductsView for CategoryID 1

		initialPreviousViewsLen := len(m.PreviousViews)

		updatedModel, cmd := m.Update(commands.LoadCategoryProductsPageMsg{CategoryID: 1, CategoryName: "Test", APIEndpoint: "test/ep", Page: 2, PerPage: 5})
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.Loading, true)
		is.True(cmd != nil)
		is.Equal(m.State, hopt.CategoryProductsView)            // State should remain CategoryProductsView
		is.Equal(len(m.PreviousViews), initialPreviousViewsLen) // PreviousViews stack should not change
		is.Equal(m.CategoryProductsModel.CategoryID(), 1)       // Still the same category
	})

	t.Run("should update LatestModel on LatestResponseMsg when already in LatestView", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Initial transition to LatestView
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{})
		m = updatedModel.(hopt.MainModel)
		updatedModel, _ = m.Update(commands.LatestResponseMsg{Products: &commands.Products{{ID: 1}}, TotalItems: 10, TotalPages: 1})
		m = updatedModel.(hopt.MainModel)

		initialPreviousViewsLen := len(m.PreviousViews)
		// Store a property of the model before the update to check if it's the same instance
		// For example, if LatestModel had an accessible field or we check CurrentPage if it's not reset by Update
		// Here, we check if CurrentView which points to m.LatestModel is the same instance after update.
		// Note: This is tricky as the model itself is a struct, so CurrentView holds a copy.
		// A better check is if the fields within m.LatestModel are updated.

		productsPage2 := &commands.Products{{ID: 2, Title: struct {
			Rendered string `json:"rendered"`
		}{Rendered: "Beer Page 2"}}}
		msg := commands.LatestResponseMsg{Products: productsPage2, TotalItems: 20, TotalPages: 2, Width: 80, Height: 24}

		// Capture current LatestModel's CurrentPage before this update for comparison if it's changed by the Update call
		// This relies on LatestModel.Update correctly setting these from the msg.
		// Let's assume for this test LatestModel.Update updates its own fields from msg.

		updatedModel, _ = m.Update(msg)
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.Loading, false)
		is.Equal(m.State, hopt.LatestView)
		is.Equal(len(m.PreviousViews), initialPreviousViewsLen) // Stack should not grow
		is.Equal(m.LatestModel.TotalPages, 2)                   // Check if model data updated
		_, ok := m.CurrentView.(latest.LatestModel)
		is.True(ok) // Corrected type assertion
	})

	t.Run("should update CategoryProductsModel on ProductsForCategoryResponseMsg when in same CategoryProductsView", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()

		// Initial transition to CategoryProductsView for CategoryID 123
		updatedModel, _ := m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 123, CategoryName: "Test Cat", APIEndpoint: "test/ep"})
		m = updatedModel.(hopt.MainModel)
		updatedModel, _ = m.Update(commands.ProductsForCategoryResponseMsg{
			CategoryID: 123, CategoryName: "Test Cat", APIEndpoint: "test/ep",
			Products: &commands.Products{{ID: 1}}, TotalItems: 10, TotalPages: 1,
		})
		m = updatedModel.(hopt.MainModel)

		initialPreviousViewsLen := len(m.PreviousViews)

		productsPage2 := &commands.Products{{ID: 2}}
		msg := commands.ProductsForCategoryResponseMsg{
			CategoryID: 123, CategoryName: "Test Cat", APIEndpoint: "test/ep",
			Products: productsPage2, TotalItems: 20, TotalPages: 2, Width: 80, Height: 24,
		}
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.Loading, false)
		is.Equal(m.State, hopt.CategoryProductsView)
		is.Equal(len(m.PreviousViews), initialPreviousViewsLen)
		is.Equal(m.CategoryProductsModel.CategoryID(), 123) // Use getter, Still same category
		is.Equal(m.CategoryProductsModel.TotalPages, 2)     // Data updated
		_, ok := m.CurrentView.(categoryproducts.Model)
		is.True(ok) // Corrected type assertion
	})

	t.Run("should create new CategoryProductsModel on ProductsForCategoryResponseMsg for a new category", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()

		// Setup initial state: in CategoryProductsView for "Old Cat"
		updatedModel, _ := m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 111, CategoryName: "Old Cat", APIEndpoint: "old/ep"})
		m = updatedModel.(hopt.MainModel)
		updatedModel, _ = m.Update(commands.ProductsForCategoryResponseMsg{
			CategoryID: 111, CategoryName: "Old Cat", APIEndpoint: "old/ep",
			Products: &commands.Products{{ID: 1}}, TotalItems: 5, TotalPages: 1,
		})
		m = updatedModel.(hopt.MainModel)
		is.Equal(m.CategoryProductsModel.CategoryID(), 111) // Pre-condition check, use getter

		initialPreviousViewsLen := len(m.PreviousViews)

		// Action: Dispatch response for a *different* category
		newCategoryProducts := &commands.Products{{ID: 100}}
		msg := commands.ProductsForCategoryResponseMsg{
			CategoryID: 999, CategoryName: "New Cat", APIEndpoint: "new/ep",
			Products: newCategoryProducts, TotalItems: 10, TotalPages: 1, Width: 80, Height: 24,
		}
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(hopt.MainModel)

		is.Equal(m.Loading, false)
		is.Equal(m.State, hopt.CategoryProductsView)
		is.Equal(m.CategoryProductsModel.CategoryID(), 999)         // Use getter, Verify it's the new model's data
		is.Equal(m.CategoryProductsModel.CategoryName(), "New Cat") // Use getter
		is.Equal(len(m.PreviousViews), initialPreviousViewsLen+1)   // Stack should grow
		_, ok := m.CurrentView.(categoryproducts.Model)
		is.True(ok) // Corrected type assertion
	})
}
