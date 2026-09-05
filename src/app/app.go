package app

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rivetron/slate/src/config"
	"github.com/rivetron/slate/src/data"
	"github.com/rivetron/slate/src/display"
	"github.com/rivetron/slate/src/models"
	"github.com/rivetron/slate/src/navigation"
	"github.com/rivetron/slate/src/theme"
)

// * ViewMode represents different view modes
type ViewMode int

const (
	ViewPresentation ViewMode = iota
	ViewHelp
)

// * BubbleTea model for App
type App struct {
	config *models.Config

	presentation *models.Presentation
	viewMode     ViewMode

	theme     *theme.Manager
	renderer  *display.Renderer
	navigator *navigation.Navigator

	width  int
	height int

	err   error
	ready bool
}

func New(filePath string) (*App, error) {
	configLoader := config.New()
	cfg, err := configLoader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// ? Validate file
	if err := data.ValidateFile(filePath); err != nil {
		return nil, err
	}

	// * Parse Presentation
	p := data.New(filePath)
	presentation, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse presentation: %w", err)
	}

	// ? Validate presentation
	if err := presentation.Validate(); err != nil {
		return nil, fmt.Errorf("invalid presentation: %w", err)
	}

	presentation.Config = cfg

	// * Create navigator
	nav := navigation.New(presentation)

	// * Create theme manager
	themeManager := theme.NewManager(&cfg.Theme)

	return &App{
		config:       cfg,
		viewMode:     ViewPresentation,
		theme:        themeManager,
		navigator:    nav,
		presentation: presentation,
	}, nil
}

func (a *App) containsKey(keys []string, key string) bool {
	return slices.Contains(keys, key)
}

func (a *App) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle help view
	if a.viewMode == ViewHelp {
		if key == "?" || key == "q" || key == "esc" {
			a.viewMode = ViewPresentation
		}
		return a, nil
	}

	// Check quit keys
	if a.containsKey(a.config.Keybindings.Quit, key) {
		return a, tea.Quit
	}

	// Handle help
	if key == "?" {
		a.viewMode = ViewHelp
		return a, nil
	}

	// Navigation keys
	if a.containsKey(a.config.Keybindings.Next, key) {
		// If on last slide and pressing next, exit the presentation
		if a.navigator.IsLast() {
			return a, tea.Quit
		}
		a.navigator.Next()
	} else if a.containsKey(a.config.Keybindings.Previous, key) {
		a.navigator.Previous()
	} else if a.containsKey(a.config.Keybindings.First, key) {
		a.navigator.First()
	} else if a.containsKey(a.config.Keybindings.Last, key) {
		a.navigator.Last()
	} else if key == "b" {
		a.navigator.Back()
	}

	return a, nil
}

func (a *App) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Padding(2).
		Align(lipgloss.Left)

	var help strings.Builder

	help.WriteString(a.theme.TitleStyle().Render("Keyboard Shortcuts"))
	help.WriteString("\n\n")

	// * Navigation
	help.WriteString(a.theme.SubtitleStyle().Render("Navigation:"))
	help.WriteString("\n")
	fmt.Fprintf(&help, "  Next slide:     %s\n", strings.Join(a.config.Keybindings.Next, ", "))
	fmt.Fprintf(&help, "  Previous slide: %s\n", strings.Join(a.config.Keybindings.Previous, ", "))
	fmt.Fprintf(&help, "  First slide:    %s\n", strings.Join(a.config.Keybindings.First, ", "))
	fmt.Fprintf(&help, "  Last slide:     %s\n", strings.Join(a.config.Keybindings.Last, ", "))
	help.WriteString("  Go back:        b\n")
	help.WriteString("\n")

	// * Other
	help.WriteString(a.theme.SubtitleStyle().Render("Other:"))
	help.WriteString("\n")
	help.WriteString("  Show help:      ?\n")
	fmt.Fprintf(&help, "  Quit:           %s\n", strings.Join(a.config.Keybindings.Quit, ", "))
	help.WriteString("\n\n")

	help.WriteString(a.theme.HelpStyle().Render("Press ? or ESC to return to presentation"))

	return helpStyle.Render(help.String())
}

func (a *App) renderCommandFooter() string {
	var commands []string

	// ? Show different commands based on position
	isFirst := a.navigator.IsFirst()
	isLast := a.navigator.IsLast()

	if !isFirst {
		commands = append(commands, "← Prev")
	}

	if !isLast {
		commands = append(commands, "→ Next")
	} else {
		// ? On last slide, emphasize the quit command
		commands = append(commands, "🏁 End")
	}

	commands = append(commands, "? Help")
	commands = append(commands, "Q Quit")

	// * Join commands with separator
	commandText := strings.Join(commands, "  •  ")

	// * Style the footer
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(a.width).
		Align(lipgloss.Center).
		Faint(true)

	// * Add special message on last slide
	if isLast {
		endMessage := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			Render("  [Press Q to exit]")
		commandText += endMessage
	}

	return footerStyle.Render(commandText)
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return a.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		// Create or update renderer
		if a.renderer == nil {
			r, err := display.New(a.config, a.width, a.height)
			if err != nil {
				a.err = err
				return a, tea.Quit
			}
			a.renderer = r
		} else {
			a.renderer.Resize(a.width, a.height)
			a.renderer.ClearCache(a.presentation)
		}

		a.ready = true
		return a, nil

	case error:
		a.err = msg
		return a, tea.Quit
	}

	return a, nil
}

func (a *App) View() string {
	if !a.ready {
		return "Loading..."
	}

	if a.err != nil {
		return a.renderer.RenderError(a.err)
	}

	// Show help view
	if a.viewMode == ViewHelp {
		return a.renderHelp()
	}

	// Get current slide
	slide, err := a.navigator.CurrentSlide()
	if err != nil {
		return a.renderer.RenderError(err)
	}

	// Render slide with progress
	rendered, err := a.renderer.RenderWithProgress(
		slide,
		a.navigator.CurrentIndex(),
		a.navigator.TotalSlides(),
	)
	if err != nil {
		return a.renderer.RenderError(err)
	}

	// Add command footer
	footer := a.renderCommandFooter()

	return rendered + "\n" + footer
}

func Run(filepath string) error {
	app, err := New(filepath)
	if err != nil {
		return err
	}

	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}
