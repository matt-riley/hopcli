# Copilot instructions for `hopcli`

## Build, test, and lint

- Build everything with `go build ./...`.
- Run the full test suite with `go test ./...`.
- Run a single package or test with Go's normal selectors, for example:
  - `go test ./internal/commands`
  - `go test ./internal/commands -run TestFormatPrice`
  - `go test ./cmd/hopt -run TestMainModelUpdate_StateTransitions`
- Lint with `golangci-lint run`. The repo currently has pre-existing findings in the baseline, so use lint output carefully and avoid treating "zero findings" as an assumption.

## High-level architecture

- `main.go` only delegates to `cmd/hopt.Run()`. The real application shell lives in `cmd/hopt/hopt.go`.
- `cmd/hopt/hopt.go` owns the top-level Bubble Tea state machine:
  - `MainModel` keeps the active screen in `CurrentView`, plus typed copies of each sub-model.
  - `PreviousViews` is the back-stack for `h` / left-arrow navigation.
  - Loading and error states are rendered centrally, not inside each sub-view.
- Each screen under `internal/` is its own Bubble Tea model:
  - `internal/default`: home menu with the two entry flows from the README (`Latest` and `Categories`).
  - `internal/latest`: paginated latest-products list.
  - `internal/categories`: category list.
  - `internal/categoryproducts`: paginated products for the selected category.
  - `internal/product`: product detail screen; converts HTML descriptions to Markdown and renders them with `glamour`.
- `internal/commands` is the shared command/data layer. It defines:
  - message types exchanged between the top-level model and child views,
  - HTTP commands for the WooCommerce Store API under `/wp-json/wc/store/v1/...`,
  - normalization helpers such as `FormatPrice` and `ExtractSummary`.

## Key conventions

- Keep HTTP and API shaping in `internal/commands`, not in the view packages. The view models emit messages/commands; `commands` performs the actual fetches and returns typed response messages.
- Preserve the stale-message guards in `cmd/hopt/hopt.go` when changing navigation or async loading:
  - `requestID` is incremented for each in-flight load and also on back-navigation to invalidate older work.
  - `wrapChildCmd` stamps child-generated messages with `NavGen`.
  - Response handlers explicitly drop messages whose `RequestID` or `NavGen` no longer matches the active navigation generation.
- When adding a new screen or new message flow, update both halves of the top-level router in `MainModel.Update`:
  - the message-handling switch that transitions between states,
  - the sync-back section that writes `CurrentView` back into the typed sub-model fields.
- Reuse `commands.FormatPrice` and `commands.ExtractSummary` instead of reformatting WooCommerce data inside list/detail views. The current UI consistently derives list subtitles and product pricing from those helpers.
- API-focused tests in `internal/commands/commands_test.go` override `commands.TheHoptimistBaseURL` and use `httptest` servers. Follow that pattern for new API tests instead of hitting the real Hoptimist site.
- The broad state-machine coverage lives in `cmd/hopt/hopt_test.go`, while the `internal/*` packages keep smaller per-view tests. Put new tests alongside the layer you are changing.
