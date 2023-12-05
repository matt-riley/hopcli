package hopt

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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

func (m Model) Init() tea.Cmd {
	return nil
}

func Run() {
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
