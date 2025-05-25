package defaultview

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/matt-riley/hopcli/internal/commands"
)

const (
	latest = iota
	categories
)

// StartLoadingLatestMsg is a message to indicate that the latest items should be loaded.
type StartLoadingLatestMsg struct{}

// StartLoadingCategoriesMsg is a message to indicate that the categories should be loaded.
type StartLoadingCategoriesMsg struct{}

type DefaultModel struct {
	choices list.Model
}

type ListItem struct {
	title string
	desc  string
}

func (i ListItem) Title() string       { return i.title }
func (i ListItem) Description() string { return i.desc }
func (i ListItem) FilterValue() string { return i.title }

func InitialModel() DefaultModel {
	items := []list.Item{
		ListItem{title: "Latest", desc: "Latest items added"},
		ListItem{title: "Categories", desc: "Browse by category"},
	}

	return DefaultModel{
		choices: list.New(items, list.NewDefaultDelegate(), 0, 0),
	}
}

func (dm DefaultModel) Init() tea.Cmd {
	return nil
}

var docStyle = lipgloss.NewStyle().Margin(2)

func (dm DefaultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return dm, tea.Quit
		case "enter":
			switch dm.choices.Index() {
			case latest:
				return dm, func() tea.Msg { return StartLoadingLatestMsg{} }
			case categories:
				return dm, func() tea.Msg { return StartLoadingCategoriesMsg{} }
			}
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		dm.choices.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	dm.choices, cmd = dm.choices.Update(msg)
	return dm, cmd
}

func (dm DefaultModel) View() string {
	dm.choices.Title = "The Hoptimist"
	dm.choices.SetFilteringEnabled(false)
	dm.choices.SetShowStatusBar(false)
	return dm.choices.View()
}
