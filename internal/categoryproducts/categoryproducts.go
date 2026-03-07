package categoryproducts

import (
	"fmt"
	"html"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

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
	products    *commands.Products // Store the fetched products
	CurrentPage int                // New
	PerPage     int                // New
	TotalItems  int                // New
	TotalPages  int                // New
}

func NewModel(categoryName string, categoryID int) Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.SetShowPagination(true)
	l.SetShowStatusBar(true)
	return Model{
		List:         l,
		categoryName: categoryName,
		categoryID:   categoryID,
		CurrentPage:  1,
		PerPage:      10,
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

		items := []list.Item{}
		if m.products != nil {
			for _, prod := range *m.products {
				formattedPrice := commands.FormatPrice(
					prod.Prices.Price,
					prod.Prices.CurrencyPrefix,
					prod.Prices.CurrencySuffix,
					prod.Prices.CurrencyMinorUnit,
				)
				onSaleMarker := ""
				if prod.OnSale {
					onSaleMarker = " 🏷️"
				}

				shortDesc := html.UnescapeString(prod.ShortDescription)
				if shortDesc == "" {
					shortDesc = html.UnescapeString(prod.Description)
				}

				var desc string
				if formattedPrice != "" {
					desc = fmt.Sprintf("%s%s | %s", formattedPrice, onSaleMarker, shortDesc)
				} else {
					desc = shortDesc
				}

				items = append(items, CategoryProductListItem{
					title:       html.UnescapeString(prod.Title),
					description: desc,
					productData: prod,
				})
			}
		}
		m.List.SetItems(items)                                                                              // Use exported field
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

func (m Model) View() tea.View {
	if m.products == nil { // Check if products are loaded yet
		return tea.NewView(docStyle.Render(fmt.Sprintf("Loading products for %s...", m.categoryName)))
	}
	if len(m.List.Items()) == 0 && m.TotalItems == 0 { // Check if there are genuinely no items ; Use exported field
		m.List.Title = fmt.Sprintf("Products in %s (No items found)", m.categoryName) // Use exported field
	} else {
		m.List.Title = fmt.Sprintf("Products in %s (Page %d/%d)", m.categoryName, m.CurrentPage, m.TotalPages) // Use exported field
		if m.TotalPages == 0 && m.TotalItems > 0 {                                                             // Edge case: API says 0 pages but has items
			m.List.Title = fmt.Sprintf("Products in %s (Page %d/%d)", m.categoryName, m.CurrentPage, 1) // Use exported field
		}
	}
	return tea.NewView(docStyle.Render(m.List.View())) // Use exported field
}

// CategoryID returns the category ID of the model.
func (m Model) CategoryID() int {
	return m.categoryID
}

// CategoryName returns the category name of the model.
func (m Model) CategoryName() string {
	return m.categoryName
}
