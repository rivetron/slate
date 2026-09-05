package data

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rivetron/slate/src/models"
	"gopkg.in/yaml.v3"
)

const slideSeparator = "---"

var (
	// Match YAML FrontMatter at the start of the file
	frontMatterRegex = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)
	// Match slide-specific metadata comments
	slideMetadataRegex = regexp.MustCompile(`<!--\s*@(\w+):\s*(.+?)\s*-->`)
)

type Parser struct {
	filePath string
}

func New(filePath string) *Parser {
	return &Parser{
		filePath: filePath,
	}
}

func (p *Parser) extractFrontMatter(content string) (string, map[string]string) {
	matches := frontMatterRegex.FindStringSubmatch(content)
	if len(matches) < 2 {
		return content, make(map[string]string)
	}

	// * Parse YAML FrontMatter
	metadata := make(map[string]any)
	if err := yaml.Unmarshal([]byte(matches[1]), &metadata); err != nil {
		return content, make(map[string]string)
	}

	// * Convert string to map
	result := make(map[string]string)
	for k, v := range metadata {
		result[k] = fmt.Sprintf("%v", v)
	}

	// * Remove FrontMatter from content
	remainingContent := frontMatterRegex.ReplaceAllString(content, "")
	return remainingContent, result
}

func (p *Parser) splitIntoSlides(content string) []string {
	// * Split by horizontal rule (---)
	parts := strings.Split(content, "\n"+slideSeparator+"\n")

	slides := make([]string, 0)
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			slides = append(slides, trimmed)
		}
	}

	return slides
}

func (p *Parser) extractSlideMetadata(content string) models.SlideMetadata {
	metadata := models.SlideMetadata{}

	matches := slideMetadataRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		key := strings.ToLower(match[1])
		value := strings.ToLower(match[2])

		switch key {
		case "notes":
			metadata.Notes = value
		case "transition":
			metadata.Transition = value
		case "background":
			metadata.Background = value
		}
	}

	return metadata
}

func (p *Parser) Parse() (*models.Presentation, error) {
	content, err := os.ReadFile(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	presentation := models.NewPresentation(p.filePath)

	// * Extract FrontMatter metadata
	contentStr := string(content)
	contentStr, metadata := p.extractFrontMatter(contentStr)
	presentation.SetMetadata(metadata)

	// * Split content into slides
	slides := p.splitIntoSlides(contentStr)

	// * Parse each slide
	for i, slideContent := range slides {
		slide := models.NewSlide(i, slideContent)

		// * Extract slide-specific metadata
		slide.Metadata = p.extractSlideMetadata(slideContent)

		presentation.AddSlide(slide)
	}

	return presentation, nil
}

func ValidateFile(filePath string) error {
	// ? Check file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	// ? Check it's a file (not a directory)
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", filePath)
	}

	// ? Check file extension (should be .md)
	if !strings.HasSuffix(strings.ToLower(filePath), ".md") {
		return fmt.Errorf("file must have .md extension: %s", filePath)
	}

	return nil
}

func ParseFromString(content string, filePath string) (*models.Presentation, error) {
	parser := &Parser{filePath: filePath}

	presentation := models.NewPresentation(filePath)

	// * Extract FrontMatter metadata
	contentStr, metadata := parser.extractFrontMatter(content)
	presentation.SetMetadata(metadata)

	// * Split content into slides
	slides := parser.splitIntoSlides(contentStr)

	// * Parse each slide
	for i, slideContent := range slides {
		slide := models.NewSlide(i, slideContent)
		slide.Metadata = parser.extractSlideMetadata(slideContent)
		presentation.AddSlide(slide)
	}

	return presentation, nil
}

func CountSlides(filePath string) (int, error) {
	// Clean path and verify it's a regular file
	cleanPath := filepath.Clean(filePath)

	info, err := os.Stat(cleanPath)
	if err != nil {
		return 0, fmt.Errorf("cannot access file: %w", err)
	}

	if !info.Mode().IsRegular() {
		return 0, errors.New("path must be a regular file")
	}
	// #nosec G304 -- path is from user CLI argument and validated as regular file
	file, err := os.Open(cleanPath)
	if err != nil {
		return 0, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close file: %v\n", closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)
	// * At least one slide
	count := 1

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == slideSeparator {
			count++
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return count, nil
}
