package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type Model struct {
	choices  list.Model
	cursor   int
	selected map[int]struct{}
}

type ListItem struct {
	title string
	desc  string
}

func (i ListItem) Title() string       { return i.title }
func (i ListItem) Description() string { return i.desc }
func (i ListItem) FilterValue() string { return i.title }

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

func initialModel() Model {
	items := []list.Item{
		ListItem{title: "Latest", desc: "Latest items added"},
		ListItem{title: "Categories", desc: "Categories of items"},
	}

	return Model{
		choices:  list.New(items, list.NewDefaultDelegate(), 0, 0),
		selected: make(map[int]struct{}),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.choices.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.choices, cmd = m.choices.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	output := lipgloss.JoinHorizontal(lipgloss.Top, logo, m.choices.View())
	return output
}

func main() {
	model := initialModel()
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
