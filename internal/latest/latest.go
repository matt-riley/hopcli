package latest

import (
	"fmt"
	"html"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/matt-riley/hopcli/internal/commands"
)

type LatestModel struct {
	products *commands.Products
	choices  list.Model
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
	return LatestModel{
		choices: list.New(items, list.NewDefaultDelegate(), 0, 0),
	}
}

func (lm LatestModel) Init() tea.Cmd {
	return nil
}

func (lm LatestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			lm.choices.CursorDown()
		case "k", "up":
			lm.choices.CursorUp()
		case "enter", "return":
			return lm, commands.HandleDisplayProduct(lm.choices.Width(), lm.choices.Height(), (*lm.products)[lm.choices.Cursor()])
		}
	case commands.LatestResponseMsg:
		if msg.Err != nil {
			fmt.Printf("Error: %s", msg.Err.Error())
			os.Exit(1)
		}

		lm.products = msg.Products
		var items []list.Item
		for _, product := range *msg.Products {
			unescapedTitle := html.UnescapeString(product.Title.Rendered)
			title := unescapedTitle
			brewery := title
			desc := strings.SplitAfter(html.UnescapeString(product.Description.Rendered), "%")
			items = append(items, LatestListItem{title: title, desc: desc[0][3:], brewery: brewery})
		}
		lm.choices.SetSize(msg.Width, msg.Height)
		lm.choices.SetItems(items)
		_, cmd = lm.choices.Update(msg)
	}
	return lm, cmd
}

func (lm LatestModel) View() string {
	lm.choices.SetShowStatusBar(false)
	lm.choices.SetShowHelp(false)
	lm.choices.Title = "The Latest Beers At The Hoptimist"
	return lm.choices.View()
}
