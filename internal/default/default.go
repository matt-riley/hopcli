package defaultview

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	Choices list.Model // Exported
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
		Choices: list.New(items, list.NewDefaultDelegate(), 0, 0), // Use exported field
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
			switch dm.Choices.Index() { // Use exported field
			case latest:
				return dm, func() tea.Msg { return StartLoadingLatestMsg{} }
			case categories:
				return dm, func() tea.Msg { return StartLoadingCategoriesMsg{} }
			}
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		dm.Choices.SetSize(msg.Width-h, msg.Height-v) // Use exported field
	}

	var cmd tea.Cmd
	dm.Choices, cmd = dm.Choices.Update(msg) // Use exported field
	return dm, cmd
}

func (dm DefaultModel) View() tea.View {
	dm.Choices.Title = "The Hoptimist"    // Use exported field
	dm.Choices.SetFilteringEnabled(false) // Use exported field
	dm.Choices.SetShowStatusBar(false)    // Use exported field
	return tea.NewView(dm.Choices.View()) // Use exported field
}
