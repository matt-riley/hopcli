package categories_test

import (
	"testing"

	// "charm.land/bubbles/v2/list" // Not directly used for assertions
	tea "charm.land/bubbletea/v2"
	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/internal/categories"
	"github.com/matt-riley/hopcli/internal/commands"
)

func TestCategoriesNewCategoriesModel(t *testing.T) {
	is := is.New(t)
	model := categories.NewCategoriesModel()

	is.True(model.List.Items() == nil || len(model.List.Items()) == 0) // Initially empty
	is.Equal(model.List.Paginator.Page, 0)                             // Default paginator page
	// Note: model.width and model.height are unexported and will be zero by default.
}

func TestCategoriesUpdate_CategoriesResponseMsg(t *testing.T) {
	is := is.New(t)
	model := categories.NewCategoriesModel()

	sampleCategories := &commands.Categories{
		{ID: 1, Name: "IPA", Slug: "ipa"},
		{ID: 2, Name: "Stout", Slug: "stout"},
	}
	msg := commands.CategoriesResponseMsg{Categories: sampleCategories, Width: 80, Height: 24}
	updatedModelTea, _ := model.Update(msg)
	updatedModel := updatedModelTea.(categories.Model)

	is.Equal(len(updatedModel.List.Items()), 2)

	firstItem, ok := updatedModel.List.Items()[0].(categories.CategoryListItem)
	is.True(ok)
	is.Equal(firstItem.ID, 1)
	is.Equal(firstItem.Name, "IPA")
	is.Equal(firstItem.Slug, "ipa")

	secondItem, ok := updatedModel.List.Items()[1].(categories.CategoryListItem)
	is.True(ok)
	is.Equal(secondItem.ID, 2)
	is.Equal(secondItem.Name, "Stout")
	is.Equal(secondItem.Slug, "stout")
}

func TestCategoriesUpdate_EnterKey(t *testing.T) {
	is := is.New(t)
	model := categories.NewCategoriesModel()

	// Manually set width and height as if a WindowSizeMsg was received
	// These are unexported, so this step simulates their internal state for the test.
	// We can't directly set them. The test will rely on their zero values if WindowSizeMsg is not sent.
	// The `StartLoadingProductsForCategoryMsg` will use these internal width/height values.
	// For robustness, we can send a WindowSizeMsg first.
	modelTea, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = modelTea.(categories.Model)

	sampleCategories := &commands.Categories{
		{ID: 1, Name: "IPA", Slug: "ipa"},
		{ID: 2, Name: "Stout", Slug: "stout"},
	}

	modelTea, _ = model.Update(commands.CategoriesResponseMsg{Categories: sampleCategories, Width: 80, Height: 24})
	model = modelTea.(categories.Model)

	model.List.Select(0) // Select the first item ("IPA")

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	is.True(cmd != nil)

	msgResult := cmd().(commands.StartLoadingProductsForCategoryMsg)
	is.Equal(msgResult.CategoryID, 1)
	is.Equal(msgResult.CategoryName, "IPA")
	is.Equal(msgResult.Width, 100) // Check if width from WindowSizeMsg is passed
	is.Equal(msgResult.Height, 50) // Check if height from WindowSizeMsg is passed
}
