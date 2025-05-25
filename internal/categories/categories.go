package categories

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/matt-riley/hopcli/internal/commands"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type CategoryListItem struct {
	ID          int
	Name        string
	Slug        string
	APIEndpoint string // Store the wp:post_type href
}

func (i CategoryListItem) Title() string       { return i.Name }
func (i CategoryListItem) Description() string { return "Slug: " + i.Slug }
func (i CategoryListItem) FilterValue() string { return i.Name }

type Model struct {
	list             list.Model
	width            int
	height           int
	selectedCategory CategoryListItem
}

func NewCategoriesModel() Model {
	return Model{
		list: list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		docStyle.Width(m.width)
		docStyle.Height(m.height)
		m.list.SetSize(msg.Width-docStyle.GetHorizontalFrameSize(), msg.Height-docStyle.GetVerticalFrameSize())
		return m, nil

	case commands.CategoriesResponseMsg:
		if msg.Err != nil {
			// Handle error, maybe return an error message to display
			// For now, just log or ignore
			return m, nil
		}
		items := []list.Item{}
		if msg.Categories != nil {
			for _, cat := range *msg.Categories {
				apiEndpoint := ""
				if len(cat.Links.WpPostType) > 0 {
					apiEndpoint = cat.Links.WpPostType[0].Href
				}
				items = append(items, CategoryListItem{
					ID:          cat.ID,
					Name:        cat.Name,
					Slug:        cat.Slug,
					APIEndpoint: apiEndpoint,
				})
			}
		}
		m.list.SetItems(items)
		m.list.Title = "Browse Categories"
		m.list.SetShowStatusBar(true)
		m.list.SetFilteringEnabled(false)                                                                   // Can enable if needed
		m.list.SetSize(m.width-docStyle.GetHorizontalFrameSize(), m.height-docStyle.GetVerticalFrameSize()) // Recalculate size based on current width/height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selectedItem, ok := m.list.SelectedItem().(CategoryListItem)
			if ok {
				m.selectedCategory = selectedItem
				// Trigger loading products for this category
				return m, func() tea.Msg {
					return commands.StartLoadingProductsForCategoryMsg{
						CategoryID:   selectedItem.ID,
						CategoryName: selectedItem.Name,
						APIEndpoint:  selectedItem.APIEndpoint,
						Width:        m.width,  // Pass current width
						Height:       m.height, // Pass current height
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.list.Items() == nil || len(m.list.Items()) == 0 {
		return docStyle.Render("Loading categories...")
	}
	return docStyle.Render(m.list.View())
}

// Helper function to get the selected category if needed by other parts of the app
func (m Model) SelectedCategory() CategoryListItem {
	return m.selectedCategory
}
