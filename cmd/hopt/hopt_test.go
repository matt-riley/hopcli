package hopt_test

import (
	"errors"
	"reflect" // Added reflect import
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/cmd/hopt"
	"github.com/matt-riley/hopcli/internal/categories"
	"github.com/matt-riley/hopcli/internal/categoryproducts" // Added import
	"github.com/matt-riley/hopcli/internal/commands"
	defaultview "github.com/matt-riley/hopcli/internal/default"
	"github.com/matt-riley/hopcli/internal/latest" // Added import
)

type countingModel struct{ msgs []tea.Msg }

func (m countingModel) Init() tea.Cmd { return nil }
func (m countingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.msgs = append(m.msgs, msg)
	return m, nil
}
func (m countingModel) View() tea.View { return tea.NewView("") }

func assertMainModel(t *testing.T, model tea.Model) hopt.MainModel {
	t.Helper()
	m, ok := model.(hopt.MainModel)
	if !ok {
		t.Fatalf("expected hopt.MainModel, got %T", model)
	}
	return m
}

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
		m = assertMainModel(t, updatedModel) // Apply type assertion

		is.Equal(m.Loading, true)
		is.True(cmd != nil) // Should return a command
	})

	t.Run("should transition to latestView on LatestResponseMsg success", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{}) // Start loading
		m = assertMainModel(t, updatedModel)

		products := &commands.Products{{
			ID:          1,
			Title:       "Test Beer",
			Description: "A tasty sample beer description.",
		}}
		msg := commands.LatestResponseMsg{Products: products, Width: 80, Height: 24, RequestID: 1}
		updatedModel, _ = m.Update(msg)
		m = assertMainModel(t, updatedModel)

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
		m = assertMainModel(t, updatedModel)

		errMsgContent := "network error"
		msg := commands.LatestResponseMsg{Err: errors.New(errMsgContent), RequestID: 1}
		updatedModel, _ = m.Update(msg)
		m = assertMainModel(t, updatedModel)

		is.Equal(m.Loading, false)
		is.Equal(m.ErrMsg, errMsgContent)
		is.Equal(m.State, originalState) // Should remain in the state it was in before loading started
	})

	t.Run("should transition to loading on StartLoadingCategoriesMsg", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		updatedModel, cmd := m.Update(defaultview.StartLoadingCategoriesMsg{})
		m = assertMainModel(t, updatedModel)

		is.Equal(m.Loading, true)
		is.True(cmd != nil)
	})

	t.Run("should transition to categoriesView on CategoriesResponseMsg success", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		updatedModel, _ := m.Update(defaultview.StartLoadingCategoriesMsg{}) // Start loading
		m = assertMainModel(t, updatedModel)

		categoriesData := &commands.Categories{{ID: 1, Name: "Test Category"}}
		msg := commands.CategoriesResponseMsg{Categories: categoriesData, RequestID: 1}
		updatedModel, _ = m.Update(msg)
		m = assertMainModel(t, updatedModel)

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
		m = assertMainModel(t, updatedModel)

		errMsgContent := "categories fetch error"
		msg := commands.CategoriesResponseMsg{Err: errors.New(errMsgContent), RequestID: 1}
		updatedModel, _ = m.Update(msg)
		m = assertMainModel(t, updatedModel)

		is.Equal(m.Loading, false)
		is.Equal(m.ErrMsg, errMsgContent)
		is.Equal(m.State, originalState)
	})

	t.Run("should transition to loading on StartLoadingProductsForCategoryMsg", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Simulate being in categories view first
		m.State = hopt.CategoriesView // Access exported field

		msg := commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "Test", Width: 80, Height: 24}
		updatedModel, cmd := m.Update(msg)
		m = assertMainModel(t, updatedModel)

		is.Equal(m.Loading, true)
		is.True(cmd != nil)
	})

	t.Run("should transition to categoryProductsView on ProductsForCategoryResponseMsg success", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Simulate loading products for a category (requires being in CategoriesView first)
		m.State = hopt.CategoriesView
		updatedModel, _ := m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "Test", Width: 80, Height: 24})
		m = assertMainModel(t, updatedModel)

		products := &commands.Products{{ID: 1, Title: "Test Beer in Category"}}
		msg := commands.ProductsForCategoryResponseMsg{Products: products, CategoryID: 1, CategoryName: "Test", RequestID: 1}
		updatedModel, _ = m.Update(msg)
		m = assertMainModel(t, updatedModel)

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
		updatedModel, _ := m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "Test", Width: 80, Height: 24})
		m = assertMainModel(t, updatedModel)

		errMsgContent := "products for category fetch error"
		msg := commands.ProductsForCategoryResponseMsg{Err: errors.New(errMsgContent), CategoryID: 1, CategoryName: "Test", RequestID: 1}
		updatedModel, _ = m.Update(msg)
		m = assertMainModel(t, updatedModel)

		is.Equal(m.Loading, false)
		is.Equal(m.ErrMsg, errMsgContent)
		is.Equal(m.State, hopt.CategoriesView) // Expect to remain in the view that initiated the loading if error
	})

	t.Run("should return Quit command on 'q' key", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		_, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
		is.True(reflect.ValueOf(cmd).Pointer() == reflect.ValueOf(tea.Quit).Pointer())
	})

	t.Run("should navigate back on 'h' key when previous view exists", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel() // State is DefaultView

		// 1. Go to LatestView
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{})
		m = assertMainModel(t, updatedModel)
		updatedModel, _ = m.Update(commands.LatestResponseMsg{Products: &commands.Products{}, Width: 80, Height: 24, RequestID: 1})
		m = assertMainModel(t, updatedModel)

		is.Equal(m.State, hopt.LatestView)
		is.Equal(len(m.PreviousViews), 1)

		// 2. Press 'h' to go back
		updatedModel, _ = m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
		m = assertMainModel(t, updatedModel)
		is.Equal(m.State, hopt.DefaultView)
		is.Equal(len(m.PreviousViews), 0)
		_, ok := m.CurrentView.(defaultview.DefaultModel)
		is.True(ok) // CurrentView is DefaultModel

		// 3. Test another level of back navigation: Default -> Categories -> ProductsForCategory -> Categories
		m = hopt.InitialModel() // Reset to DefaultView
		updatedModel, _ = m.Update(defaultview.StartLoadingCategoriesMsg{})
		m = assertMainModel(t, updatedModel)
		updatedModel, _ = m.Update(commands.CategoriesResponseMsg{Categories: &commands.Categories{{ID: 1, Name: "cat1"}}, RequestID: 1})
		m = assertMainModel(t, updatedModel)
		is.Equal(m.State, hopt.CategoriesView)
		is.Equal(len(m.PreviousViews), 1)

		updatedModel, _ = m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "cat1", Width: 80, Height: 24, NavGen: 1})
		m = assertMainModel(t, updatedModel)
		updatedModel, _ = m.Update(commands.ProductsForCategoryResponseMsg{Products: &commands.Products{}, CategoryID: 1, CategoryName: "cat1", RequestID: 2})
		m = assertMainModel(t, updatedModel)
		is.Equal(m.State, hopt.CategoryProductsView)
		is.Equal(len(m.PreviousViews), 2)

		updatedModel, _ = m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"}) // Back to CategoriesView
		m = assertMainModel(t, updatedModel)
		is.Equal(m.State, hopt.CategoriesView)
		is.Equal(len(m.PreviousViews), 1)

		updatedModel, _ = m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"}) // Back to DefaultView
		m = assertMainModel(t, updatedModel)
		is.Equal(m.State, hopt.DefaultView)
		is.Equal(len(m.PreviousViews), 0)
	})

	t.Run("should not navigate back on 'h' key when no previous view exists", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		is.Equal(m.State, hopt.DefaultView) // Pre-condition
		is.Equal(len(m.PreviousViews), 0)   // Pre-condition

		updatedModel, _ := m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
		m = assertMainModel(t, updatedModel)
		is.Equal(m.State, hopt.DefaultView) // Should remain in DefaultView
		is.Equal(len(m.PreviousViews), 0)   // PreviousViews stack should still be empty
	})

	t.Run("should handle LoadLatestPageMsg", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Transition to LatestView first to simulate being in that view
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{})
		m = assertMainModel(t, updatedModel)
		updatedModel, _ = m.Update(commands.LatestResponseMsg{Products: &commands.Products{}, TotalItems: 20, TotalPages: 2, RequestID: 1})
		m = assertMainModel(t, updatedModel) // Now in LatestView

		initialPreviousViewsLen := len(m.PreviousViews)

		updatedModel, cmd := m.Update(commands.LoadLatestPageMsg{Page: 2, PerPage: 5, NavGen: 1})
		m = assertMainModel(t, updatedModel)

		// Debounce: Loading stays false until the debounce timer fires.
		is.Equal(m.Loading, false)                              // debounce doesn't set loading immediately
		is.True(cmd != nil)                                     // debounce timer is returned
		is.Equal(m.State, hopt.LatestView)                      // State should remain LatestView
		is.Equal(len(m.PreviousViews), initialPreviousViewsLen) // PreviousViews stack should not change
		// Fire the debounce timer — this should set Loading=true and issue the API call.
		debounceMsg := cmd()
		updatedModel, apiCmd := m.Update(debounceMsg)
		m = assertMainModel(t, updatedModel)
		is.Equal(m.Loading, true)
		is.True(apiCmd != nil) // HandleGetLatest cmd
	})

	t.Run("should handle LoadCategoryProductsPageMsg", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Transition to CategoryProductsView first (requires being in CategoriesView)
		m.State = hopt.CategoriesView
		updatedModel, _ := m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "Test"})
		m = assertMainModel(t, updatedModel)
		updatedModel, _ = m.Update(commands.ProductsForCategoryResponseMsg{CategoryID: 1, CategoryName: "Test", Products: &commands.Products{}, TotalItems: 20, TotalPages: 2, RequestID: 1})
		m = assertMainModel(t, updatedModel) // Now in CategoryProductsView for CategoryID 1

		initialPreviousViewsLen := len(m.PreviousViews)

		updatedModel, cmd := m.Update(commands.LoadCategoryProductsPageMsg{CategoryID: 1, CategoryName: "Test", Page: 2, PerPage: 5, NavGen: 1})
		m = assertMainModel(t, updatedModel)

		// Debounce: Loading stays false until the debounce timer fires.
		is.Equal(m.Loading, false)
		is.True(cmd != nil)                                     // debounceCmd is not nil
		is.Equal(m.State, hopt.CategoryProductsView)            // State should remain CategoryProductsView
		is.Equal(len(m.PreviousViews), initialPreviousViewsLen) // PreviousViews stack should not change
		is.Equal(m.CategoryProductsModel.CategoryID(), 1)       // Still the same category

		// Fire the debounce timer — this should set Loading=true and issue the API call.
		debounceMsg := cmd()
		updatedModel, apiCmd := m.Update(debounceMsg)
		m = assertMainModel(t, updatedModel)
		is.Equal(m.Loading, true)
		is.True(apiCmd != nil) // HandleGetProductsByCategory cmd
	})

	t.Run("should update LatestModel on LatestResponseMsg when already in LatestView", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// Initial transition to LatestView
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{})
		m = assertMainModel(t, updatedModel)
		updatedModel, _ = m.Update(commands.LatestResponseMsg{Products: &commands.Products{{ID: 1}}, TotalItems: 10, TotalPages: 1, RequestID: 1})
		m = assertMainModel(t, updatedModel)

		initialPreviousViewsLen := len(m.PreviousViews)
		// Store a property of the model before the update to check if it's the same instance
		// For example, if LatestModel had an accessible field or we check CurrentPage if it's not reset by Update
		// Here, we check if CurrentView which points to m.LatestModel is the same instance after update.
		// Note: This is tricky as the model itself is a struct, so CurrentView holds a copy.
		// A better check is if the fields within m.LatestModel are updated.

		productsPage2 := &commands.Products{{ID: 2, Title: "Beer Page 2"}}
		msg := commands.LatestResponseMsg{Products: productsPage2, TotalItems: 20, TotalPages: 2, Width: 80, Height: 24, RequestID: 1}

		// Simulate a page-change load being in-flight before the response arrives
		m.Loading = true
		updatedModel, _ = m.Update(msg)
		m = assertMainModel(t, updatedModel)

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

		// Initial transition to CategoryProductsView for CategoryID 123 (requires CategoriesView first)
		m.State = hopt.CategoriesView
		updatedModel, _ := m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 123, CategoryName: "Test Cat"})
		m = assertMainModel(t, updatedModel)
		updatedModel, _ = m.Update(commands.ProductsForCategoryResponseMsg{
			CategoryID: 123, CategoryName: "Test Cat",
			Products: &commands.Products{{ID: 1}}, TotalItems: 10, TotalPages: 1, RequestID: 1,
		})
		m = assertMainModel(t, updatedModel)

		initialPreviousViewsLen := len(m.PreviousViews)

		productsPage2 := &commands.Products{{ID: 2}}
		msg := commands.ProductsForCategoryResponseMsg{
			CategoryID: 123, CategoryName: "Test Cat",
			Products: productsPage2, TotalItems: 20, TotalPages: 2, RequestID: 1,
		}
		// Simulate a page-change load being in-flight before the response arrives
		m.Loading = true
		updatedModel, _ = m.Update(msg)
		m = assertMainModel(t, updatedModel)

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

		// Setup initial state: in CategoryProductsView for "Old Cat" (requires CategoriesView first)
		m.State = hopt.CategoriesView
		updatedModel, _ := m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 111, CategoryName: "Old Cat"})
		m = assertMainModel(t, updatedModel)
		updatedModel, _ = m.Update(commands.ProductsForCategoryResponseMsg{
			CategoryID: 111, CategoryName: "Old Cat",
			Products: &commands.Products{{ID: 1}}, TotalItems: 5, TotalPages: 1, RequestID: 1,
		})
		m = assertMainModel(t, updatedModel)
		is.Equal(m.CategoryProductsModel.CategoryID(), 111) // Pre-condition check, use getter

		initialPreviousViewsLen := len(m.PreviousViews)

		// Action: Dispatch response for a *different* category (simulating a new load in-flight)
		newCategoryProducts := &commands.Products{{ID: 100}}
		msg := commands.ProductsForCategoryResponseMsg{
			CategoryID: 999, CategoryName: "New Cat",
			Products: newCategoryProducts, TotalItems: 10, TotalPages: 1, RequestID: 1,
		}
		m.Loading = true
		updatedModel, _ = m.Update(msg)
		m = assertMainModel(t, updatedModel)

		is.Equal(m.Loading, false)
		is.Equal(m.State, hopt.CategoryProductsView)
		is.Equal(m.CategoryProductsModel.CategoryID(), 999)         // Use getter, Verify it's the new model's data
		is.Equal(m.CategoryProductsModel.CategoryName(), "New Cat") // Use getter
		is.Equal(len(m.PreviousViews), initialPreviousViewsLen+1)   // Stack should grow
		_, ok := m.CurrentView.(categoryproducts.Model)
		is.True(ok) // Corrected type assertion
	})

	t.Run("'h' key does not leak into restored view", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()

		// Place a counting stub as the previous view to restore
		stub := countingModel{}
		m.PreviousViews = []tea.Model{stub}
		m.State = hopt.LatestView // simulate being in a forward view

		// Press 'h' to pop back to the stub
		updatedModel, _ := m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
		m = assertMainModel(t, updatedModel)

		is.Equal(len(m.PreviousViews), 0)

		restoredStub, ok := m.CurrentView.(countingModel)
		is.True(ok)

		// Width=0 and Height=0 in tests, so the WindowSizeMsg guard won't fire.
		// The stub must have received no messages at all.
		is.Equal(len(restoredStub.msgs), 0)
	})

	t.Run("'left' key does not leak into restored view", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()

		stub := countingModel{}
		m.PreviousViews = []tea.Model{stub}
		m.State = hopt.LatestView

		updatedModel, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
		m = assertMainModel(t, updatedModel)

		is.Equal(len(m.PreviousViews), 0)

		restoredStub, ok := m.CurrentView.(countingModel)
		is.True(ok)

		// Width=0 and Height=0 in tests, so the WindowSizeMsg guard won't fire.
		is.Equal(len(restoredStub.msgs), 0)
	})

	t.Run("back nav resends WindowSizeMsg when dimensions are set", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		m.Width = 100
		m.Height = 50

		stub := countingModel{}
		m.PreviousViews = []tea.Model{stub}
		m.State = hopt.LatestView

		updatedModel, _ := m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
		m = assertMainModel(t, updatedModel)

		is.Equal(len(m.PreviousViews), 0)

		restoredStub, ok := m.CurrentView.(countingModel)
		is.True(ok)

		is.Equal(len(restoredStub.msgs), 1)
		wsMsg, ok := restoredStub.msgs[0].(tea.WindowSizeMsg)
		is.True(ok)
		is.Equal(wsMsg.Width, 100)
		is.Equal(wsMsg.Height, 50)
	})

	t.Run("back nav syncs typed CategoryProductsModel field with resized CurrentView", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		m.Width = 100
		m.Height = 50

		// Navigate back to CategoryProductsView: put CategoryProductsModel in PreviousViews.
		cpm := categoryproducts.NewModel("TestCat", 42)
		m.CategoryProductsModel = cpm
		m.PreviousViews = []tea.Model{cpm}
		m.State = hopt.ProductView // current state we're navigating away from

		updatedModel, _ := m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
		m = assertMainModel(t, updatedModel)

		is.Equal(m.State, hopt.CategoryProductsView)
		is.Equal(len(m.PreviousViews), 0)

		// CurrentView must be a categoryproducts.Model after back-nav.
		cvTyped, ok := m.CurrentView.(categoryproducts.Model)
		is.True(ok)

		// The resize was applied to CurrentView.
		is.True(cvTyped.List.Width() > 0)

		// The typed field must be in sync with CurrentView after the resize.
		// list.Model contains function fields that reflect.DeepEqual treats as unequal,
		// so compare the key observable dimensions rather than the whole struct.
		is.Equal(m.CategoryProductsModel.List.Width(), cvTyped.List.Width())
		is.Equal(m.CategoryProductsModel.List.Height(), cvTyped.List.Height())
	})

	t.Run("typed LatestModel field stays in sync with CurrentView after key update", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		m.State = hopt.LatestView
		m.LatestModel = latest.NewLatestModel()
		m.LatestModel.CurrentPage = 1
		m.CurrentView = m.LatestModel
		m.Loading = false

		updatedModel, _ := m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
		m = assertMainModel(t, updatedModel)

		latestFromView, ok := m.CurrentView.(latest.LatestModel)
		is.True(ok)
		is.Equal(m.LatestModel.CurrentPage, latestFromView.CurrentPage)
		is.Equal(m.LatestModel.TotalPages, latestFromView.TotalPages)
		is.Equal(m.LatestModel.PerPage, latestFromView.PerPage)
	})

	t.Run("ordinary key calls CurrentView.Update exactly once", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()

		stub := countingModel{}
		m.CurrentView = stub
		m.Loading = false

		updatedModel, _ := m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
		m = assertMainModel(t, updatedModel)

		restoredStub, ok := m.CurrentView.(countingModel)
		is.True(ok)
		is.Equal(len(restoredStub.msgs), 1)
	})

	t.Run("Enter on LatestView transitions to ProductView", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()

		// Load into LatestView with one product (requestID becomes 1 after StartLoadingLatestMsg)
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{})
		m = assertMainModel(t, updatedModel)

		products := &commands.Products{{
			ID:          1,
			Title:       "Test Beer",
			Description: "A tasty sample beer description.",
		}}
		updatedModel, _ = m.Update(commands.LatestResponseMsg{
			Products: products, Width: 80, Height: 24, RequestID: 1,
		})
		m = assertMainModel(t, updatedModel)
		is.Equal(m.State, hopt.LatestView)

		// Press Enter — LatestModel returns HandleDisplayProduct as a cmd
		updatedModel, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		m = assertMainModel(t, updatedModel)
		is.True(cmd != nil)

		// Execute the cmd to obtain the ProductsMsg, then feed it back
		if cmd != nil {
			productsMsg := cmd()
			updatedModel, _ = m.Update(productsMsg)
			m = assertMainModel(t, updatedModel)
		}

		is.Equal(m.State, hopt.ProductView)
	})

	t.Run("h key clears ErrMsg even when PreviousViews is empty", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		m.ErrMsg = "network error"
		// PreviousViews is empty by default

		updatedModel, _ := m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
		m = assertMainModel(t, updatedModel)

		is.Equal(m.ErrMsg, "")
		is.Equal(m.State, hopt.DefaultView)
	})

	t.Run("ProductsMsg ignored when not on list view", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// State is DefaultView — neither LatestView nor CategoryProductsView
		prod := &commands.Product{ID: 1, Title: "Test Beer"}
		updatedModel, _ := m.Update(commands.ProductsMsg{Product: prod})
		m = assertMainModel(t, updatedModel)

		is.Equal(m.State, hopt.DefaultView) // State must not change
	})

	t.Run("LoadLatestPageMsg ignored when not on LatestView", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// State is DefaultView — not LatestView

		updatedModel, _ := m.Update(commands.LoadLatestPageMsg{Page: 2, PerPage: 5})
		m = assertMainModel(t, updatedModel)

		is.Equal(m.State, hopt.DefaultView) // State must not change
		is.Equal(m.Loading, false)          // Must not start loading
	})

	t.Run("StartLoadingProductsForCategoryMsg ignored when not on CategoriesView", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// State is DefaultView — not CategoriesView

		updatedModel, _ := m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "Test"})
		m = assertMainModel(t, updatedModel)

		is.Equal(m.State, hopt.DefaultView) // State must not change
		is.Equal(m.Loading, false)          // Must not start loading
	})

	t.Run("LoadCategoryProductsPageMsg ignored when not on CategoryProductsView", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		// State is DefaultView — not CategoryProductsView

		updatedModel, _ := m.Update(commands.LoadCategoryProductsPageMsg{CategoryID: 1, CategoryName: "Test", Page: 2, PerPage: 5})
		m = assertMainModel(t, updatedModel)

		is.Equal(m.State, hopt.DefaultView) // State must not change
		is.Equal(m.Loading, false)          // Must not start loading
	})

	t.Run("ProductsMsg with stale NavGen is dropped even when state matches", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()

		// Get to LatestView (requestID becomes 1 after StartLoadingLatestMsg)
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{})
		m = assertMainModel(t, updatedModel)
		updatedModel, _ = m.Update(commands.LatestResponseMsg{Products: &commands.Products{}, RequestID: 1})
		m = assertMainModel(t, updatedModel)
		is.Equal(m.State, hopt.LatestView) // Pre-condition: in LatestView

		// Send ProductsMsg with stale NavGen (0 != mm.requestID which is 1)
		prod := &commands.Product{ID: 1, Title: "Stale Beer"}
		updatedModel, _ = m.Update(commands.ProductsMsg{Product: prod, NavGen: 0})
		m = assertMainModel(t, updatedModel)

		is.Equal(m.State, hopt.LatestView) // Must not transition to ProductView
	})

	t.Run("LoadLatestPageMsg with stale NavGen is dropped", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()

		// Get to LatestView (requestID becomes 1)
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{})
		m = assertMainModel(t, updatedModel)
		updatedModel, _ = m.Update(commands.LatestResponseMsg{Products: &commands.Products{}, RequestID: 1})
		m = assertMainModel(t, updatedModel)
		is.Equal(m.State, hopt.LatestView) // Pre-condition: in LatestView

		// Send LoadLatestPageMsg with stale NavGen (0 != mm.requestID which is 1)
		updatedModel, _ = m.Update(commands.LoadLatestPageMsg{Page: 2, PerPage: 5, NavGen: 0})
		m = assertMainModel(t, updatedModel)

		is.Equal(m.State, hopt.LatestView) // State must not change
		is.Equal(m.Loading, false)         // Must not start loading
	})

	t.Run("CategoriesResponseMsg re-dispatches WindowSizeMsg to new view", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		m.Width = 100
		m.Height = 50

		updatedModel, _ := m.Update(defaultview.StartLoadingCategoriesMsg{})
		m = assertMainModel(t, updatedModel)

		categoriesData := &commands.Categories{{ID: 1, Name: "Test Category"}}
		msg := commands.CategoriesResponseMsg{Categories: categoriesData, RequestID: 1}
		updatedModel, _ = m.Update(msg)
		m = assertMainModel(t, updatedModel)

		is.Equal(m.State, hopt.CategoriesView)

		cvTyped, ok := m.CurrentView.(categories.Model)
		is.True(ok)
		is.True(cvTyped.List.Width() > 0)
		is.Equal(m.CategoriesModel.List.Width(), cvTyped.List.Width())
		is.Equal(m.CategoriesModel.List.Height(), cvTyped.List.Height())
	})

	t.Run("ProductsForCategoryResponseMsg re-dispatches WindowSizeMsg to new view", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		m.Width = 100
		m.Height = 50
		m.State = hopt.CategoriesView

		updatedModel, _ := m.Update(commands.StartLoadingProductsForCategoryMsg{CategoryID: 1, CategoryName: "Test"})
		m = assertMainModel(t, updatedModel)

		products := &commands.Products{{ID: 1, Title: "Test Beer in Category"}}
		msg := commands.ProductsForCategoryResponseMsg{Products: products, CategoryID: 1, CategoryName: "Test", RequestID: 1}
		updatedModel, _ = m.Update(msg)
		m = assertMainModel(t, updatedModel)

		is.Equal(m.State, hopt.CategoryProductsView)

		cvTyped, ok := m.CurrentView.(categoryproducts.Model)
		is.True(ok)
		is.True(cvTyped.List.Width() > 0)
		is.Equal(m.CategoryProductsModel.List.Width(), cvTyped.List.Width())
		is.Equal(m.CategoryProductsModel.List.Height(), cvTyped.List.Height())
	})

	t.Run("LoadLatestPageMsg corrects CurrentPage after optimistic double-advance", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()

		// Get to LatestView (requestID becomes 1 after StartLoadingLatestMsg)
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{})
		m = assertMainModel(t, updatedModel)
		updatedModel, _ = m.Update(commands.LatestResponseMsg{
			Products: &commands.Products{}, TotalItems: 30, TotalPages: 3, RequestID: 1,
		})
		m = assertMainModel(t, updatedModel)
		is.Equal(m.State, hopt.LatestView)

		// Simulate the double-advance: user pressed 'n' twice quickly, so the optimistic
		// CurrentPage jumped to 3 even though we're only requesting page 2.
		m.LatestModel.CurrentPage = 3
		m.CurrentView = m.LatestModel

		// First LoadLatestPageMsg (Page:2, NavGen:1) arrives
		updatedModel, _ = m.Update(commands.LoadLatestPageMsg{Page: 2, PerPage: 10, NavGen: 1})
		m = assertMainModel(t, updatedModel)

		// The fix must correct CurrentPage to what is actually being fetched
		is.Equal(m.LatestModel.CurrentPage, 2)
		cvTyped, ok := m.CurrentView.(latest.LatestModel)
		is.True(ok)
		is.Equal(cvTyped.CurrentPage, 2)
	})

	t.Run("ProductsMsg re-dispatches WindowSizeMsg to ProductModel", func(t *testing.T) {
		is := is.New(t)
		m := hopt.InitialModel()
		m.Width = 100
		m.Height = 50

		// Get to LatestView (requestID becomes 1)
		updatedModel, _ := m.Update(defaultview.StartLoadingLatestMsg{})
		m = assertMainModel(t, updatedModel)
		products := &commands.Products{{ID: 1, Title: "Test Beer", Description: "A tasty sample beer."}}
		updatedModel, _ = m.Update(commands.LatestResponseMsg{Products: products, RequestID: 1})
		m = assertMainModel(t, updatedModel)
		is.Equal(m.State, hopt.LatestView)

		// Send ProductsMsg with matching NavGen (requestID is 1 after StartLoadingLatestMsg)
		prod := &commands.Product{ID: 1, Title: "Test Beer"}
		updatedModel, _ = m.Update(commands.ProductsMsg{Product: prod, NavGen: 1})
		m = assertMainModel(t, updatedModel)

		is.Equal(m.State, hopt.ProductView)
		is.True(m.ProductModel.Width > 0)
		is.Equal(m.ProductModel.Width, 100)
	})
}
