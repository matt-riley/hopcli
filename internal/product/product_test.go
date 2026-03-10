package productview_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/internal/commands"
	productview "github.com/matt-riley/hopcli/internal/product"
)

var (
	oscSequenceRe = regexp.MustCompile(`\x1b\][^\x07]*\x07`)
	csiSequenceRe = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)
)

func stripTerminalSequences(s string) string {
	s = oscSequenceRe.ReplaceAllString(s, "")
	return csiSequenceRe.ReplaceAllString(s, "")
}

func TestProductNewProductModel(t *testing.T) {
	is := is.New(t)
	model := productview.NewProductModel()

	is.Equal(model.Product.Title, "")       // Product should be zero-value initially
	is.Equal(model.Product.Description, "") // Product description should be zero-value
	is.Equal(model.Product.URL, "")         // Product URL should be zero-value
	is.Equal(model.Product.Price, "")       // Product price should be zero-value
	is.Equal(model.Width, 0)                // Width should be zero
	is.Equal(model.Height, 0)               // Height should be zero
}

func TestProductUpdate_ProductsMsg_Success(t *testing.T) {
	is := is.New(t)
	model := productview.NewProductModel()

	sampleCmdProduct := &commands.Product{
		ID:          1,
		Link:        "http://store.com/product/1",
		Title:       "Test Product",
		Description: "<p>Hello <strong>World</strong></p>",
		Prices: commands.ProductPrices{
			Price:             "420",
			CurrencyPrefix:    "£",
			CurrencySuffix:    "",
			CurrencyMinorUnit: 2,
		},
	}

	msg := commands.ProductsMsg{Product: sampleCmdProduct, Width: 80, Height: 24, Err: nil}
	updatedModelTea, _ := model.Update(msg)
	updatedModel := updatedModelTea.(productview.ProductModel)

	is.Equal(updatedModel.Product.Title, "Test Product")
	is.True(strings.Contains(updatedModel.Product.Description, "Hello"))
	is.True(strings.Contains(updatedModel.Product.Description, "World"))
	is.Equal(updatedModel.Product.URL, "http://store.com/product/1")
	is.Equal(updatedModel.Product.Price, "£4.20")
	is.Equal(updatedModel.Width, 80)
}

// Panic regression test: description with no '%' must not panic.
func TestProductUpdate_ProductsMsg_NoPercentInDescription(t *testing.T) {
	is := is.New(t)
	model := productview.NewProductModel()

	sampleCmdProduct := &commands.Product{
		ID:          2,
		Link:        "http://store.com/product/2",
		Title:       "No Percent Product",
		Description: "<p>A description with no percent sign at all.</p>",
	}

	msg := commands.ProductsMsg{Product: sampleCmdProduct, Width: 80, Err: nil}
	updatedModelTea, _ := model.Update(msg)
	updatedModel := updatedModelTea.(productview.ProductModel)

	is.Equal(updatedModel.Product.Title, "No Percent Product")
	is.True(strings.Contains(updatedModel.Product.Description, "no percent sign"))
}

func TestProductUpdate_ProductsMsg_Error(t *testing.T) {
	is := is.New(t)
	model := productview.NewProductModel()
	originalProduct := model.Product

	msg := commands.ProductsMsg{Product: nil, Width: 80, Height: 24, Err: fmt.Errorf("API error")}
	updatedModelTea, _ := model.Update(msg)
	updatedModel := updatedModelTea.(productview.ProductModel)

	is.Equal(updatedModel.Product.Title, originalProduct.Title)
	is.Equal(updatedModel.Product.Description, originalProduct.Description)
	is.Equal(updatedModel.Product.URL, originalProduct.URL)
}

func TestProductView_Rendering_WithPrice(t *testing.T) {
	is := is.New(t)
	model := productview.NewProductModel()
	model.Product = productview.Product{
		Title:       "My Product",
		Description: "Some description.",
		URL:         "http://example.com",
		Price:       "£4.20",
	}
	model.Width = 80

	view := model.View()
	plainContent := stripTerminalSequences(view.Content)

	is.True(view.Content != "")
	is.True(strings.Contains(plainContent, "My Product"))
	is.True(strings.Contains(plainContent, "£4.20"))
	is.True(strings.Contains(plainContent, "Some description."))
	is.True(strings.Contains(plainContent, "View product"))
	is.True(strings.Contains(plainContent, "example.com"))
}

func TestProductView_Rendering_WithoutPrice(t *testing.T) {
	is := is.New(t)
	model := productview.NewProductModel()
	model.Product = productview.Product{
		Title:       "Free Product",
		Description: "No price here.",
		URL:         "http://example.com/free",
		Price:       "",
	}
	model.Width = 80

	view := model.View()
	plainContent := stripTerminalSequences(view.Content)

	is.True(view.Content != "")
	is.True(strings.Contains(plainContent, "Free Product"))
	is.True(!strings.Contains(plainContent, "Price:"))
}
