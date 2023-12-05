package hopt

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func handleGetLatest() tea.Cmd {
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
			Products: products,
		}
	}
}
