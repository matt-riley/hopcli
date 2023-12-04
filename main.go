package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LatestResponseMsg struct {
	Products Products
	Err      error
}
type Product struct {
	ID    int    `json:"id"`
	Link  string `json:"link"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Description struct {
		Rendered string `json:"rendered"`
	} `json:"excerpt"`
}

type Products []Product

type Model struct {
	choices  list.Model
	cursor   int
	selected map[int]struct{}
}

type ListItem struct {
	title string
	desc  string
	link  string
}

func (i ListItem) Title() string       { return i.title }
func (i ListItem) Description() string { return i.desc }
func (i ListItem) Link() string        { return i.link }
func (i ListItem) FilterValue() string { return i.title }

func initialModel() Model {
	items := []list.Item{
		ListItem{title: "Latest", desc: "Latest items added"},
	}

	return Model{
		choices:  list.New(items, list.NewDefaultDelegate(), 0, 0),
		selected: make(map[int]struct{}),
	}
}

var logo = `
⠀⠀⠀⠀⠀⠀⠀⢀⣰⣶⣶⣤⣤⣤⣤⣤⣀⠀⣀⣀⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⣀⣤⣤⣤⣬⣽⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣄⠀⠀⠀⠀⠀⠀
⠀⠐⠚⣻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢿⣿⣿⣿⣿⣿⣷⣄⣄⠀⠀⠀
⠐⣶⣾⣿⣿⣿⢿⣿⣿⣿⣿⣦⣍⠻⣿⣿⣿⣿⣦⡙⢿⣿⣿⣿⣿⣿⣿⣷⡄⠀
⠀⢿⣿⣿⣿⣿⣶⣍⠛⢿⣿⣿⣿⣷⡈⠻⣿⣿⣿⣿⡄⠹⣿⣿⣿⣿⣿⣿⣷⡀
⢠⣌⣿⣿⣿⡟⠁⠀⠀⠀⠈⠉⠛⠛⠛⠂⠈⠛⠋⠉⠁⠀⠀⠀⠙⣱⣿⣿⣿⡇
⠘⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⡿⠁
⠀⢹⣿⣿⡟⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢻⣿⣿⠇⠀
⠀⠀⢻⣿⣀⣠⠤⠤⠤⠤⠤⠤⣄⡀⠀⠀⠀⣀⡤⠤⠤⠤⠤⢤⣀⣈⣿⡟⠀⠀
⠠⡖⢸⡟⡟⠀⠀⠀⠀⠀⠀⠀⠀⢹⡶⢶⡏⠀⠀⠀⠀⠀⠀⠀⠈⣿⢻⡇⢾⡄
⠀⡇⠀⡇⠰⡀⠀⠀⠀⠀⠀⠀⢠⠟⠀⠈⢣⡀⠀⠀⠀⠀⠀⠀⢠⠏⢸⠁⢸⠃
⠀⠱⠀⡇⠀⠙⠢⠤⠤⠤⠤⠖⠋⠀⠀⠀⠀⠙⠢⠤⠤⠤⠤⠴⠋⠀⣼⠀⠃⠀
⠀⠀⠀⣿⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⡏⠀⠀⠀
⠀⠀⠀⣿⣧⠀⠀⠀⠀⢀⣠⣴⣶⡷⠶⠶⢶⣶⣦⣤⣀⠀⠀⠀⠀⣾⡇⠀⠀⠀
⠀⠀⠀⠘⣿⣦⠀⠀⣴⡟⠉⠉⠁⢀⣀⣀⡀⠀⠈⠉⠻⣧⡀⢀⣼⣿⡇⠀⠀⠀
⠀⠀⠀⠀⢸⣿⣷⣼⣿⣤⣤⣤⣴⣿⣿⣿⣿⣦⣤⣤⣤⣿⣷⣼⣿⣿⠃⠀⠀⠀
⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⣿⡿⢿⣿⣿⣿⣿⡿⢿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀
⠀⠀⠀⠀⢼⣿⣿⡿⠿⠛⣉⣴⡈⢿⣿⣿⡿⢡⣦⣙⠻⠿⢿⣿⣿⡇⠀⠀⠀⠀
⠀⠀⠀⠀⠀⣶⣶⣶⠀⣿⣿⣿⣿⣦⠙⢋⣴⣿⣿⣿⣿⠀⣶⣶⡦⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⢿⣿⡿⠀⠿⠿⠿⢛⣡⣶⣷⣌⠛⠿⠿⠿⠀⣿⣿⡟⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⢰⣶⣶⣶⣷⠘⣿⣿⣿⡿⢁⣿⣶⣶⣶⡆⠈⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠈⢿⣿⣿⡿⠓⣈⡛⢛⣁⠺⢿⣿⣿⡿⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠀⠐⢿⣿⣿⣿⣿⡗⠀⠈⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⠿⠟⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀`

func (m Model) View() string {
	output := lipgloss.JoinHorizontal(lipgloss.Top, logo, m.choices.View())
	return output
}

func (m Model) Init() tea.Cmd {
	return nil
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "h":

			items := []list.Item{
				ListItem{title: "Latest", desc: "Latest items added"},
			}
			m.choices.SetItems(items)
			return m, nil
		case "enter":
			if m.choices.Index() == 0 {
				return m, handleGetLatest()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.choices.SetSize(msg.Width-h, msg.Height-v)

	case LatestResponseMsg:
		if msg.Err != nil {
			panic(msg.Err.Error())
		}

		var items []list.Item
		for _, product := range msg.Products {
			title := html.UnescapeString(product.Title.Rendered)
			desc := strings.SplitAfter(html.UnescapeString(product.Description.Rendered), "%")
			items = append(items, ListItem{title: title, desc: desc[0][3:]})
		}
		m.choices.SetItems(items)
	}

	var cmd tea.Cmd
	m.choices, cmd = m.choices.Update(msg)
	return m, cmd
}

func main() {
	model := initialModel()
	model.choices.SetShowTitle(true)
	model.choices.Title = "The Hoptimist"
	model.choices.SetFilteringEnabled(false)
	model.choices.SetShowHelp(false)
	model.choices.SetShowStatusBar(false)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

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
