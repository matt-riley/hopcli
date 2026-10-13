package latest_test

import (
	"fmt"
	"strings"
	"testing"

	// "charm.land/bubbles/v2/list" // Not directly used for assertions, list.Model is in latest.LatestModel
	tea "charm.land/bubbletea/v2"
	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/internal/commands"
	"github.com/matt-riley/hopcli/internal/latest"
)

func TestNewLatestModel_InitialValues(t *testing.T) {
	is := is.New(t)
	model := latest.NewLatestModel()
	is.Equal(model.CurrentPage, 1)
	is.Equal(model.PerPage, 10)
	// For list.Model, Paginator.ShowPagination is not directly exported/accessible.
	// We rely on SetShowPagination(true) being called in NewLatestModel.
	// Similarly for ShowStatusBar. We can infer from default list behavior or if methods to check exist.
	// Since these are internal to the list component and set by its constructor or our SetShow methods,
	// we'll trust they are set. If specific assertions are needed, the list model might need getters.
	// For now, we'll assume the list is configured as intended by NewLatestModel's calls.
	// is.True(model.Choices.Paginator.ShowPagination) // This field is not exported
	// is.True(model.Choices.ShowStatusBar) // This field is not exported
	// We can check if the list itself is not nil
	is.True(model.Choices.Items() != nil) // Choices.items is initialized as an empty slice
}

func TestLatestModel_Update_LatestResponseMsg(t *testing.T) {
	is := is.New(t)
	model := latest.NewLatestModel()
	testProducts := &commands.Products{{ID: 1, Title: "Beer 1"}}
	msg := commands.LatestResponseMsg{Products: testProducts, TotalItems: 25, TotalPages: 3, Width: 80, Height: 24}
	updatedModel, _ := model.Update(msg)
	lm, ok := updatedModel.(latest.LatestModel)
	if !ok {
		t.Fatalf("expected latest.LatestModel, got %T", updatedModel)
	}

	is.Equal(lm.TotalItems, 25)
	is.Equal(lm.TotalPages, 3)
	is.Equal(len(lm.Choices.Items()), 1)
	// The list renders from the start of the page it was handed; the global
	// page number lives in PaginatedModel, not in the list's own paginator.
	is.Equal(lm.Choices.Paginator.Page, 0)
}

func TestLatestModel_Update_PageNavigation(t *testing.T) {
	// is := is.New(t) // is created per sub-test

	// Initial setup for all sub-tests in this group
	baseModel := latest.NewLatestModel()
	baseModel.CurrentPage = 2
	baseModel.TotalPages = 3
	baseModel.PerPage = 10
	// Manually set some dummy products for context, as Update might try to access them
	// Although for 'n' and 'p' key presses, it shouldn't matter if list items are empty.
	// baseModel.Choices.SetItems([]list.Item{latest.LatestListItem{Title: "Dummy"}})

	t.Run("next page", func(t *testing.T) {
		is := is.New(t)
		model := baseModel // Use a copy or the base model if state changes are isolated
		updatedModel, cmd := model.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
		lm, ok := updatedModel.(latest.LatestModel)
		if !ok {
			t.Fatalf("expected latest.LatestModel, got %T", updatedModel)
		}

		is.Equal(lm.CurrentPage, 3)
		is.True(cmd != nil)
		pageMsg, ok := cmd().(commands.LoadLatestPageMsg)
		if !ok {
			t.Fatalf("expected LoadLatestPageMsg, got %T", cmd())
		}
		is.Equal(pageMsg.Page, 3)
		is.Equal(pageMsg.PerPage, 10)
	})

	t.Run("next page on last page", func(t *testing.T) {
		is := is.New(t)
		lastPageModel := baseModel
		lastPageModel.CurrentPage = 3 // Explicitly set to last page
		updatedModel, cmd := lastPageModel.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
		_, ok := updatedModel.(latest.LatestModel)
		is.True(ok) // Ensure it's the correct type

		is.True(cmd == nil) // Should do nothing
	})

	t.Run("previous page", func(t *testing.T) {
		is := is.New(t)
		model := baseModel // model is CurrentPage=2 from initial setup
		updatedModel, cmd := model.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
		lm, ok := updatedModel.(latest.LatestModel)
		if !ok {
			t.Fatalf("expected latest.LatestModel, got %T", updatedModel)
		}

		is.Equal(lm.CurrentPage, 1)
		is.True(cmd != nil)
		pageMsg, ok := cmd().(commands.LoadLatestPageMsg)
		if !ok {
			t.Fatalf("expected LoadLatestPageMsg, got %T", cmd())
		}
		is.Equal(pageMsg.Page, 1)
		is.Equal(pageMsg.PerPage, 10) // Ensure PerPage is carried over
	})

	t.Run("previous page on first page", func(t *testing.T) {
		is := is.New(t)
		firstPageModel := baseModel
		firstPageModel.CurrentPage = 1 // Explicitly set to first page
		updatedModel, cmd := firstPageModel.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
		_, ok := updatedModel.(latest.LatestModel)
		is.True(ok) // Ensure correct type

		is.True(cmd == nil)
	})
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

// TestLatestModel_View_LaterPage reproduces the crash from the demo gif: the
// view died with "slice bounds out of range [20:10]" because the list was told
// it was on page 3 (Paginator.Page = 2) while holding a single page of ten
// items, so it tried to slice items[20:10].
func TestLatestModel_View_LaterPage(t *testing.T) {
	is := is.New(t)

	model := latest.NewLatestModel()
	model.CurrentPage = 3
	model.PerPage = 10
	model.TotalPages = 4

	updated, _ := model.Update(commands.LatestResponseMsg{
		Products:   onePageOfProducts(),
		Width:      120,
		Height:     80,
		TotalItems: 34,
		TotalPages: 4,
	})
	model, ok := updated.(latest.LatestModel)
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
