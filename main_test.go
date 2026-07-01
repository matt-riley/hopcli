package main

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/matryer/is"

	"github.com/matt-riley/hopcli/cmd/hopt"
	"github.com/matt-riley/hopcli/internal/api"
	"github.com/matt-riley/hopcli/internal/commands"
	defaultview "github.com/matt-riley/hopcli/internal/default"
)

// TestNewMainModel_Boundary exercises the model initialization that Run() invokes,
// verifying the MainModel is created with proper defaults before tea.NewProgram runs.
func TestNewMainModel_Boundary(t *testing.T) {
	is := is.New(t)

	model := hopt.InitialModel()
	is.True(model.CurrentView != nil)

	// Verify the default view is a DefaultModel
	_, ok := model.CurrentView.(defaultview.DefaultModel)
	is.True(ok)

	// Verify initial State
	is.True(model.State == hopt.DefaultView)

	// Verify no error on initialization
	is.Equal(model.ErrMsg, "")
	is.Equal(model.Loading, false)

	// Verify Spinner is initialized (always set to spinner.Dot in InitialModel)
	is.True(model.Spinner.Spinner.Frames != nil)
}

// TestRunBoundary_WithMockClient exercises the Run() boundary by
// verifying that Init() returns a valid command and that Update with
// Quit terminates cleanly — the two code paths that Run() exercises.
func TestRunBoundary_WithMockClient(t *testing.T) {
	is := is.New(t)

	// Ensure ApiClient is set so Init/Update behaviour is consistent.
	origClient := commands.ApiClient
	defer func() { commands.ApiClient = origClient }()

	commands.ApiClient = &mockClient{}

	model := hopt.InitialModel()

	// Init must return a non-nil command (spinner tick).
	initCmd := model.Init()
	is.True(initCmd != nil)

	// Sending 'q' must return a Quit command, terminating the program.
	nextModel, quitCmd := model.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	is.True(nextModel != nil)
	is.True(quitCmd != nil)
}

// mockClient implements api.Client for the root boundary test.
type mockClient struct{}

func (m *mockClient) FetchProducts(ctx context.Context, page, perPage int) (api.Products, api.Pagination, error) {
	return nil, api.Pagination{}, nil
}
func (m *mockClient) FetchCategories(ctx context.Context) (api.Categories, error) {
	return nil, nil
}
func (m *mockClient) FetchProduct(ctx context.Context, productID int) (api.Product, error) {
	return api.Product{}, nil
}
func (m *mockClient) FetchProductsByCategory(ctx context.Context, categoryID, page, perPage int) (api.Products, api.Pagination, error) {
	return nil, api.Pagination{}, nil
}
