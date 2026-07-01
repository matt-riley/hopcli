package hopt

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/matt-riley/hopcli/internal/api"
	"github.com/matt-riley/hopcli/internal/categories"
	"github.com/matt-riley/hopcli/internal/categoryproducts"
	"github.com/matt-riley/hopcli/internal/commands"
	defaultview "github.com/matt-riley/hopcli/internal/default"
	"github.com/matt-riley/hopcli/internal/latest"
	productview "github.com/matt-riley/hopcli/internal/product"
)

var logo = `
⠀⠀⠀⠀⠀⠀⠀⢀⣰⣶⣶⣤⣤⣤⣤⣤⣀⠀⣀⣀⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⣀⣤⣤⣤⣬⣽⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣄⠀⠀⠀⠀⠀⠀
⠀⠐⠚⣻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢿⣿⣿⣿⣿⣿⣷⣄⣄⠀⠀⠀
⠐⣶⣾⣿⣿⣿⢿⣿⣿⣿⣿⣦⣍⠻⣿⣿⣿⣿⣦⡙⢿⣿⣿⣿⣿⣿⣿⣷⡄⠀
⠀⢿⣿⣿⣿⣿⣶⣍⠛⢿⣿⣿⣿⣷⡈⠻⣿⣿⣿⣿⡄⠹⣿⣿⣿⣿⣿⣿⣷⡀
⢠⣌⣿⣿⣿⡟⠁⠀⠀⠀⠈⠉⠛⠛⠛⠂⠈⠛⠋⠉⠁⠀⠀⠀⠙⣱⣿⣿⣿⡇
⠘⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⡿⠁
⠀⢹⣿⣿⡟⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢻⣿⣿⠇⠀
⠀⠀⢻⣿⣀⣠⠤⠤⠤⠤⠤⠤⣄⡀⠀⠀⠀⣀⡤⠤⠤⠤⠤⢤⣀⣈⣿⡟⠀⠀
⠠⡖⢸⡟⡟⠀⠀⠀⠀⠀⠀⠀⠀⢹⡶⢶⡏⠀⠀⠀⠀⠀⠀⠀⠈⣿⢻⡇⢾⡄
⠀⡇⠀⡇⠰⡀⠀⠀⠀⠀⠀⠀⢠⠟⠀⠈⢣⡀⠀⠀⠀⠀⠀⠀⢠⠏⢸⠁⢸⠃
⠀⠱⠀⡇⠀⠙⠢⠤⠤⠤⠤⠖⠋⠀⠀⠀⠀⠙⠢⠤⠤⠤⠤⠴⠋⠀⣼⠀⠃⠀
⠀⠀⠀⣿⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⡏⠀⠀⠀
⠀⠀⠀⣿⣧⠀⠀⠀⠀⢀⣠⣴⣶⡷⠶⠶⢶⣶⣦⣤⣀⠀⠀⠀⠀⣾⡇⠀⠀⠀
⠀⠀⠀⠘⣿⣦⠀⠀⣴⡟⠉⠉⠁⢀⣀⣀⡀⠀⠈⠉⠻⣧⡀⢀⣼⣿⡇⠀⠀⠀
⠀⠀⠀⠀⢸⣿⣷⣼⣿⣤⣤⣤⣴⣿⣿⣿⣿⣦⣤⣤⣤⣿⣷⣼⣿⣿⠃⠀⠀⠀
⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⣿⡿⢿⣿⣿⣿⣿⡿⢿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀
⠀⠀⠀⠀⢼⣿⣿⡿⠿⠛⣉⣴⡈⢿⣿⣿⡿⢡⣦⣙⠻⠿⢿⣿⣿⡇⠀⠀⠀⠀
⠀⠀⠀⠀⠀⣶⣶⣶⠀⣿⣿⣿⣿⣦⠙⢋⣴⣿⣿⣿⣿⠀⣶⣶⡦⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⢿⣿⡿⠀⠿⠿⠿⢛⣡⣶⣷⣌⠛⠿⠿⠿⠀⣿⣿⡟⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⢰⣶⣶⣶⣷⠘⣿⣿⣿⡿⢁⣿⣶⣶⣶⡆⠈⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠈⢿⣿⣿⡿⠓⣈⡛⢛⣁⠺⢿⣿⣿⡿⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠀⠐⢿⣿⣿⣿⣿⡗⠀⠈⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⠿⠟⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀`

type SessionState uint

const (
	DefaultView SessionState = iota
	LatestView
	CategoriesView
	CategoryProductsView
	ProductView
)

type MainModel struct {
	State                 SessionState
	CurrentView           tea.Model
	PreviousViews         []tea.Model // Stack to store previous views for back navigation
	DefaultModel          defaultview.DefaultModel
	LatestModel           latest.LatestModel
	CategoriesModel       categories.Model
	CategoryProductsModel categoryproducts.Model
	ProductModel          productview.ProductModel
	Loading               bool
	Spinner               spinner.Model
	Width                 int
	Height                int
	ErrMsg                string // To store and display error messages
	requestID             int    // Tracks the current in-flight request; used to discard stale responses

	// Context cancellation for aborting in-flight HTTP requests on back-navigation.
	cancelFunc context.CancelFunc

	// Categories cache: avoid re-fetching on subsequent visits to CategoriesView.
	cachedCategories  *api.Categories
	categoriesFetched bool

	// Debounce state for rapid 'n'/'p' keypresses.
	debouncePage         int
	debounceCategoryID   int
	debounceCategoryName string
	debounceGen          int
	debounceMode         string // "latest" or "category"
}

// debounceFireMsg is delivered when a debounce timer expires.
// Only the latest generation is processed; stale ticks are dropped.
type debounceFireMsg struct {
	gen  int
	mode string
}

func InitialModel() MainModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(commands.SpinnerColor))

	dm := defaultview.InitialModel()
	lm := latest.NewLatestModel()
	cm := categories.NewCategoriesModel()
	// For cpm, categoryName and categoryID will be updated when a category is chosen.
	// Initialize with placeholder values.
	cpm := categoryproducts.NewModel("", 0)
	pm := productview.NewProductModel()

	return MainModel{
		State:                 DefaultView,
		CurrentView:           dm,
		DefaultModel:          dm,
		LatestModel:           lm,
		CategoriesModel:       cm,
		CategoryProductsModel: cpm,
		ProductModel:          pm,
		PreviousViews:         []tea.Model{},
		Loading:               false,
		Spinner:               s,
		// Width and Height will be set by tea.WindowSizeMsg
	}
}

func (mm MainModel) Init() tea.Cmd {
	return mm.Spinner.Tick
}

func (mm MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	var updatedViewModel tea.Model
	handled := false

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mm.Width = msg.Width
		mm.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if mm.cancelFunc != nil {
				mm.cancelFunc()
			}
			return mm, tea.Quit
		case "h", "left":
			mm.Loading = false // cancel any in-flight load
			if mm.cancelFunc != nil {
				mm.cancelFunc() // abort in-flight HTTP request
			}
			mm.requestID++ // invalidate any in-flight request
			mm.ErrMsg = "" // always clear error on back-nav
			if len(mm.PreviousViews) > 0 {
				lastViewIndex := len(mm.PreviousViews) - 1
				mm.CurrentView = mm.PreviousViews[lastViewIndex]
				mm.PreviousViews = mm.PreviousViews[:lastViewIndex]

				switch mm.CurrentView.(type) {
				case defaultview.DefaultModel:
					mm.State = DefaultView
				case latest.LatestModel:
					mm.State = LatestView
				case categories.Model:
					mm.State = CategoriesView
				case categoryproducts.Model:
					mm.State = CategoryProductsView
				case productview.ProductModel:
					mm.State = ProductView
				}

				if mm.Width > 0 && mm.Height > 0 {
					var resized tea.Model
					var resizeCmd tea.Cmd
					resized, resizeCmd = mm.CurrentView.Update(
						tea.WindowSizeMsg{Width: mm.Width, Height: mm.Height})
					mm.CurrentView = resized
					cmds = append(cmds, resizeCmd)
					switch v := mm.CurrentView.(type) {
					case defaultview.DefaultModel:
						mm.DefaultModel = v
					case latest.LatestModel:
						mm.LatestModel = v
					case categories.Model:
						mm.CategoriesModel = v
					case categoryproducts.Model:
						mm.CategoryProductsModel = v
					case productview.ProductModel:
						mm.ProductModel = v
					}
				}
			}
			return mm, tea.Batch(cmds...)
		case "r":
			if mm.State == CategoriesView {
				mm.categoriesFetched = false
				mm.cachedCategories = nil
				mm.Loading = true
				mm.ErrMsg = ""
				mm.requestID++
				ctx, cancel := context.WithCancel(context.Background())
				mm.cancelFunc = cancel
				cmds = append(cmds, commands.HandleGetCategories(ctx, mm.requestID))
				return mm, tea.Batch(cmds...)
			}
		}
	case defaultview.StartLoadingLatestMsg:
		handled = true
		mm.Loading = true
		mm.ErrMsg = ""
		mm.LatestModel.CurrentPage = 1
		mm.requestID++
		mm.debounceGen = 0 // reset debounce on fresh navigation
		ctx, cancel := context.WithCancel(context.Background())
		mm.cancelFunc = cancel
		// Ensure initial load is page 1, use PerPage from the model (defaulted in NewLatestModel)
		cmds = append(cmds, commands.HandleGetLatest(ctx, mm.Width, mm.Height, 1, mm.LatestModel.PerPage, mm.requestID))

	case defaultview.StartLoadingCategoriesMsg:
		handled = true
		mm.ErrMsg = ""
		mm.debounceGen = 0 // reset debounce on fresh navigation
		if mm.categoriesFetched && mm.cachedCategories != nil {
			// Serve from cache — no HTTP request needed.
			cmds = append(cmds, func() tea.Msg {
				return commands.CategoriesResponseMsg{
					Categories: mm.cachedCategories,
					RequestID:  mm.requestID,
				}
			})
		} else {
			mm.Loading = true
			mm.requestID++
			ctx, cancel := context.WithCancel(context.Background())
			mm.cancelFunc = cancel
			cmds = append(cmds, commands.HandleGetCategories(ctx, mm.requestID))
		}

	case commands.LatestResponseMsg:
		handled = true
		if msg.RequestID != mm.requestID {
			break // stale response from a cancelled or superseded request
		}
		mm.Loading = false
		if msg.Err != nil {
			mm.ErrMsg = msg.Err.Error()
		} else {
			// If not already in LatestView, it's an initial load for this view type.
			if mm.State != LatestView {
				// mm.LatestModel = latest.NewLatestModel() // Already initialized in InitialModel
				mm.PreviousViews = append(mm.PreviousViews, mm.CurrentView) // Only add to stack if changing view type
				mm.State = LatestView
			}
			// Update the existing (or newly created) LatestModel instance
			updatedViewModel, cmd = mm.LatestModel.Update(msg)
			if lm, ok := updatedViewModel.(latest.LatestModel); ok {
				mm.LatestModel = lm
			} else {
				log.Printf("unexpected type from LatestModel.Update: %T", updatedViewModel)
				mm.ErrMsg = "internal error: unexpected model type"
				break
			}
			cmds = append(cmds, cmd)
			mm.CurrentView = mm.LatestModel // Ensure CurrentView points to the updated model
			if mm.Width > 0 && mm.Height > 0 {
				resized, sizeCmd := mm.CurrentView.Update(tea.WindowSizeMsg{Width: mm.Width, Height: mm.Height})
				mm.CurrentView = resized
				cmds = append(cmds, sizeCmd)
				if v, ok := mm.CurrentView.(latest.LatestModel); ok {
					mm.LatestModel = v
				}
			}
			mm.ErrMsg = ""
		}

	case commands.LoadLatestPageMsg:
		handled = true
		if msg.NavGen != mm.requestID {
			break // stale message from a previous view visit
		}
		if mm.State != LatestView {
			// Stale message: user navigated away before this cmd was delivered.
			break
		}
		// Debounce: store the latest target page and (re)start a 200ms timer.
		mm.debounceMode = "latest"
		mm.debouncePage = msg.Page
		mm.debounceGen++
		mm.LatestModel.CurrentPage = msg.Page
		mm.CurrentView = mm.LatestModel
		cmds = append(cmds, debounceCmd(200*time.Millisecond, mm.debounceGen, "latest"))

	case commands.CategoriesResponseMsg:
		handled = true
		if msg.RequestID != mm.requestID {
			break // stale response from a cancelled or superseded request
		}
		mm.Loading = false
		if msg.Err != nil {
			mm.ErrMsg = msg.Err.Error()
		} else {
			// Sub-model sizing is handled by tea.WindowSizeMsg re-dispatch.
			updatedViewModel, cmd = mm.CategoriesModel.Update(msg)
			if cm, ok := updatedViewModel.(categories.Model); ok {
				mm.CategoriesModel = cm
			} else {
				log.Printf("unexpected type from CategoriesModel.Update: %T", updatedViewModel)
				mm.ErrMsg = "internal error: unexpected model type"
				break
			}
			cmds = append(cmds, cmd)
			mm.PreviousViews = append(mm.PreviousViews, mm.CurrentView)
			mm.CurrentView = mm.CategoriesModel
			mm.State = CategoriesView
			if mm.Width > 0 && mm.Height > 0 {
				resized, sizeCmd := mm.CurrentView.Update(tea.WindowSizeMsg{Width: mm.Width, Height: mm.Height})
				mm.CurrentView = resized
				cmds = append(cmds, sizeCmd)
				if v, ok := mm.CurrentView.(categories.Model); ok {
					mm.CategoriesModel = v
				}
			}
			mm.ErrMsg = ""
			// Cache categories on successful fetch.
			if msg.Categories != nil {
				mm.cachedCategories = msg.Categories
				mm.categoriesFetched = true
			}
		}

	case commands.StartLoadingProductsForCategoryMsg:
		handled = true
		if msg.NavGen != mm.requestID {
			break // stale message from a previous view visit
		}
		if mm.State != CategoriesView {
			// Stale message: user navigated away before this cmd was delivered.
			break
		}
		mm.Loading = true
		mm.ErrMsg = ""
		mm.CategoryProductsModel.CurrentPage = 1
		mm.requestID++
		ctx, cancel := context.WithCancel(context.Background())
		mm.cancelFunc = cancel
		// Ensure initial load is page 1, use PerPage from the model (defaulted in NewCategoryProductsModel)
		cmds = append(cmds, commands.HandleGetProductsByCategory(ctx, msg.CategoryID, msg.CategoryName, 1, mm.CategoryProductsModel.PerPage, mm.requestID))

	case commands.ProductsForCategoryResponseMsg:
		handled = true
		if msg.RequestID != mm.requestID {
			break // stale response from a cancelled or superseded request
		}
		mm.Loading = false
		if msg.Err != nil {
			mm.ErrMsg = msg.Err.Error()
		} else {
			// Sub-model sizing is handled by tea.WindowSizeMsg re-dispatch.
			// If current state is not CategoryProductsView OR if the category ID differs,
			// it's a new category product listing.
			if mm.State != CategoryProductsView || mm.CategoryProductsModel.CategoryID() != msg.CategoryID { // Used getter
				mm.CategoryProductsModel = categoryproducts.NewModel(msg.CategoryName, msg.CategoryID)
				mm.PreviousViews = append(mm.PreviousViews, mm.CurrentView)
				mm.State = CategoryProductsView
			}
			// Update the model (either newly created or existing)
			updatedViewModel, cmd := mm.CategoryProductsModel.Update(msg)
			if cpm, ok := updatedViewModel.(categoryproducts.Model); ok {
				mm.CategoryProductsModel = cpm
			} else {
				log.Printf("unexpected type from CategoryProductsModel.Update: %T", updatedViewModel)
				mm.ErrMsg = "internal error: unexpected model type"
				break
			}
			cmds = append(cmds, cmd)
			mm.CurrentView = mm.CategoryProductsModel
			if mm.Width > 0 && mm.Height > 0 {
				resized, sizeCmd := mm.CurrentView.Update(tea.WindowSizeMsg{Width: mm.Width, Height: mm.Height})
				mm.CurrentView = resized
				cmds = append(cmds, sizeCmd)
				if v, ok := mm.CurrentView.(categoryproducts.Model); ok {
					mm.CategoryProductsModel = v
				}
			}
			mm.ErrMsg = ""
		}

	case commands.LoadCategoryProductsPageMsg:
		handled = true
		if msg.NavGen != mm.requestID {
			break // stale message from a previous view visit
		}
		if mm.State != CategoryProductsView {
			// Stale message: user navigated away before this cmd was delivered.
			break
		}
		if msg.CategoryID != mm.CategoryProductsModel.CategoryID() {
			break // stale message for a different category
		}
		// Debounce: store the latest target page and (re)start a 200ms timer.
		mm.debounceMode = "category"
		mm.debouncePage = msg.Page
		mm.debounceCategoryID = msg.CategoryID
		mm.debounceCategoryName = msg.CategoryName
		mm.debounceGen++
		mm.CategoryProductsModel.CurrentPage = msg.Page
		mm.CurrentView = mm.CategoryProductsModel
		cmds = append(cmds, debounceCmd(200*time.Millisecond, mm.debounceGen, "category"))

	case commands.ProductsMsg: // This is for displaying a single product
		handled = true
		if msg.NavGen != mm.requestID {
			break // stale message from a previous view visit
		}
		if mm.State != LatestView && mm.State != CategoryProductsView {
			// Stale message: user navigated away before HandleDisplayProduct cmd was delivered.
			break
		}
		if msg.Err != nil {
			mm.ErrMsg = msg.Err.Error()
		} else {
			updatedViewModel, cmd = mm.ProductModel.Update(msg)
			if pm, ok := updatedViewModel.(productview.ProductModel); ok {
				mm.ProductModel = pm
			} else {
				log.Printf("unexpected type from ProductModel.Update: %T", updatedViewModel)
				mm.ErrMsg = "internal error: unexpected model type"
				break
			}
			cmds = append(cmds, cmd)
			mm.PreviousViews = append(mm.PreviousViews, mm.CurrentView)
			mm.CurrentView = mm.ProductModel
			mm.State = ProductView
			mm.ErrMsg = ""
			if mm.Width > 0 && mm.Height > 0 {
				resized, sizeCmd := mm.CurrentView.Update(tea.WindowSizeMsg{Width: mm.Width, Height: mm.Height})
				mm.CurrentView = resized
				cmds = append(cmds, sizeCmd)
				if pv, ok := mm.CurrentView.(productview.ProductModel); ok {
					mm.ProductModel = pv
				}
			}
		}

	case debounceFireMsg:
		handled = true
		// Only process the latest generation; stale timers are dropped.
		if msg.gen != mm.debounceGen {
			break
		}
		// Guard against user having navigated away while timer was running.
		if msg.mode == "latest" && mm.State != LatestView {
			break
		}
		if msg.mode == "category" && mm.State != CategoryProductsView {
			break
		}
		mm.Loading = true
		mm.ErrMsg = ""
		mm.requestID++
		ctx, cancel := context.WithCancel(context.Background())
		mm.cancelFunc = cancel
		if msg.mode == "latest" {
			cmds = append(cmds, commands.HandleGetLatest(ctx, mm.Width, mm.Height, mm.debouncePage, mm.LatestModel.PerPage, mm.requestID))
		} else {
			cmds = append(cmds, commands.HandleGetProductsByCategory(ctx, mm.debounceCategoryID, mm.debounceCategoryName, mm.debouncePage, mm.CategoryProductsModel.PerPage, mm.requestID))
		}
	}

	if !handled {
		if mm.Loading {
			var spinCmd tea.Cmd
			mm.Spinner, spinCmd = mm.Spinner.Update(msg)
			cmds = append(cmds, spinCmd)
		} else {
			currentGen := mm.requestID
			var currentViewCmd tea.Cmd
			mm.CurrentView, currentViewCmd = mm.CurrentView.Update(msg)
			cmds = append(cmds, wrapChildCmd(currentViewCmd, currentGen))
			// Sync back to typed fields
			switch m := mm.CurrentView.(type) {
			case defaultview.DefaultModel:
				mm.DefaultModel = m
			case latest.LatestModel:
				mm.LatestModel = m
			case categories.Model:
				mm.CategoriesModel = m
			case categoryproducts.Model:
				mm.CategoryProductsModel = m
			case productview.ProductModel:
				mm.ProductModel = m
			}
		}
	}

	return mm, tea.Batch(cmds...)
}

func (mm MainModel) View() tea.View {
	// var mainViewContent string // Removed as it was declared and not used before re-assignment
	var helpContent string

	switch mm.State {
	case DefaultView:
		helpContent = "↑/↓: navigate | enter: select | q: quit"
	case LatestView, CategoryProductsView:
		helpContent = "↑/↓: navigate | enter: select | h/←: back | n: next | p: prev | q: quit"
	case CategoriesView: // Categories view does not have n/p for its own list
		helpContent = "↑/↓: navigate | enter: select | h/←: back | r: refresh | q: quit"
	case ProductView:
		helpContent = "h/←: back | q: quit"
	default:
		helpContent = "q: quit"
	}

	if mm.ErrMsg != "" {
		helpContent = "h/←: back | q: quit"
	}

	footerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		PaddingTop(1).
		Foreground(lipgloss.Color("240"))

	styledFooter := footerStyle.Width(mm.Width).Render(helpContent)

	if mm.ErrMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Width(mm.Width).Align(lipgloss.Center)
		errorMessage := errorStyle.Render("Error: " + mm.ErrMsg)
		errorViewHeight := max(mm.Height-lipgloss.Height(styledFooter), 0)
		centeredError := lipgloss.Place(mm.Width, errorViewHeight, lipgloss.Center, lipgloss.Center, errorMessage)
		v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, centeredError, styledFooter))
		v.AltScreen = true
		return v
	}

	if mm.Loading {
		loadingViewContent := lipgloss.JoinVertical(
			lipgloss.Center,
			logo,
			mm.Spinner.View(),
			"Loading...",
		)
		// Calculate height for loading view area, leave space for footer
		loadingViewHeight := max(mm.Height-lipgloss.Height(styledFooter), 0)
		centeredLoadingView := lipgloss.Place(mm.Width, loadingViewHeight, lipgloss.Center, lipgloss.Center, loadingViewContent)
		v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, centeredLoadingView, styledFooter))
		v.AltScreen = true
		return v
	}

	// If not loading and no error, show the current view with logo and footer
	currentViewRender := mm.CurrentView.View().Content

	// Join logo and current view horizontally
	logoAndCurrentView := lipgloss.JoinHorizontal(lipgloss.Top, logo, currentViewRender)

	finalLayout := lipgloss.JoinVertical(lipgloss.Left, logoAndCurrentView, styledFooter)

	v := tea.NewView(finalLayout)
	v.AltScreen = true
	return v
}

// wrapChildCmd stamps locally-generated navigation messages with the current
// view generation so stale deliveries after back-navigation are dropped.
func wrapChildCmd(cmd tea.Cmd, gen int) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		switch m := msg.(type) {
		case commands.ProductsMsg:
			m.NavGen = gen
			return m
		case commands.StartLoadingProductsForCategoryMsg:
			m.NavGen = gen
			return m
		case commands.LoadLatestPageMsg:
			m.NavGen = gen
			return m
		case commands.LoadCategoryProductsPageMsg:
			m.NavGen = gen
			return m
		case tea.BatchMsg:
			wrapped := make([]tea.Cmd, len(m))
			for i, c := range m {
				wrapped[i] = wrapChildCmd(c, gen)
			}
			return tea.BatchMsg(wrapped)
		}
		return msg
	}
}

// debounceCmd returns a tea.Cmd that fires a debounceFireMsg after d. Each call
// to debounceCmd returns a new command with a generation counter; only the latest
// generation's message is processed, effectively resetting the timer on every
// rapid keypress.
func debounceCmd(d time.Duration, gen int, mode string) tea.Cmd {
	return func() tea.Msg {
		<-time.After(d)
		return debounceFireMsg{gen: gen, mode: mode}
	}
}

func Run() {
	// Initialize the API client before the TUI starts.
	if commands.ApiClient == nil {
		commands.ApiClient = api.NewHTTPClientWithRetry(
			commands.TheHoptimistBaseURL,
			"hopcli/"+commands.Version,
		)
	}

	model := InitialModel()
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
