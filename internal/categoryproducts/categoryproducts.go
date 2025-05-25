package categoryproducts

import (
	"fmt"
	"html"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/matt-riley/hopcli/internal/commands"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

// CategoryProductListItem defines a list item for products within a category.
type CategoryProductListItem struct {
	title       string
	description string
	productData commands.Product
}

func (i CategoryProductListItem) Title() string       { return i.title }
func (i CategoryProductListItem) Description() string { return i.description }
func (i CategoryProductListItem) FilterValue() string { return i.title }

type Model struct {
	List         list.Model // Exported
	categoryName string
	categoryID   int
	width        int
	height       int
	products     *commands.Products // Store the fetched products
	CurrentPage  int                // New
	PerPage      int                // New
	TotalItems   int                // New
	TotalPages   int                // New
	apiEndpoint  string             // New: Store the API endpoint for this category
}

func NewModel(categoryName string, categoryID int, apiEndpoint string) Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.SetShowPagination(true)
	l.SetShowStatusBar(true)
	return Model{
		List:         l, // Use exported field
		categoryName: categoryName,
		categoryID:   categoryID,
		apiEndpoint:  apiEndpoint, // Store the endpoint
		CurrentPage:  1,
		PerPage:      10, // Default items per page
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
		m.List.SetSize(msg.Width-docStyle.GetHorizontalFrameSize(), msg.Height-docStyle.GetVerticalFrameSize()) // Use exported field
		return m, nil

	case commands.ProductsForCategoryResponseMsg:
		if msg.Err != nil {
			// Handle error - perhaps set a message on the model to display in View()
			return m, nil
		}
		if msg.CategoryID != m.categoryID { // Ensure this message is for the current category
			return m, nil
		}

		m.products = msg.Products // Store products
		m.TotalItems = msg.TotalItems
		m.TotalPages = msg.TotalPages
		// m.CurrentPage is implicitly correct as this response is for the current page request

		items := []list.Item{}
		if m.products != nil {
			for _, prod := range *m.products {
				unescapedTitle := html.UnescapeString(prod.Title.Rendered)
				// Create a short description, ensuring HTML is unescaped first
				fullDesc := html.UnescapeString(prod.Description.Rendered)
				var shortDesc string
				if len(fullDesc) > 0 {
					// Attempt to split by sentences, take the first.
					// This is a simple approach; more robust HTML to plain text conversion might be needed.
					sentences := strings.Split(fullDesc, ".")
					shortDesc = sentences[0]
					if len(sentences) > 1 { // Add ellipsis if there was more than one sentence
						shortDesc += "."
					}
				}

				if len(shortDesc) > 100 { // Arbitrary length limit for display
					shortDesc = shortDesc[:100] + "..."
				}

				items = append(items, CategoryProductListItem{
					title:       unescapedTitle,
					description: shortDesc,
					productData: prod,
				})
			}
		}
		m.List.SetItems(items) // Use exported field
		// Title is set in View()
		// m.List.SetShowStatusBar(true) // Already set in NewModel
		m.List.SetFilteringEnabled(false)                                                                   // Use exported field
		m.List.SetSize(m.width-docStyle.GetHorizontalFrameSize(), m.height-docStyle.GetVerticalFrameSize()) // Use exported field

		// Update list paginator
		m.List.Paginator.PerPage = m.PerPage      // Use exported field
		m.List.Paginator.Page = m.CurrentPage - 1 // list.Paginator is 0-indexed
		m.List.Paginator.TotalPages = m.TotalPages
		// No explicit command needed from this message handling itself for list updates.

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if len(m.List.VisibleItems()) > 0 && m.List.Index() < len(*m.products) { // Use exported field
				selectedItem, ok := m.List.SelectedItem().(CategoryProductListItem) // Use exported field
				if ok {
					// Trigger displaying the product details
					return m, commands.HandleDisplayProduct(m.width, m.height, selectedItem.productData)
				}
			}
		case "n": // Next page
			if m.CurrentPage < m.TotalPages {
				m.CurrentPage++
				return m, func() tea.Msg {
					return commands.LoadCategoryProductsPageMsg{
						CategoryID:   m.categoryID,
						CategoryName: m.categoryName,
						APIEndpoint:  m.apiEndpoint, // Use stored apiEndpoint
						Page:         m.CurrentPage,
						PerPage:      m.PerPage,
					}
				}
			}
		case "p": // Previous page
			if m.CurrentPage > 1 {
				m.CurrentPage--
				return m, func() tea.Msg {
					return commands.LoadCategoryProductsPageMsg{
						CategoryID:   m.categoryID,
						CategoryName: m.categoryName,
						APIEndpoint:  m.apiEndpoint, // Use stored apiEndpoint
						Page:         m.CurrentPage,
						PerPage:      m.PerPage,
					}
				}
			}
		}
	}

	var listCmd tea.Cmd
	m.List, listCmd = m.List.Update(msg) // Use exported field
	cmds = append(cmds, listCmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.products == nil { // Check if products are loaded yet
		return docStyle.Render(fmt.Sprintf("Loading products for %s...", m.categoryName))
	}
	if len(m.List.Items()) == 0 && m.TotalItems == 0 { // Check if there are genuinely no items ; Use exported field
		m.List.Title = fmt.Sprintf("Products in %s (No items found)", m.categoryName) // Use exported field
	} else {
		m.List.Title = fmt.Sprintf("Products in %s (Page %d/%d)", m.categoryName, m.CurrentPage, m.TotalPages) // Use exported field
		if m.TotalPages == 0 && m.TotalItems > 0 {                                                             // Edge case: API says 0 pages but has items
			m.List.Title = fmt.Sprintf("Products in %s (Page %d/%d)", m.categoryName, m.CurrentPage, 1) // Use exported field
		}
	}
	return docStyle.Render(m.List.View()) // Use exported field
}

// CategoryID returns the category ID of the model.
func (m Model) CategoryID() int {
	return m.categoryID
}

// CategoryName returns the category name of the model.
func (m Model) CategoryName() string {
	return m.categoryName
}
