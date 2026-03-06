package productview

import (
	"fmt"
	"html"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	md "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/charmbracelet/glamour"

	"github.com/matt-riley/hopcli/internal/commands"
)

type (
	Product struct {
		Title       string
		Description string
		URL         string
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
	case commands.ProductsMsg:
		if msg.Err != nil {
			return pm, nil
		}
		md, err := md.ConvertString(msg.Product.Description.Rendered)
		if err != nil {
			return pm, nil
		}
		splitString := strings.Split(md, "%")
		formatted := fmt.Sprintf("%s\n\n%s", splitString[0], strings.TrimLeftFunc(splitString[1], unicode.IsSpace))
		pm.Width = msg.Width
		pm.Product = Product{
			Title:       html.UnescapeString(msg.Product.Title.Rendered),
			Description: formatted,
			URL:         msg.Product.Link,
		}
		return pm, nil
	}
	return pm, nil
}

func (pm ProductModel) View() tea.View {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(), glamour.WithWordWrap((pm.Width/3)*2),
	)
	if err != nil {
		return tea.NewView("")
	}

	txt, err := renderer.Render(fmt.Sprintf("# %s\n\n%s\n\n[LINK](%s)", pm.Product.Title, pm.Product.Description, pm.Product.URL))
	if err != nil {
		return tea.NewView("")
	}
	return tea.NewView(txt)
}
