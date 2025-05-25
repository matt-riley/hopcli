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
	list         list.Model
	categoryName string
	categoryID   int
	width        int
	height       int
	products     *commands.Products // Store the fetched products
}

func NewModel(categoryName string, categoryID int) Model {
	return Model{
		list:         list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
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
		m.list.SetSize(msg.Width-docStyle.GetHorizontalFrameSize(), msg.Height-docStyle.GetVerticalFrameSize())
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
		m.list.SetItems(items)
		m.list.Title = fmt.Sprintf("Products in %s", m.categoryName)
		m.list.SetShowStatusBar(true)
		m.list.SetFilteringEnabled(false)
		m.list.SetSize(m.width-docStyle.GetHorizontalFrameSize(), m.height-docStyle.GetVerticalFrameSize())
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selectedItem, ok := m.list.SelectedItem().(CategoryProductListItem)
			if ok {
				// Trigger displaying the product details
				return m, commands.HandleDisplayProduct(m.width, m.height, selectedItem.productData)
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.products == nil { // Check if products are loaded yet
		return docStyle.Render(fmt.Sprintf("Loading products for %s...", m.categoryName))
	}
	if len(m.list.Items()) == 0 {
		return docStyle.Render(fmt.Sprintf("No products found in %s.", m.categoryName))
	}
	return docStyle.Render(m.list.View())
}
