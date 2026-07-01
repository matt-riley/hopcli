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
		ErrMsg  string // error message to display in View()
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
		if err := commands.ResponseError(msg); err != nil {
			pm.ErrMsg = fmt.Sprintf("Error loading product: %v", err)
			return pm, nil
		}
		if msg.Product == nil {
			pm.ErrMsg = "Error: product data is missing"
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
	if pm.ErrMsg != "" {
		return tea.NewView(pm.ErrMsg)
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap((pm.Width/3)*2),
	)
	if err != nil {
		return tea.NewView(fmt.Sprintf("Error initializing renderer: %v", err))
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
		return tea.NewView(fmt.Sprintf("Error rendering product: %v", err))
	}
	return tea.NewView(txt)
}
