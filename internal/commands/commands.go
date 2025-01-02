package commands

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	Product struct {
		ID    int    `json:"id"`
		Link  string `json:"link"`
		Title struct {
			Rendered string `json:"rendered"`
		} `json:"title"`
		Description struct {
			Rendered string `json:"rendered"`
		} `json:"excerpt"`
	}
	Products          []Product
	LatestResponseMsg struct {
		Products *Products
		Width    int
		Height   int
		Err      error
	}
	ProductsMsg struct {
		Product *Product
		Width   int
		Height  int
		Err     error
	}
)

func HandleGetLatest(w int, h int) tea.Cmd {
	return func() tea.Msg {
		url := "https://thehoptimist.co.uk/wp-json/wp/v2/product?order_by=date"
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return LatestResponseMsg{
				Err: err,
			}
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return LatestResponseMsg{
				Err: err,
			}
		}
		defer res.Body.Close()

		var products Products
		err = json.NewDecoder(res.Body).Decode(&products)
		if err != nil {
			return LatestResponseMsg{
				Err: err,
			}
		}

		return LatestResponseMsg{
			Products: &products,
			Width:    w,
			Height:   h,
		}
	}
}

func HandleDisplayProduct(w int, h int, prod Product) tea.Cmd {
	return func() tea.Msg {
		return ProductsMsg{
			Product: &prod,
			Width:   w,
			Height:  h,
		}
	}
}
