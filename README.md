<a href="https://thehoptimist.co.uk"><img src="https://github.com/matt-riley/hopcli/assets/2221636/ce43318d-6d4d-41c6-b838-ac772070aacd" width="100%" /></a>

# hopcli

An **UNOFFICIAL** terminal UI (TUI) for browsing [The Hoptimist](https://thehoptimist.co.uk) online craft beer store. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and the [WooCommerce Store API](https://woocommerce.com/document/woocommerce-store-api/).

## Features

- Browse the latest items added to the store with pagination (`n` next / `p` previous)
- Explore products by category
- View product details with markdown-rendered descriptions and prices
- Full keyboard navigation (arrows, enter, back with `h`/`←`, quit with `q`)
- Loading spinner and error display for a smooth TUI experience

## Demo

![Hoptimist CLI Demo](assets/demo.gif)

The demo shows the main flow: landing screen → latest products → selecting a product for details → browsing categories → exploring products within a category.

## Prerequisites

- **Go** 1.25 or later (see [go.mod](go.mod) for the exact minimum version)
- A terminal with 256-color support (the TUI uses Lipgloss styling)

## Installation

### go install

```bash
go install github.com/matt-riley/hopcli@latest
```

### go build

```bash
git clone https://github.com/matt-riley/hopcli.git
cd hopcli
go build -o hopt .
./hopt
```

### Build with version info

```bash
go build -ldflags "-X github.com/matt-riley/hopcli/internal/commands.Version=$(git describe --tags --always --dirty)" -o hopt .
```

### Homebrew

```bash
brew install matt-riley/tap/hopt
```

## Configuration

By default, hopcli connects to the production API at `https://thehoptimist.co.uk`. To point it at a different WooCommerce store:

```go
// In your own main.go or before calling hopt.Run():
commands.TheHoptimistBaseURL = "https://your-store.example.com"
```

The API client (`commands.ApiClient`) can also be replaced with a custom implementation that satisfies the `api.Client` interface — useful for testing or working with a different backend.

## Usage

| Key            | Action                                              |
|----------------|-----------------------------------------------------|
| `↑` / `↓`      | Navigate lists                                      |
| `Enter`        | Select item (open category, view product details)   |
| `h` / `←`      | Go back to the previous view                        |
| `n`            | Next page (in paginated views)                      |
| `p`            | Previous page (in paginated views)                  |
| `q`            | Quit                                                |

### Views

1. **Home** — Choose between *Latest* and *Categories*
2. **Latest Beers** — Paginated list of recently added products
3. **Categories** — Browse all product categories
4. **Category Products** — Paginated list of products within a selected category
5. **Product Detail** — Markdown-rendered description, price, and link to the store page

## Known Limitations

### Brewery / Brand Information

The upstream WooCommerce Store API does not expose brewery or brand data for products. When displaying product details, the product title is returned as-is — there is no separate brewery field to display. This is a limitation of the data model provided by The Hoptimist's API, not a bug in hopcli.

If brewery data becomes available through the API in the future, it can be surfaced by adding a `Brewery` field to `api.Product` and threading it through the `product.ProductModel`.

## License

MIT — see [LICENSE](LICENSE) for full text.
