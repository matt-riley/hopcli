package latest

import (
	"fmt"
	"html"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/matt-riley/hopcli/internal/commands"
)

type LatestModel struct {
	products    *commands.Products
	Choices     list.Model // Exported
	CurrentPage int        // New
	PerPage     int        // New
	TotalItems  int        // New
	TotalPages  int        // New
	// width, height if needed for page change messages, or get from MainModel
}
type LatestListItem struct {
	title   string
	desc    string
	brewery string
}

func (i LatestListItem) Title() string       { return i.title }
func (i LatestListItem) Description() string { return i.desc }
func (i LatestListItem) Brewery() string     { return i.brewery }
func (i LatestListItem) FilterValue() string { return i.title }

func NewLatestModel() LatestModel {
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.SetShowPagination(true) // Show list's pagination
	l.SetShowStatusBar(true)  // Ensure status bar is shown for pagination info

	return LatestModel{
		Choices:     l, // Use exported field
		CurrentPage: 1,
		PerPage:     10, // Default items per page
	}
}

func (lm LatestModel) Init() tea.Cmd {
	return nil
}

func (lm LatestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd // Declare cmds here
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "return":
			if len(lm.Choices.VisibleItems()) > 0 && lm.Choices.Index() < len(*lm.products) { // Use exported field, Check bounds
				return lm, commands.HandleDisplayProduct(lm.Choices.Width(), lm.Choices.Height(), (*lm.products)[lm.Choices.Index()]) // Use exported field
			}
		case "n": // Next page
			if lm.CurrentPage < lm.TotalPages {
				lm.CurrentPage++
				// Return a command to load the new page
				return lm, func() tea.Msg {
					return commands.LoadLatestPageMsg{Page: lm.CurrentPage, PerPage: lm.PerPage}
				}
			}
		case "p": // Previous page
			if lm.CurrentPage > 1 {
				lm.CurrentPage--
				// Return a command to load the new page
				return lm, func() tea.Msg {
					return commands.LoadLatestPageMsg{Page: lm.CurrentPage, PerPage: lm.PerPage}
				}
			}
		}
	case commands.LatestResponseMsg:
		if msg.Products == nil {
			return lm, cmd // cmd might be nil here, which is fine
		}

		lm.products = msg.Products
		lm.TotalItems = msg.TotalItems // Store TotalItems
		lm.TotalPages = msg.TotalPages // Store TotalPages

		var items []list.Item
		for _, product := range *msg.Products {
			processedDesc := commands.ExtractSummary(product.Description)

			formattedPrice := commands.FormatPrice(
				product.Prices.Price,
				product.Prices.CurrencyPrefix,
				product.Prices.CurrencySuffix,
				product.Prices.CurrencyMinorUnit,
			)
			onSaleMarker := ""
			if product.OnSale {
				onSaleMarker = " 🏷️"
			}

			var desc string
			if formattedPrice != "" {
				desc = fmt.Sprintf("%s%s | %s", formattedPrice, onSaleMarker, processedDesc)
			} else {
				desc = processedDesc
			}

			items = append(items, LatestListItem{
				title:   html.UnescapeString(product.Title),
				desc:    desc,
				brewery: html.UnescapeString(product.Title),
			})
		}
		lm.Choices.SetSize(msg.Width, msg.Height) // Use exported field
		lm.Choices.SetItems(items)                // Use exported field

		// Update list paginator
		lm.Choices.Paginator.PerPage = lm.PerPage      // Use exported field
		lm.Choices.Paginator.Page = lm.CurrentPage - 1 // list.Paginator is 0-indexed
		lm.Choices.Paginator.TotalPages = lm.TotalPages

		// No explicit cmd needed here unless list.Update itself returns one of interest
		// The list's state is updated by SetItems and paginator settings.
	}

	// Propagate other messages (like key presses for list navigation) to the list model
	var listCmd tea.Cmd
	lm.Choices, listCmd = lm.Choices.Update(msg) // Use exported field
	cmds = append(cmds, listCmd)                 // Ensure 'cmds' is declared if this is the first append

	return lm, tea.Batch(cmds...)
}

func (lm LatestModel) View() tea.View {
	lm.Choices.SetShowHelp(false)                                                              // Use exported field
	lm.Choices.Title = fmt.Sprintf("Latest Beers (Page %d/%d)", lm.CurrentPage, lm.TotalPages) // Use exported field
	if lm.TotalPages == 0 && lm.TotalItems > 0 {                                               // Case where API might return 0 pages but has items (should be 1)
		lm.Choices.Title = fmt.Sprintf("Latest Beers (Page %d/%d)", lm.CurrentPage, 1) // Use exported field
	} else if lm.TotalItems == 0 {
		lm.Choices.Title = "Latest Beers (No items found)" // Use exported field
	}
	return tea.NewView(lm.Choices.View()) // Use exported field
}
