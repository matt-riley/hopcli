package latest_test

import (
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
	is.Equal(lm.Choices.Paginator.Page, 0) // CurrentPage is 1, Paginator.Page is 0-indexed
	is.Equal(lm.Choices.Paginator.PerPage, 10)
	is.Equal(lm.Choices.Paginator.TotalPages, 3)
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
