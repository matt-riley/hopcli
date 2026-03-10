package productview

import (
	"fmt"
	"html"

	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	md "github.com/JohannesKaufmann/html-to-markdown/v2"

	"github.com/matt-riley/hopcli/internal/commands"
)

type (
	Product struct {
		Title       string
		Description string
		URL         string
		Price       string // formatted price string, e.g. "£4.20"; empty if unavailable
	}
	ProductModel struct {
		Product Product
		Width   int
		Height  int
	}
)

func NewProductModel() ProductModel {
	return ProductModel{}
}

func (pm ProductModel) Init() tea.Cmd {
	return nil
}

func (pm ProductModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		pm.Width = msg.Width
		pm.Height = msg.Height
		return pm, nil
	case commands.ProductsMsg:
		if msg.Err != nil {
			return pm, nil
		}
		markdownDesc, err := md.ConvertString(msg.Product.Description)
		if err != nil {
			markdownDesc = html.UnescapeString(msg.Product.Description)
		}

		price := commands.FormatPrice(
			msg.Product.Prices.Price,
			msg.Product.Prices.CurrencyPrefix,
			msg.Product.Prices.CurrencySuffix,
			msg.Product.Prices.CurrencyMinorUnit,
		)

		pm.Width = msg.Width
		pm.Product = Product{
			Title:       html.UnescapeString(msg.Product.Title),
			Description: markdownDesc,
			URL:         msg.Product.Link,
			Price:       price,
		}
		return pm, nil
	}
	return pm, nil
}

func (pm ProductModel) View() tea.View {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap((pm.Width/3)*2),
	)
	if err != nil {
		return tea.NewView("")
	}

	var priceSection string
	if pm.Product.Price != "" {
		priceSection = fmt.Sprintf("**Price: %s**\n\n", pm.Product.Price)
	}
	txt, err := renderer.Render(fmt.Sprintf(
		"# %s\n\n%s%s\n\n[View product](%s)",
		pm.Product.Title,
		priceSection,
		pm.Product.Description,
		pm.Product.URL,
	))
	if err != nil {
		return tea.NewView("")
	}
	return tea.NewView(txt)
}
