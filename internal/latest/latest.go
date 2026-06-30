package latest

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/matt-riley/hopcli/internal/commands"
)

type LatestModel struct {
	commands.PaginatedModel
	products *commands.Products
	Choices  list.Model // Exported
}

func NewLatestModel() LatestModel {
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.SetShowPagination(true) // Show list's pagination
	l.SetShowStatusBar(true)  // Ensure status bar is shown for pagination info
	l.SetShowHelp(false)      // Hide default help to keep the view clean

	return LatestModel{
		PaginatedModel: commands.PaginatedModel{
			CurrentPage: 1,
			PerPage:     10, // Default items per page
		},
		Choices: l, // Use exported field
	}
}

func (lm LatestModel) Init() tea.Cmd {
	return nil
}

func (lm LatestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check for 'n'/'p' navigation first
		if pageChanged, newPage := lm.PaginatedModel.UpdatePageNavigation(msg); pageChanged {
			return lm, func() tea.Msg {
				return commands.LoadLatestPageMsg{Page: newPage, PerPage: lm.PerPage}
			}
		}
		switch msg.String() {
		case "enter", "return":
			if len(lm.Choices.VisibleItems()) > 0 && lm.Choices.Index() < len(*lm.products) {
				return lm, commands.HandleDisplayProduct(lm.Choices.Width(), lm.Choices.Height(), (*lm.products)[lm.Choices.Index()])
			}
		}
	case commands.LatestResponseMsg:
		if err := commands.ResponseError(msg); err != nil {
			return lm, nil
		}
		if msg.Products == nil {
			return lm, nil
		}

		lm.products = msg.Products
		lm.TotalItems = msg.TotalItems
		lm.TotalPages = msg.TotalPages

		// Set the list title here (pure View() — no side effects).
		if lm.TotalItems == 0 {
			lm.Choices.Title = "Latest Beers (No items found)"
		} else if lm.TotalPages == 0 {
			// Edge case: API says 0 pages but has items — treat as page 1.
			lm.Choices.Title = fmt.Sprintf("Latest Beers (Page %d/%d)", lm.CurrentPage, 1)
		} else {
			lm.Choices.Title = fmt.Sprintf("Latest Beers (Page %d/%d)", lm.CurrentPage, lm.TotalPages)
		}

		var items []list.Item
		for _, product := range *msg.Products {
			items = append(items, commands.NewProductListItem(product))
		}
		lm.Choices.SetSize(msg.Width, msg.Height)
		lm.Choices.SetItems(items)

		// Update list paginator
		lm.Choices.Paginator.PerPage = lm.PerPage
		lm.Choices.Paginator.Page = lm.CurrentPage - 1 // list.Paginator is 0-indexed
		lm.Choices.Paginator.TotalPages = lm.TotalPages
	}

	// Propagate other messages (like key presses for list navigation) to the list model
	var listCmd tea.Cmd
	lm.Choices, listCmd = lm.Choices.Update(msg)
	cmds = append(cmds, listCmd)

	return lm, tea.Batch(cmds...)
}

func (lm LatestModel) View() tea.View {
	if lm.products == nil {
		return tea.NewView("Loading latest beers...")
	}
	return tea.NewView(lm.Choices.View())
}
