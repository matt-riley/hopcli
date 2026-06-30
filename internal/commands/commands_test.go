package commands_test

import (
	"context"
	"errors"
	"testing"

	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/internal/api"
	"github.com/matt-riley/hopcli/internal/commands"
)

// mockClient implements api.Client for testing.
type mockClient struct {
	productsFn           func(ctx context.Context, page, perPage int) (api.Products, api.Pagination, error)
	categoriesFn         func(ctx context.Context) (api.Categories, error)
	productFn            func(ctx context.Context, productID int) (api.Product, error)
	productsByCategoryFn func(ctx context.Context, categoryID, page, perPage int) (api.Products, api.Pagination, error)
}

func (m *mockClient) FetchProducts(ctx context.Context, page, perPage int) (api.Products, api.Pagination, error) {
	return m.productsFn(ctx, page, perPage)
}

func (m *mockClient) FetchCategories(ctx context.Context) (api.Categories, error) {
	return m.categoriesFn(ctx)
}

func (m *mockClient) FetchProduct(ctx context.Context, productID int) (api.Product, error) {
	return m.productFn(ctx, productID)
}

func (m *mockClient) FetchProductsByCategory(ctx context.Context, categoryID, page, perPage int) (api.Products, api.Pagination, error) {
	return m.productsByCategoryFn(ctx, categoryID, page, perPage)
}

func TestHandleGetLatest_Pagination(t *testing.T) {
	is := is.New(t)

	origClient := commands.ApiClient
	defer func() { commands.ApiClient = origClient }()

	commands.ApiClient = &mockClient{
		productsFn: func(ctx context.Context, page, perPage int) (api.Products, api.Pagination, error) {
			is.Equal(page, 2)
			is.Equal(perPage, 5)
			return api.Products{
				{ID: 1, Title: "Test Beer 1", Description: "Desc 1", ShortDescription: "Short 1",
					Link: "https://example.com/test-beer-1",
					Prices: api.ProductPrices{
						Price: "420", RegularPrice: "420", CurrencyCode: "GBP",
						CurrencySymbol: "£", CurrencyMinorUnit: 2, CurrencyPrefix: "£",
					}},
			}, api.Pagination{TotalItems: 50, TotalPages: 10}, nil
		},
	}

	cmd := commands.HandleGetLatest(context.Background(), 80, 24, 2, 5, 1)
	msg := cmd()

	is.True(msg != nil)
	latestMsg, ok := msg.(commands.LatestResponseMsg)
	is.True(ok)

	is.True(latestMsg.Err == nil)
	is.Equal(latestMsg.RequestID, 1)
	is.Equal(latestMsg.TotalItems, 50)
	is.Equal(latestMsg.TotalPages, 10)
	is.True(latestMsg.Products != nil)
	is.Equal(len(*latestMsg.Products), 1)
	if len(*latestMsg.Products) == 1 {
		is.Equal((*latestMsg.Products)[0].ID, 1)
		is.Equal((*latestMsg.Products)[0].Title, "Test Beer 1")
	}
}

func TestHandleGetProductsByCategory_Pagination(t *testing.T) {
	is := is.New(t)

	origClient := commands.ApiClient
	defer func() { commands.ApiClient = origClient }()

	expectedCatID := 123
	expectedTotalItems := 24
	expectedTotalPages := 3

	commands.ApiClient = &mockClient{
		productsByCategoryFn: func(ctx context.Context, categoryID, page, perPage int) (api.Products, api.Pagination, error) {
			is.Equal(categoryID, expectedCatID)
			is.Equal(page, 3)
			is.Equal(perPage, 8)
			return api.Products{
				{ID: 2, Title: "Category Beer 2", Description: "Desc 2", ShortDescription: "Short 2",
					Link: "https://example.com/category-beer-2",
					Prices: api.ProductPrices{
						Price: "599", RegularPrice: "599", CurrencyCode: "GBP",
						CurrencySymbol: "£", CurrencyMinorUnit: 2, CurrencyPrefix: "£",
					}},
			}, api.Pagination{TotalItems: expectedTotalItems, TotalPages: expectedTotalPages}, nil
		},
	}

	cmd := commands.HandleGetProductsByCategory(context.Background(), expectedCatID, "Test Cat", 3, 8, 1)
	msg := cmd()

	is.True(msg != nil)
	catMsg, ok := msg.(commands.ProductsForCategoryResponseMsg)
	is.True(ok)

	is.True(catMsg.Err == nil)
	is.Equal(catMsg.RequestID, 1)
	is.Equal(catMsg.TotalItems, expectedTotalItems)
	is.Equal(catMsg.TotalPages, expectedTotalPages)
	is.Equal(catMsg.CategoryID, expectedCatID)
	is.Equal(catMsg.CategoryName, "Test Cat")
	is.True(catMsg.Products != nil)
	is.Equal(len(*catMsg.Products), 1)
	if len(*catMsg.Products) == 1 {
		is.Equal((*catMsg.Products)[0].ID, 2)
		is.Equal((*catMsg.Products)[0].Title, "Category Beer 2")
	}
}

func TestHandleGetCategories_HappyPath(t *testing.T) {
	is := is.New(t)

	origClient := commands.ApiClient
	defer func() { commands.ApiClient = origClient }()

	commands.ApiClient = &mockClient{
		categoriesFn: func(ctx context.Context) (api.Categories, error) {
			return api.Categories{
				{ID: 1, Name: "Category 1", Slug: "category-1", Parent: 0, Count: 5},
			}, nil
		},
	}

	cmd := commands.HandleGetCategories(context.Background(), 1)
	msg := cmd()

	is.True(msg != nil)
	catMsg, ok := msg.(commands.CategoriesResponseMsg)
	is.True(ok)
	is.True(catMsg.Err == nil)
	is.Equal(catMsg.RequestID, 1)
	is.True(catMsg.Categories != nil && len(*catMsg.Categories) == 1)
	if len(*catMsg.Categories) == 1 {
		is.Equal((*catMsg.Categories)[0].Name, "Category 1")
	}
}

func TestHandleGetLatest_HTTPErrorStatus(t *testing.T) {
	is := is.New(t)

	origClient := commands.ApiClient
	defer func() { commands.ApiClient = origClient }()

	commands.ApiClient = &mockClient{
		productsFn: func(ctx context.Context, page, perPage int) (api.Products, api.Pagination, error) {
			return nil, api.Pagination{}, errors.New("HTTP 404: not found")
		},
	}

	cmd := commands.HandleGetLatest(context.Background(), 80, 24, 1, 10, 1)
	msg := cmd()

	latestMsg, ok := msg.(commands.LatestResponseMsg)
	is.True(ok)
	is.True(latestMsg.Err != nil)
}

func TestHandleGetLatest_ErrorResponseCarriesRequestID(t *testing.T) {
	is := is.New(t)

	origClient := commands.ApiClient
	defer func() { commands.ApiClient = origClient }()

	commands.ApiClient = &mockClient{
		productsFn: func(ctx context.Context, page, perPage int) (api.Products, api.Pagination, error) {
			return nil, api.Pagination{}, errors.New("server error")
		},
	}

	cmd := commands.HandleGetLatest(context.Background(), 80, 24, 1, 10, 42)
	msg := cmd()

	latestMsg, ok := msg.(commands.LatestResponseMsg)
	is.True(ok)
	is.True(latestMsg.Err != nil)
	is.Equal(latestMsg.RequestID, 42)
}

func TestFormatPrice(t *testing.T) {
	is := is.New(t)

	is.Equal(commands.FormatPrice("420", "£", "", 2), "£4.20")
	is.Equal(commands.FormatPrice("0", "£", "", 2), "£0.00")
	is.Equal(commands.FormatPrice("100", "¥", "", 0), "¥100")
	is.Equal(commands.FormatPrice("", "£", "", 2), "")
	is.Equal(commands.FormatPrice("free", "£", "", 2), "£free")
	is.Equal(commands.FormatPrice("420", "£", "", -1), "£420")
}

func TestExtractSummary(t *testing.T) {
	is := is.New(t)

	is.Equal(
		commands.ExtractSummary(`<p>Lager &#8211; Helles &#8211; Bottle 500ml &#8211; 5.0%</p><p>Full description here.</p>`),
		"Lager – Helles – Bottle 500ml – 5.0%",
	)

	is.Equal(
		commands.ExtractSummary(`<p>IPA &#8211; New England / Hazy &#8211; Can 440ml &#8211; 6.5%</p>`),
		"IPA – New England / Hazy – Can 440ml – 6.5%",
	)

	is.Equal(
		commands.ExtractSummary(`<p>Pale Ale &#8211; Can 440ml &#8211; 5.4%</p>`),
		"Pale Ale – Can 440ml – 5.4%",
	)

	is.Equal(
		commands.ExtractSummary(`<p class="summary">Stout &#8211; Imperial &#8211; Can 330ml &#8211; 10.0%</p>`),
		"Stout – Imperial – Can 330ml – 10.0%",
	)

	is.Equal(
		commands.ExtractSummary(`<p><strong>Wheat Beer</strong> &#8211; Hefeweizen &#8211; Bottle 500ml &#8211; 5.3%</p>`),
		"Wheat Beer – Hefeweizen – Bottle 500ml – 5.3%",
	)

	is.Equal(commands.ExtractSummary(`Just plain text, no tags`), "")
	is.Equal(commands.ExtractSummary(""), "")
	is.Equal(
		commands.ExtractSummary(`<p>  Lager &#8211; Helles &#8211; Bottle 500ml &#8211; 5.0%  </p>`),
		"Lager – Helles – Bottle 500ml – 5.0%",
	)
}

func TestMockClientImplementsInterface(t *testing.T) {
	var _ api.Client = (*mockClient)(nil)
}
