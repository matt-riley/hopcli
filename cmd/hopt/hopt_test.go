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
	m := hopt.InitialModel()

	is.Equal(m.State, hopt.DefaultView)
	is.True(m.CurrentView != nil) // Check that CurrentView is not nil
	is.Equal(m.Loading, false)
	is.Equal(m.ErrMsg, "")
}

func TestMainModelUpdate_StateTransitions(t *testing.T) {
	is := is.New(t)

	t.Run("should transition to loading on StartLoadingLatestMsg", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		newModel, cmd := m.Update(defaultview.StartLoadingLatestMsg{})

		mm := newModel.(hopt.MainModel)
		is.Equal(mm.Loading, true)
		is.True(cmd != nil) // Should return a command
	})

	t.Run("should transition to latestView on LatestResponseMsg success", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		m, _ = m.Update(defaultview.StartLoadingLatestMsg{}) // Start loading

		products := &commands.Products{{ID: 1, Title: commands.ProductRendered{Rendered: "Test Beer"}}}
		msg := commands.LatestResponseMsg{Products: products, Width: 80, Height: 24}
		newModel, _ := m.Update(msg)

		mm := newModel.(hopt.MainModel)
		is.Equal(mm.Loading, false)
		is.Equal(mm.State, hopt.LatestView)
		is.Equal(mm.ErrMsg, "")
		is.True(mm.CurrentView != nil)
		// Check if CurrentView is of type latest.LatestModel (or whatever it's called)
		// _, ok := mm.CurrentView.(latest.LatestModel)
		// is.True(ok) // CurrentView is LatestModel
	})

	t.Run("should set error message on LatestResponseMsg failure", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		originalState := m.State // Store original state before loading
		m, _ = m.Update(defaultview.StartLoadingLatestMsg{}) // Start loading

		errMsgContent := "network error"
		msg := commands.LatestResponseMsg{Err: fmt.Errorf(errMsgContent)}
		newModel, _ := m.Update(msg)

		mm := newModel.(hopt.MainModel)
		is.Equal(mm.Loading, false)
		is.Equal(mm.ErrMsg, errMsgContent)
		is.Equal(mm.State, originalState) // Should remain in the state it was in before loading started
	})

	t.Run("should transition to loading on StartLoadingCategoriesMsg", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		newModel, cmd := m.Update(defaultview.StartLoadingCategoriesMsg{})

		mm := newModel.(hopt.MainModel)
		is.Equal(mm.Loading, true)
		is.True(cmd != nil)
	})

	t.Run("should transition to categoriesView on CategoriesResponseMsg success", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		m, _ = m.Update(defaultview.StartLoadingCategoriesMsg{}) // Start loading

		categories := &commands.Categories{{ID: 1, Name: "Test Category"}}
		msg := commands.CategoriesResponseMsg{Categories: categories, Width: 80, Height: 24}
		newModel, _ := m.Update(msg)

		mm := newModel.(hopt.MainModel)
		is.Equal(mm.Loading, false)
		is.Equal(mm.State, hopt.CategoriesView)
		is.Equal(mm.ErrMsg, "")
		is.True(mm.CurrentView != nil)
		// _, ok := mm.CurrentView.(categories.Model)
		// is.True(ok) // CurrentView is CategoriesModel
	})

	t.Run("should set error message on CategoriesResponseMsg failure", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		originalState := m.State
		m, _ = m.Update(defaultview.StartLoadingCategoriesMsg{}) // Start loading

		errMsgContent := "categories fetch error"
		msg := commands.CategoriesResponseMsg{Err: fmt.Errorf(errMsgContent)}
		newModel, _ := m.Update(msg)

		mm := newModel.(hopt.MainModel)
		is.Equal(mm.Loading, false)
		is.Equal(mm.ErrMsg, errMsgContent)
		is.Equal(mm.State, originalState)
	})

	t.Run("should transition to loading on StartLoadingProductsForCategoryMsg", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Simulate being in categories view first
		m.State = hopt.CategoriesView
		// m.CurrentView = categories.NewCategoriesModel() // Or however it's initialized

		msg := commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "Test", APIEndpoint: "some/api", Width: 80, Height: 24}
		newModel, cmd := m.Update(msg)

		mm := newModel.(hopt.MainModel)
		is.Equal(mm.Loading, true)
		is.True(cmd != nil)
	})

	t.Run("should transition to categoryProductsView on ProductsForCategoryResponseMsg success", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Simulate loading products for a category
		m, _ = m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "Test", APIEndpoint: "some/api", Width: 80, Height: 24})

		products := &commands.Products{{ID: 1, Title: commands.ProductRendered{Rendered: "Test Beer in Category"}}}
		msg := commands.ProductsForCategoryResponseMsg{Products: products, CategoryID: 1, CategoryName: "Test", Width: 80, Height: 24}
		newModel, _ := m.Update(msg)

		mm := newModel.(hopt.MainModel)
		is.Equal(mm.Loading, false)
		is.Equal(mm.State, hopt.CategoryProductsView)
		is.Equal(mm.ErrMsg, "")
		is.True(mm.CurrentView != nil)
		// _, ok := mm.CurrentView.(categoryproducts.Model)
		// is.True(ok) // CurrentView is CategoryProductsModel
	})

	t.Run("should set error message on ProductsForCategoryResponseMsg failure", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Simulate we were in CategoriesView before trying to load products
		m.State = hopt.CategoriesView // Manually set the state we expect to return to
		m, _ = m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "Test", APIEndpoint: "some/api", Width: 80, Height: 24})

		errMsgContent := "products for category fetch error"
		msg := commands.ProductsForCategoryResponseMsg{Err: errors.New(errMsgContent), CategoryID: 1, CategoryName: "Test"}
		newModel, _ := m.Update(msg)

		mm := newModel.(hopt.MainModel)
		is.Equal(mm.Loading, false)
		is.Equal(mm.ErrMsg, errMsgContent)
		// The state should revert to what it was before loading started.
		// In the real app, this is handled by not changing the current view when an error occurs.
		// The test setup here for originalState is a bit tricky because Update() itself sets the new state.
		// For this specific test, if an error occurs, the state *within the Update function*
		// doesn't change from what it was when the loading message was processed.
		// The MainModel's behavior on error is to set mm.ErrMsg and not change the view/state.
		// So, if it was DefaultView, then StartLoading, then error, it should still be DefaultView.
		// If it was CategoriesView, then StartLoadingProducts, then error, it should still be CategoriesView.
		// The current test for failure cases like `LatestResponseMsg` failure already checks this.
		is.Equal(mm.State, hopt.CategoriesView) // Expect to remain in the view that initiated the loading if error
	})

	t.Run("should return Quit command on 'q' key", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		// Check if cmd is tea.Quit
		// Direct comparison might not work if tea.Quit is a function/struct.
		// Instead, we check if the command message is of type tea.QuitMsg (if that's how bubbletea signals quit)
		// or if the command itself is the tea.Quit sentinel.
		// For bubbletea, tea.Quit is a function that returns a tea.Cmd.
		// The command returned by Update should be the same as tea.Quit.
		is.Equal(cmd, tea.Quit)
	})

	t.Run("should navigate back on 'h' key when previous view exists", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel() // State is DefaultView

		// 1. Go to LatestView
		m, _ = m.Update(defaultview.StartLoadingLatestMsg{})
		m, _ = m.Update(commands.LatestResponseMsg{Products: &commands.Products{}, Width: 80, Height: 24})
		mm := m.(hopt.MainModel)
		is.Equal(mm.State, hopt.LatestView)       // Pre-condition: current state is LatestView
		is.Equal(len(mm.PreviousViews), 1) // One view (DefaultModel) should be in PreviousViews

		// 2. Press 'h' to go back
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		mm = m.(hopt.MainModel)
		is.Equal(mm.State, hopt.DefaultView) // Should be back to DefaultView
		is.Equal(len(mm.PreviousViews), 0)   // PreviousViews stack should be empty
		_, ok := mm.CurrentView.(defaultview.DefaultModel)
		is.True(ok) // CurrentView is DefaultModel

		// 3. Test another level of back navigation: Default -> Categories -> ProductsForCategory -> Categories
		m = hopt.InitialModel() // Reset to DefaultView
		m, _ = m.Update(defaultview.StartLoadingCategoriesMsg{})
		m, _ = m.Update(commands.CategoriesResponseMsg{Categories: &commands.Categories{{ID:1, Name:"cat1"}}, Width: 80, Height: 24})
		mm = m.(hopt.MainModel)
		is.Equal(mm.State, hopt.CategoriesView)
		is.Equal(len(mm.PreviousViews), 1) // DefaultView is previous

		m, _ = m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "cat1", APIEndpoint: "test/ep", Width: 80, Height: 24})
		m, _ = m.Update(commands.ProductsForCategoryResponseMsg{Products: &commands.Products{}, CategoryID: 1, CategoryName: "cat1", Width: 80, Height: 24})
		mm = m.(hopt.MainModel)
		is.Equal(mm.State, hopt.CategoryProductsView)
		is.Equal(len(mm.PreviousViews), 2) // DefaultView, CategoriesView are previous

		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}) // Back to CategoriesView
		mm = m.(hopt.MainModel)
		is.Equal(mm.State, hopt.CategoriesView)
		is.Equal(len(mm.PreviousViews), 1) // DefaultView is previous
		
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}) // Back to DefaultView
		mm = m.(hopt.MainModel)
		is.Equal(mm.State, hopt.DefaultView)
		is.Equal(len(mm.PreviousViews), 0)
	})

	t.Run("should not navigate back on 'h' key when no previous view exists", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		is.Equal(m.State, hopt.DefaultView) // Pre-condition
		is.Equal(len(m.PreviousViews), 0)   // Pre-condition

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		mm := newModel.(hopt.MainModel)
		is.Equal(mm.State, hopt.DefaultView) // Should remain in DefaultView
		is.Equal(len(mm.PreviousViews), 0)   // PreviousViews stack should still be empty
	})
}
