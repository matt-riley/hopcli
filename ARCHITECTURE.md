# hopcli Architecture

## Overview

hopcli is a terminal UI (TUI) for browsing [The Hoptimist](https://thehoptimist.co.uk) online craft beer store. It follows the **Elm Architecture** (Model → Update → View) through the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework. User input and async API responses flow through a single `Update` function that produces a new model and zero or more side-effect commands.

## Package Layout

```
hopcli/
├── cmd/hopt/                   # Main binary — TUI orchestration
│   ├── hopt.go                 # MainModel, navigation, message dispatch
│   └── hopt_test.go            # Boundary tests for model init and Run()
├── internal/
│   ├── api/                    # Pure HTTP client (no TUI dependency)
│   │   ├── api.go              # Client interface, HTTPClient, Product/Category types
│   │   └── api_test.go         # HTTP client tests
│   ├── commands/               # Shared types, message structs, command handlers
│   │   ├── commands.go         # Message types (Bubble Tea Msg), Handle* funcs
│   │   └── commands_test.go    # Tests for formatting, pagination, error handling
│   ├── default/                # Home screen (Latest / Categories choice)
│   │   └── default.go          # DefaultModel, StartLoading*Msg types
│   ├── latest/                 # Latest products paginated list
│   │   ├── latest.go           # LatestModel (embeds PaginatedModel)
│   │   └── latest_test.go      # LatestModel tests
│   ├── categories/             # Category browser
│   │   ├── categories.go       # Model with category list
│   │   └── categories_test.go  # Category model tests
│   ├── categoryproducts/       # Products within a selected category
│   │   ├── categoryproducts.go # Model (embeds PaginatedModel)
│   │   └── categoryproducts_test.go
│   └── product/                # Product detail view
│       ├── product.go          # ProductModel with markdown rendering
│       └── product_test.go     # Product model tests
├── main_test.go                # Top-level integration boundary tests
├── go.mod / go.sum             # Go module definition
├── README.md                   # User-facing documentation
├── ARCHITECTURE.md             # This file
└── LICENSE                     # MIT License
```

### Dependency arrows

```
cmd/hopt
  ├── internal/api          (via commands type aliases and ApiClient)
  ├── internal/commands     (message types, Handle* functions, shared helpers)
  ├── internal/default
  ├── internal/latest
  ├── internal/categories
  ├── internal/categoryproducts
  └── internal/product

internal/commands
  └── internal/api          (type aliases + ApiClient)
```

Key design rule: `internal/api` has **zero** dependency on Bubble Tea. It is a pure HTTP client that can be tested in isolation and reused outside the TUI. `internal/commands` bridges the two worlds — it defines Bubble Tea `Msg` types and `tea.Cmd` functions that call `api.Client`, and it re-exports API types via type aliases for backward compatibility.

## Model Hierarchy (ASCII Diagram)

```
                        ┌──────────────────────────────────┐
                        │           MainModel              │
                        │  (cmd/hopt/hopt.go)              │
                        │                                  │
                        │  State         SessionState      │
                        │  CurrentView   tea.Model         │
                        │  PreviousViews []tea.Model  ◄────┤ stack
                        │  requestID     int               │
                        │  Loading       bool              │
                        │  Spinner       spinner.Model     │
                        │  ErrMsg        string            │
                        │                                  │
                        │  DefaultModel          ──────────┤
                        │  LatestModel           ──────────┤
                        │  CategoriesModel       ──────────┤
                        │  CategoryProductsModel ──────────┤
                        │  ProductModel          ──────────┤
                        └──────┬───────────┬───────────────┘
                               │ owns      │ delegates to
                               ▼           ▼
          ┌─────────────────────────────────────────────────────┐
          │                  Sub-models                         │
          │                                                     │
          │  ┌─────────────────┐  ┌───────────────────────────┐ │
          │  │ DefaultModel    │  │ LatestModel               │ │
          │  │ (default.go)    │  │ (latest.go)               │ │
          │  │                 │  │                            │ │
          │  │ Choices list    │  │ PaginatedModel (embedded)  │ │
          │  │ "Latest"        │  │   CurrentPage, PerPage,    │ │
          │  │ "Categories"    │  │   TotalItems, TotalPages   │ │
          │  └────────┬────────┘  │ Choices list.Model         │ │
          │           │           │ products *Products         │ │
          │           │           └────────────┬──────────────┘ │
          │           │                        │                │
          │  ┌────────┴────────┐  ┌────────────┴──────────────┐ │
          │  │ Categories      │  │ CategoryProducts          │ │
          │  │ Model           │  │ Model                     │ │
          │  │ (categories.go) │  │ (categoryproducts.go)     │ │
          │  │                 │  │                            │ │
          │  │ List list.Model │  │ PaginatedModel (embedded)  │ │
          │  │ CategoryListItem│  │ List list.Model            │ │
          │  └────────┬────────┘  │ categoryID, categoryName   │ │
          │           │           │ products *Products         │ │
          │           │           └────────────┬──────────────┘ │
          │           │                        │                │
          │           │           ┌────────────┴──────────────┐ │
          │           │           │ ProductModel              │ │
          │           │           │ (product.go)              │ │
          │           │           │                            │ │
          │           │           │ Product struct:            │ │
          │           │           │   Title, Description,      │ │
          │           │           │   URL, Price               │ │
          │           │           │ Width, Height              │ │
          │           │           └───────────────────────────┘ │
          └─────────────────────────────────────────────────────┘

          ┌─────────────────────────────────────────────────────┐
          │                  Shared Layer                       │
          │  (internal/commands/commands.go)                    │
          │                                                     │
          │  ProductListItem    — shared list.Item for products │
          │  PaginatedModel     — embedded pagination state     │
          │  UpdatePageNavigation — handles 'n'/'p' keys        │
          │  ResponseError      — unified error extraction      │
          │  FormatPrice        — minor-unit price formatting   │
          │  ExtractSummary     — HTML → plain-text summary     │
          │  HandleGetLatest(), HandleGetCategories(), etc.     │
          └──────────────────────┬──────────────────────────────┘
                                 │ calls
                                 ▼
          ┌─────────────────────────────────────────────────────┐
          │                  API Layer                           │
          │  (internal/api/api.go)                               │
          │                                                     │
          │  Client interface — FetchProducts, FetchCategories,  │
          │    FetchProduct, FetchProductsByCategory             │
          │  HTTPClient — doRequest() with retry, User-Agent,    │
          │    timeout, Content-Type validation                  │
          │  RetryConfig — exponential backoff, max duration     │
          └─────────────────────────────────────────────────────┘
```

## View Lifecycle (Elm-like Model → Update → View)

Each view in hopcli follows the Bubble Tea pattern, which is directly analogous to the Elm Architecture:

```
        ┌──────────┐
        │   Init   │  → returns initial tea.Cmd (side effects)
        └────┬─────┘
             │
             ▼
        ┌──────────┐    tea.Msg     ┌──────────┐
        │  Update  │ ◄──────────────│ Runtime  │
        └────┬─────┘                └──────────┘
             │                           ▲
             │ returns (Model, tea.Cmd)  │ tea.Cmd produces Msg
             ▼                           │
        ┌──────────┐                     │
        │   View   │  → string (rendered TUI, displayed to user)
        └──────────┘
```

### MainModel.Update flow

1. User presses a key → `tea.KeyMsg` arrives
2. `MainModel.Update` pattern-matches the key:
   - `q` → `tea.Quit`
   - `h`/`←` → pop navigation stack (see below)
   - Other keys → delegate to `CurrentView.Update(msg)`
3. If not handled, the message is forwarded to `CurrentView.Update`
4. Sub-model may return a `tea.Cmd` (e.g., `HandleGetLatest(...)`) that performs async work
5. When the async work completes, it returns a response `Msg` (e.g., `LatestResponseMsg`)
6. `MainModel.Update` matches the response, updates the appropriate sub-model, and switches `CurrentView` if needed

### Sub-model boilerplate

Every sub-model implements the `tea.Model` interface:

```go
type tea.Model interface {
    Init() tea.Cmd
    Update(tea.Msg) (tea.Model, tea.Cmd)
    View() tea.View
}
```

`MainModel` stores a typed reference to each sub-model (e.g., `LatestModel`, `CategoriesModel`) so it can sync state back after updates via type assertion:

```go
if lm, ok := updatedViewModel.(latest.LatestModel); ok {
    mm.LatestModel = lm
}
```

## Navigation Stack

hopcli maintains a linear navigation stack via `MainModel.PreviousViews`, a `[]tea.Model` slice.

### Forward navigation

When the user selects an item that opens a new view (e.g., Latest → Product Detail), the current view is pushed onto the stack:

```go
mm.PreviousViews = append(mm.PreviousViews, mm.CurrentView)
mm.CurrentView = mm.ProductModel
mm.State = ProductView
```

### Back navigation (h / ←)

Back navigation pops the stack:

```go
if len(mm.PreviousViews) > 0 {
    lastViewIndex := len(mm.PreviousViews) - 1
    mm.CurrentView = mm.PreviousViews[lastViewIndex]
    mm.PreviousViews = mm.PreviousViews[:lastViewIndex]
    // restore State based on CurrentView type
}
```

Important: on back navigation, the `requestID` is incremented and `Loading` is set to `false`. This cancels any in-flight request — its response, when it arrives, will carry a stale `RequestID` and be silently dropped.

### View states

`SessionState` is an enum tracking which view is active:

```go
const (
    DefaultView SessionState = iota
    LatestView
    CategoriesView
    CategoryProductsView
    ProductView
)
```

It is used in `View()` to select the correct help bar text and in `Update()` to guard against stale messages from views the user has already left.

## Request-ID Invalidation Pattern

Concurrent TUI applications face a classic problem: the user triggers a request, navigates away, and the response arrives for a view no longer shown. hopcli solves this with a monotonically increasing `requestID` counter.

### Mechanism

1. `MainModel.requestID` starts at 0 (zero value)
2. Every new API request increments `requestID`:
   ```go
   mm.requestID++
   cmds = append(cmds, commands.HandleGetLatest(..., mm.requestID))
   ```
3. The response message carries the `requestID` at the time the request was sent:
   ```go
   type LatestResponseMsg struct {
       // ...payload...
       RequestID int
   }
   ```
4. On receiving a response, `MainModel.Update` compares:
   ```go
   if msg.RequestID != mm.requestID {
       break // stale — discard
   }
   ```
5. Back navigation also increments `requestID`, invalidating any in-flight request from the view being left

### NavGen — child-generated message invalidation

Some messages originate from sub-models rather than from API responses (e.g., `ProductsMsg` from `HandleDisplayProduct`, `LoadLatestPageMsg` from page navigation). These carry a `NavGen` field instead of `RequestID`:

```go
type LoadLatestPageMsg struct {
    Page   int
    NavGen int
}
```

The `wrapChildCmd` function stamps `NavGen` with the current `requestID` at the time the sub-model's `Update` produced the command:

```go
func wrapChildCmd(cmd tea.Cmd, gen int) tea.Cmd {
    return func() tea.Msg {
        msg := cmd()
        switch m := msg.(type) {
        case commands.ProductsMsg:
            m.NavGen = gen
            return m
        // ... other NavGen-bearing types ...
        }
        return msg
    }
}
```

When the message arrives, a stale `NavGen` (user navigated away in the meantime) is detected and the message is discarded:

```go
if msg.NavGen != mm.requestID {
    break // stale
}
```

### Race condition protection

This pattern is necessary because Bubble Tea commands run asynchronously. Between the time a sub-model returns a command and the time that command's message arrives at `MainModel.Update`, the user may have pressed `h` to go back, changing the view. Without the invalidation pattern, the product detail view could overwrite the categories view, or a page-load could populate a list the user already left.

## Message Flow

### Example: User selects "Latest" → views a product

```
User presses Enter on "Latest"
  │
  ▼
DefaultModel.Update(tea.KeyMsg{Enter})
  │  returns (DefaultModel, func() Msg { return StartLoadingLatestMsg{} })
  ▼
MainModel.Update(StartLoadingLatestMsg{})
  │  mm.requestID++  (→ 1)
  │  returns (MainModel, HandleGetLatest(w, h, 1, 10, 1))
  ▼
HandleGetLatest goroutine
  │  ApiClient.FetchProducts(ctx, 1, 10)
  │  returns LatestResponseMsg{Products: ..., RequestID: 1}
  ▼
MainModel.Update(LatestResponseMsg{RequestID: 1})
  │  msg.RequestID (1) == mm.requestID (1) ✓
  │  mm.LatestModel.Update(msg)
  │  mm.CurrentView = mm.LatestModel
  │  mm.State = LatestView
  │  mm.Loading = false
  ▼
User presses Enter on a product
  │
  ▼
LatestModel.Update(tea.KeyMsg{Enter})
  │  returns (LatestModel, HandleDisplayProduct(...))
  ▼
MainModel receives ProductsMsg{NavGen: 1} from wrapChildCmd
  │  msg.NavGen (1) == mm.requestID (1) ✓
  │  mm.PreviousViews = append(mm.PreviousViews, mm.CurrentView)
  │  mm.CurrentView = mm.ProductModel
  │  mm.State = ProductView
  ▼
User sees product detail
```

### Example: Stale response rejection

```
User presses Enter on "Latest"
  │  mm.requestID = 1
  │  HandleGetLatest(..., requestID=1) starts fetching
  ▼
User presses 'h' (back)
  │  mm.requestID → 2 (incremented)
  │  mm.Loading = false
  │  mm.CurrentView popped from PreviousViews
  ▼
Network response arrives: LatestResponseMsg{RequestID: 1}
  │  msg.RequestID (1) != mm.requestID (2) ✗
  │  → break (message discarded, no state change)
```

### Example: Pagination

```
User presses 'n' (next page)
  │
  ▼
LatestModel.Update(tea.KeyMsg{"n"})
  │  PaginatedModel.UpdatePageNavigation → pageChanged=true, newPage=2
  │  returns (LatestModel, func() Msg { LoadLatestPageMsg{Page: 2} })
  ▼
wrapChildCmd stamps NavGen with current requestID
  ▼
MainModel.Update(LoadLatestPageMsg{NavGen: current})
  │  msg.NavGen == mm.requestID ✓
  │  mm.State == LatestView ✓
  │  mm.requestID++
  │  HandleGetLatest(w, h, 2, 10, newRequestID)
  ▼
Response arrives, same flow as initial load
```

## Key Design Decisions

### api.Client interface

The `api.Client` interface decouples HTTP logic from TUI concerns. The global `commands.ApiClient` variable can be swapped for testing:

```go
var ApiClient api.Client  // set by main, replaceable in tests
```

The `HTTPClient` implementation provides:
- Exponential backoff retry (configurable via `RetryConfig`)
- User-Agent header (`hopcli/<version>`)
- Content-Type validation (rejects non-JSON responses)
- Per-request timeout (default 5s)

### Shared composable types

To eliminate duplication between `latest` and `categoryproducts` packages:

| Type | Purpose | Used by |
|---|---|---|
| `ProductListItem` | `list.Item` implementation for products | latest, categoryproducts |
| `PaginatedModel` | Embedded pagination state + `UpdatePageNavigation` | LatestModel, Model (categoryproducts) |
| `ResponseError` | Unified error extraction from response Msg types | All sub-models |
| `FormatPrice` | Minor-unit price → display string | ProductListItem, ProductModel |
| `ExtractSummary` | HTML description → plain-text summary | ProductListItem |

### Type aliases in commands

`internal/commands` re-exports API types as type aliases:

```go
type Product = api.Product
type Products = api.Products
type Category = api.Category
```

This preserves backward compatibility for packages that imported these types from `commands` before they were moved to `api`.

## Testing

Tests follow table-driven patterns using [matryer/is](https://github.com/matryer/is) for assertions:

- `internal/api/api_test.go` — HTTP client with mock server
- `internal/commands/commands_test.go` — FormatPrice, ExtractSummary, PaginatedModel, ResponseError
- `internal/latest/latest_test.go` — LatestModel Update/View
- `internal/categories/categories_test.go` — Category model
- `internal/categoryproducts/categoryproducts_test.go` — CategoryProducts model
- `internal/product/product_test.go` — Product model
- `cmd/hopt/hopt_test.go` — MainModel initialization and Run boundary
- `main_test.go` — Integration boundary tests

Run all tests with the race detector:

```bash
go test -race ./...
```
