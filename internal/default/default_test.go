package defaultview_test

import (
	"fmt"
	"reflect" // Added reflect import
	"testing"

	// "charm.land/bubbles/v2/list" // Not directly used for assertions
	tea "charm.land/bubbletea/v2"
	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/internal/default" // To use defaultview.InitialModel, defaultview.DefaultModel, etc.
)

func TestDefaultInitialModel(t *testing.T) {
	is := is.New(t)
	model := defaultview.InitialModel()

	is.Equal(len(model.Choices.Items()), 2) // Should have two items: "Latest" and "Categories"
	items := model.Choices.Items()

	// Check first item ("Latest")
	latestItem, ok := items[0].(defaultview.ListItem)
	is.True(ok)
	is.Equal(latestItem.Title(), "Latest")
	is.Equal(latestItem.Description(), "Latest items added")

	// Check second item ("Categories")
	categoriesItem, ok := items[1].(defaultview.ListItem)
	is.True(ok)
	is.Equal(categoriesItem.Title(), "Categories")
	// Based on previous steps, this description was "Browse by category"
	// Let's use the value from the actual default.go:
	// items := []list.Item{
	// ListItem{title: "Latest", desc: "Latest items added"},
	// ListItem{title: "Categories", desc: "Browse by category"},
	// }
	// The subtask description had "Browse product categories", but the code has "Browse by category".
	// Using "Browse by category" from the actual code.
	is.Equal(categoriesItem.Description(), "Browse by category")
}

func TestDefaultUpdate_EnterKey(t *testing.T) {
	// is := is.New(t) // is for the main test - not used at this scope
	// model := defaultview.InitialModel() // model is initialized per sub-test

	t.Run("select Latest", func(t *testing.T) {
		is := is.New(t) // is for the sub-test
		// Reset model or ensure selection is clean for each sub-test if model state persists
		// For this model, simply selecting is fine as it doesn't change internal state beyond list selection
		m := defaultview.InitialModel() // Use a fresh model for each sub-test for isolation
		m.Choices.Select(0)             // Select "Latest"
		_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		is.True(cmd != nil)
		msg := cmd()
		_, ok := msg.(defaultview.StartLoadingLatestMsg)
		is.True(ok) // Check if the message is of the correct type
	})

	t.Run("select Categories", func(t *testing.T) {
		is := is.New(t)                 // is for the sub-test
		m := defaultview.InitialModel() // Use a fresh model
		m.Choices.Select(1)             // Select "Categories"
		_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		is.True(cmd != nil)
		msg := cmd()
		_, ok := msg.(defaultview.StartLoadingCategoriesMsg)
		is.True(ok)
	})
}

func TestDefaultUpdate_QuitKeys(t *testing.T) {
	// model := defaultview.InitialModel() // This outer model is not used if sub-tests create their own.
	quitKeys := []struct {
		name   string
		keyMsg tea.KeyPressMsg
		// keyStr string // keyStr was not used
	}{
		{"q", tea.KeyPressMsg{Code: 'q', Text: "q"}},
		{"esc", tea.KeyPressMsg{Code: tea.KeyEscape}},
		{"ctrl+c", tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}},
	}

	for _, tt := range quitKeys {
		t.Run(fmt.Sprintf("quit with %s", tt.name), func(t *testing.T) {
			is := is.New(t)
			// Use a fresh model for each sub-test to ensure isolation
			m := defaultview.InitialModel()
			_, cmd := m.Update(tt.keyMsg) // Use the model 'm' declared in the sub-test
			is.True(reflect.ValueOf(cmd).Pointer() == reflect.ValueOf(tea.Quit).Pointer())
		})
	}
}
