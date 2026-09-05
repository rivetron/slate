package display

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/rivetron/slate/src/models"
)

type Renderer struct {
	glamourRender *glamour.TermRenderer
	width         int
	height        int
	config        *models.Config
	style         lipgloss.Style
}

func New(config *models.Config, width, height int) (*Renderer, error) {
	glamourStyle := config.Theme.GlamourStyle
	if glamourStyle == "" {
		glamourStyle = "dark"
	}

	// * Create glamour renderer
	gr, err := glamour.NewTermRenderer(
		glamour.WithStylePath(glamourStyle),
		glamour.WithWordWrap(config.Presentation.WordWrap),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create glamour renderer: %w", err)
	}

	// * Create base style
	style := lipgloss.NewStyle().
		Width(width - (config.Presentation.Margin * 2)).
		Height(height - (config.Presentation.Margin * 2)).
		Padding(config.Presentation.Padding).
		Margin(config.Presentation.Margin)

	return &Renderer{
		glamourRender: gr,
		width:         width,
		height:        height,
		config:        config,
		style:         style,
	}, nil

}

func (r *Renderer) stripMetadataComments(content string) string {
	// * Remove <!-- @key: value --> style comments
	lines := strings.Split(content, "\n")
	filtered := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<!--") && strings.Contains(trimmed, "@") {
			continue
		}
		filtered = append(filtered, line)
	}

	return strings.Join(filtered, "\n")
}

func (r *Renderer) renderSlideNumber(current, total int) string {
	text := fmt.Sprintf("%d / %d", current+1, total)

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(r.width).
		Align(lipgloss.Right)

	return style.Render(text)
}

func (r *Renderer) renderProgressBar(current, total int) string {
	if total == 0 {
		return ""
	}

	barWidth := max(r.width-(r.config.Presentation.Margin*4), 10)

	filled := (current * barWidth) / total
	empty := barWidth - filled

	bar := strings.Repeat("━", filled) + strings.Repeat("─", empty)

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Width(r.width).
		Align(lipgloss.Center)

	return style.Render(bar)
}

func (r *Renderer) RenderSlide(slide *models.Slide) (string, error) {
	if slide.HasCache() {
		return slide.GetRenderedCache(), nil
	}

	content := slide.Content()

	// * Remove slide metadata comments
	content = r.stripMetadataComments(content)

	rendered, err := r.glamourRender.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}

	// * Apply styling
	styled := r.style.Render(rendered)

	// * Cache rendered output
	slide.SetRenderedCache(styled)

	return styled, nil
}

func (r *Renderer) RenderWithProgress(slide *models.Slide, current, total int) (string, error) {
	slideContent, err := r.RenderSlide(slide)
	if err != nil {
		return "", err
	}

	// * Add Progress bar if enabled
	if r.config.Theme.ShowProgress {
		progress := r.renderProgressBar(current, total)
		slideContent = slideContent + "\n" + progress
	}

	// * Add slide number if enabled
	if r.config.Theme.ShowSlideNum {
		slideNum := r.renderSlideNumber(current, total)
		slideContent = slideContent + "\n" + slideNum
	}

	return slideContent, nil
}

func (r *Renderer) RenderTitle(title, subtitle string) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		Width(r.width).
		Align(lipgloss.Center).
		MarginTop(r.height / 3)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(r.width).
		Align(lipgloss.Center).
		MarginTop(1)

	var content string
	content += titleStyle.Render(title)
	if subtitle != "" {
		content += "\n" + subtitleStyle.Render(subtitle)
	}

	return content
}

func (r *Renderer) RenderError(err error) string {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true).
		Width(r.width).
		Align(lipgloss.Center).
		MarginTop(r.height / 3)

	return errorStyle.Render(fmt.Sprintf("Error: %s", err.Error()))
}

func (r *Renderer) Resize(width, height int) {
	r.width = width
	r.height = height

	r.style = r.style.
		Width(width - (r.config.Presentation.Margin * 2)).
		Height(height - (r.config.Presentation.Margin * 2))
}

func (r *Renderer) ClearCache(presentation *models.Presentation) {
	for i := 0; i < presentation.SlideCount(); i++ {
		if slide, err := presentation.GetSlide(i); err == nil {
			slide.ClearCache()
		}
	}
}
