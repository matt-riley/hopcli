package hopt

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
}

func InitialModel() MainModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

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

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mm.Width = msg.Width
		mm.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return mm, tea.Quit
		case "h", "left":
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
				mm.ErrMsg = ""
			}
		}
	case defaultview.StartLoadingLatestMsg:
		mm.Loading = true
		mm.ErrMsg = ""
		cmds = append(cmds, commands.HandleGetLatest(mm.Width, mm.Height))

	case defaultview.StartLoadingCategoriesMsg:
		mm.Loading = true
		mm.ErrMsg = ""
		cmds = append(cmds, commands.HandleGetCategories(mm.Width, mm.Height))

	case commands.LatestResponseMsg:
		mm.Loading = false
		if msg.Err != nil {
			mm.ErrMsg = msg.Err.Error()
		} else {
			updatedViewModel, cmd = mm.LatestModel.Update(msg)
			mm.LatestModel = updatedViewModel.(latest.LatestModel)
			cmds = append(cmds, cmd)
			mm.PreviousViews = append(mm.PreviousViews, mm.CurrentView)
			mm.CurrentView = mm.LatestModel
			mm.State = LatestView
			mm.ErrMsg = ""
		}

	case commands.CategoriesResponseMsg:
		mm.Loading = false
		if msg.Err != nil {
			mm.ErrMsg = msg.Err.Error()
		} else {
			updatedViewModel, cmd = mm.CategoriesModel.Update(msg)
			mm.CategoriesModel = updatedViewModel.(categories.Model)
			cmds = append(cmds, cmd)
			mm.PreviousViews = append(mm.PreviousViews, mm.CurrentView)
			mm.CurrentView = mm.CategoriesModel
			mm.State = CategoriesView
			mm.ErrMsg = ""
		}

	case commands.StartLoadingProductsForCategoryMsg:
		mm.Loading = true
		mm.ErrMsg = ""
		cmds = append(cmds, commands.HandleGetProductsByCategory(msg.Width, msg.Height, msg.CategoryID, msg.CategoryName, msg.APIEndpoint))

	case commands.ProductsForCategoryResponseMsg:
		mm.Loading = false
		if msg.Err != nil {
			mm.ErrMsg = msg.Err.Error()
		} else {
			// CategoryProductsModel is typically created new each time, so no zero check needed here.
			// It's initialized directly.
			mm.CategoryProductsModel = categoryproducts.NewModel(msg.CategoryName, msg.CategoryID)
			updatedViewModel, cmd = mm.CategoryProductsModel.Update(msg)
			mm.CategoryProductsModel = updatedViewModel.(categoryproducts.Model)
			cmds = append(cmds, cmd)
			mm.PreviousViews = append(mm.PreviousViews, mm.CurrentView)
			mm.CurrentView = mm.CategoryProductsModel
			mm.State = CategoryProductsView
			mm.ErrMsg = ""
		}

	case commands.ProductsMsg: // This is for displaying a single product
		mm.Loading = false
		if msg.Err != nil {
			mm.ErrMsg = msg.Err.Error()
		} else {
			updatedViewModel, cmd = mm.ProductModel.Update(msg)
			mm.ProductModel = updatedViewModel.(productview.ProductModel)
			cmds = append(cmds, cmd)
			mm.PreviousViews = append(mm.PreviousViews, mm.CurrentView)
			mm.CurrentView = mm.ProductModel
			mm.State = ProductView
			mm.ErrMsg = ""
		}
	}

	// Update current view & spinner if loading
	// Only pass non-nil messages or messages the current view specifically handles
	// This avoids passing spinner ticks or other messages to views not expecting them if not loading
	if !mm.Loading {
		var currentViewCmd tea.Cmd
		mm.CurrentView, currentViewCmd = mm.CurrentView.Update(msg)
		cmds = append(cmds, currentViewCmd)
	}

	if mm.Loading {
		var spinCmd tea.Cmd
		mm.Spinner, spinCmd = mm.Spinner.Update(msg) // Spinner needs ticks
		cmds = append(cmds, spinCmd)
	} else {
		// If not loading, ensure current view is updated even if no specific msg type matched above
		// This handles general key presses for list navigation within the current view
		if _, ok := msg.(tea.KeyMsg); ok { // Only pass key messages if not loading and not handled above
			var currentViewCmd tea.Cmd
			mm.CurrentView, currentViewCmd = mm.CurrentView.Update(msg)
			cmds = append(cmds, currentViewCmd)
		}
	}

	return mm, tea.Batch(cmds...)
}

func (mm MainModel) View() string {
	// var mainViewContent string // Removed as it was declared and not used before re-assignment
	var helpContent string

	switch mm.State {
	case DefaultView:
		helpContent = "↑/↓: navigate | enter: select | q: quit"
	case LatestView, CategoriesView, CategoryProductsView:
		helpContent = "↑/↓: navigate | enter: select | h/←: back | q: quit"
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
		errorViewHeight := mm.Height - lipgloss.Height(styledFooter)
		if errorViewHeight < 0 {errorViewHeight = 0}
		centeredError := lipgloss.Place(mm.Width, errorViewHeight, lipgloss.Center, lipgloss.Center, errorMessage)
		return lipgloss.JoinVertical(lipgloss.Left, centeredError, styledFooter)
	}

	if mm.Loading {
		loadingViewContent := lipgloss.JoinVertical(
			lipgloss.Center,
			logo,
			mm.Spinner.View(),
			"Loading...",
		)
		// Calculate height for loading view area, leave space for footer
		loadingViewHeight := mm.Height - lipgloss.Height(styledFooter)
		if loadingViewHeight < 0 {
			loadingViewHeight = 0
		}
		centeredLoadingView := lipgloss.Place(mm.Width, loadingViewHeight, lipgloss.Center, lipgloss.Center, loadingViewContent)
		return lipgloss.JoinVertical(lipgloss.Left, centeredLoadingView, styledFooter)
	}

	// If not loading and no error, show the current view with logo and footer
	currentViewRender := mm.CurrentView.View()

	// Calculate available height for the main content (logo + current view)
	mainContentHeight := mm.Height - lipgloss.Height(styledFooter)
	if mainContentHeight < 0 {
		mainContentHeight = 0
	}

	// Join logo and current view horizontally
	logoAndCurrentView := lipgloss.JoinHorizontal(lipgloss.Top, logo, currentViewRender)

	// This part is tricky without knowing the exact height of logo and currentViewRender.
	// We will assume that logoAndCurrentView might be taller than mainContentHeight and let it be clipped or scroll.
	// For a more robust solution, the heights of logo and currentViewRender would need to be managed.
	// For now, we'll join them and then join with the footer.

	finalLayout := lipgloss.JoinVertical(lipgloss.Left, logoAndCurrentView, styledFooter)

	// If the total height is still too much, we might need to Place the main content area.
	// However, JoinVertical doesn't inherently know about mm.Height to constrain itself.
	// Let's try to ensure the main interactive area (currentViewRender) is what gets space.

	// A slightly better approach for the main content area to respect height:
	// Calculate height for the current view area (mainContentHeight - logo height)
	// This is still an approximation as logo height isn't fixed based on terminal width.
	// For now, the simpler JoinVertical is used as per Option A.

	return finalLayout
}

func Run() {
	model := InitialModel()
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
