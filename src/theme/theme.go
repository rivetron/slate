package theme

import (
	"os"
	"os/exec"
	"strings"

	"github.com/rivetron/slate/src/models"
	"github.com/charmbracelet/lipgloss"
)

// * Theme Modes
const (
	ModeAuto  = "auto"
	ModeDark  = "dark"
	ModeLight = "light"
)

// * Glamour Styles
const (
	GlamourPink    = "pink"
	GlamourDark    = "dark"
	GlamourLight   = "light"
	GlamourDracula = "dracula"
)

type ColorScheme struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color

	Background lipgloss.Color
	Foreground lipgloss.Color

	Accent lipgloss.Color
	Muted  lipgloss.Color

	Error   lipgloss.Color
	Success lipgloss.Color
	Warning lipgloss.Color

	Border lipgloss.Color

	ProgressBar lipgloss.Color
}

type Manager struct {
	config      *models.ThemeConfig
	isDark      bool
	colorScheme ColorScheme
}

func NewManager(config *models.ThemeConfig) *Manager {
	manager := &Manager{
		config: config,
	}

	// * Detect if we should use dark mode
	manager.isDark = manager.detectDarkMode()

	// * Set up color scheme
	manager.colorScheme = manager.createColorScheme()

	// * Set glamour style if not specified
	if manager.config.GlamourStyle == "" {
		if manager.isDark {
			manager.config.GlamourStyle = GlamourDark
		} else {
			manager.config.GlamourStyle = GlamourLight
		}
	}

	return manager
}

func (m *Manager) detectMacOSTheme() *bool {
	cmd := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle")
	output, err := cmd.Output()
	if err != nil {
		result := false
		return &result
	}

	isDark := strings.TrimSpace(string(output)) == "Dark"
	return &isDark
}

func (m *Manager) detectSystemTheme() bool {
	if colorTerm := os.Getenv("COLORTERM"); colorTerm != "" {
		return true
	}

	if isDark := m.detectMacOSTheme(); isDark != nil {
		return *isDark
	}

	return true
}

func (m *Manager) detectDarkMode() bool {
	switch m.config.Mode {
	case ModeDark:
		return true
	case ModeLight:
		return false
	case ModeAuto:
		return m.detectSystemTheme()
	default:
		return m.detectSystemTheme()
	}
}

func (m *Manager) createColorScheme() ColorScheme {
	if m.isDark {
		return ColorScheme{
			Primary:     lipgloss.Color("63"),  // Purple
			Secondary:   lipgloss.Color("99"),  // Light purple
			Background:  lipgloss.Color("235"), // Dark gray
			Foreground:  lipgloss.Color("255"), // White
			Accent:      lipgloss.Color("212"), // Pink
			Muted:       lipgloss.Color("241"), // Gray
			Error:       lipgloss.Color("196"), // Red
			Success:     lipgloss.Color("46"),  // Green
			Warning:     lipgloss.Color("214"), // Orange
			Border:      lipgloss.Color("240"), // Dark border
			ProgressBar: lipgloss.Color("63"),  // Purple
		}
	}

	return ColorScheme{
		Primary:     lipgloss.Color("63"),  // Purple
		Secondary:   lipgloss.Color("99"),  // Light purple
		Background:  lipgloss.Color("255"), // White
		Foreground:  lipgloss.Color("235"), // Dark gray
		Accent:      lipgloss.Color("212"), // Pink
		Muted:       lipgloss.Color("246"), // Light gray
		Error:       lipgloss.Color("160"), // Dark red
		Success:     lipgloss.Color("28"),  // Dark green
		Warning:     lipgloss.Color("166"), // Dark orange
		Border:      lipgloss.Color("246"), // Light border
		ProgressBar: lipgloss.Color("63"),  // Purple
	}
}

func (m *Manager) IsDark() bool {
	return m.isDark
}

func (m *Manager) GetColorScheme() ColorScheme {
	return m.colorScheme
}

func (m *Manager) GetGlamourStyle() string {
	return m.config.GlamourStyle
}

func (m *Manager) SetGlamourStyle(style string) {
	m.config.GlamourStyle = style
}

func (m *Manager) ToggleMode() {
	switch m.config.Mode {
	case ModeAuto:
		// ? If auto, switch to explicit mode (opposite of current)
		if m.isDark {
			m.config.Mode = ModeLight
		} else {
			m.config.Mode = ModeDark
		}
	case ModeDark:
		m.config.Mode = ModeLight
	case ModeLight:
		m.config.Mode = ModeDark
	default:
		// * Fallback to dark mode if unknown mode
		m.config.Mode = ModeDark
	}

	// ? Re-detect and update
	m.isDark = m.detectDarkMode()
	m.colorScheme = m.createColorScheme()

	// ? Update glamour style based on detected theme
	if m.isDark {
		m.config.GlamourStyle = GlamourDark
	} else {
		m.config.GlamourStyle = GlamourLight
	}
}

func (m *Manager) TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.colorScheme.Primary)
}

func (m *Manager) SubtitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(m.colorScheme.Muted)
}

func (m *Manager) SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.colorScheme.Success)
}

func (m *Manager) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(m.colorScheme.Error)
}

func (m *Manager) BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(m.colorScheme.Border)
}

func (m *Manager) HelpStyle() lipgloss.Style {
	return lipgloss.NewStyle().Italic(true).Foreground(m.colorScheme.Muted)
}
