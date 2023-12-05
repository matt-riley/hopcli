package hopt

import (
	"fmt"
	"html"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
			fmt.Printf("Error: %s", msg.Err.Error())
			os.Exit(1)
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
