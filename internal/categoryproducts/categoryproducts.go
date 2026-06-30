package categoryproducts

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/matt-riley/hopcli/internal/commands"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type Model struct {
	commands.PaginatedModel
	List         list.Model // Exported
	categoryName string
	categoryID   int
	width        int
	height       int
	products     *commands.Products // Store the fetched products
	ErrMsg       string             // error message to display in View()
}

func NewModel(categoryName string, categoryID int) Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.SetShowPagination(true)
	l.SetShowStatusBar(true)
	return Model{
		PaginatedModel: commands.PaginatedModel{
			CurrentPage: 1,
			PerPage:     10,
		},
		List:         l,
		categoryName: categoryName,
		categoryID:   categoryID,
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
		m.List.SetSize(msg.Width-docStyle.GetHorizontalFrameSize(), msg.Height-docStyle.GetVerticalFrameSize())
		return m, nil

	case commands.ProductsForCategoryResponseMsg:
		if err := commands.ResponseError(msg); err != nil {
			m.ErrMsg = err.Error()
			return m, nil
		}
		if msg.CategoryID != m.categoryID { // Ensure this message is for the current category
			return m, nil
		}

		m.products = msg.Products // Store products
		m.TotalItems = msg.TotalItems
		m.TotalPages = msg.TotalPages

		// Set the list title here (pure View() — no side effects).
		switch {
		case m.TotalItems == 0:
			m.List.Title = fmt.Sprintf("Products in %s (No items found)", m.categoryName)
		case m.TotalPages == 0:
			// Edge case: API says 0 pages but has items — treat as page 1.
			m.List.Title = fmt.Sprintf("Products in %s (Page %d/%d)", m.categoryName, m.CurrentPage, 1)
		default:
			m.List.Title = fmt.Sprintf("Products in %s (Page %d/%d)", m.categoryName, m.CurrentPage, m.TotalPages)
		}

		items := []list.Item{}
		if m.products != nil {
			for _, prod := range *m.products {
				items = append(items, commands.NewProductListItem(prod))
			}
		}
		m.List.SetItems(items)
		m.List.SetFilteringEnabled(false)
		m.List.SetSize(m.width-docStyle.GetHorizontalFrameSize(), m.height-docStyle.GetVerticalFrameSize())

		// Update list paginator
		m.List.Paginator.PerPage = m.PerPage
		m.List.Paginator.Page = m.CurrentPage - 1 // list.Paginator is 0-indexed
		m.List.Paginator.TotalPages = m.TotalPages

		return m, nil

	case tea.KeyMsg:
		// Check for 'n'/'p' navigation first
		if pageChanged, newPage := m.UpdatePageNavigation(msg); pageChanged {
			return m, func() tea.Msg {
				return commands.LoadCategoryProductsPageMsg{
					CategoryID:   m.categoryID,
					CategoryName: m.categoryName,
					Page:         newPage,
					PerPage:      m.PerPage,
				}
			}
		}
		if msg.String() == "enter" {
			if len(m.List.VisibleItems()) > 0 && m.List.Index() < len(*m.products) {
				selectedItem, ok := m.List.SelectedItem().(commands.ProductListItem)
				if ok {
					// Trigger displaying the product details
					return m, commands.HandleDisplayProduct(m.width, m.height, selectedItem.ProductData())
				}
			}
		}
	}

	var listCmd tea.Cmd
	m.List, listCmd = m.List.Update(msg)
	cmds = append(cmds, listCmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() tea.View {
	if m.ErrMsg != "" {
		return tea.NewView(docStyle.Render("Error: " + m.ErrMsg))
	}
	if m.products == nil { // Check if products are loaded yet
		return tea.NewView(docStyle.Render(fmt.Sprintf("Loading products for %s...", m.categoryName)))
	}
	return tea.NewView(docStyle.Render(m.List.View()))
}

// CategoryID returns the category ID of the model.
func (m Model) CategoryID() int {
	return m.categoryID
}

// CategoryName returns the category name of the model.
func (m Model) CategoryName() string {
	return m.categoryName
}
