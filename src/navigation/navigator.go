package navigation

import (
	"fmt"

	"github.com/rivetron/slate/src/models"
)

type Navigator struct {
	presentation *models.Presentation
	currentIndex int
	history      []int
	maxHistory   int
}

func New(presentation *models.Presentation) *Navigator {
	return &Navigator{
		presentation: presentation,
		currentIndex: 0,
		history:      make([]int, 0),
		maxHistory:   100,
	}
}

func (n *Navigator) recordHistory() {
	// * Add current position to history
	n.history = append(n.history, n.currentIndex)

	// * Trim history if it exceeds max size
	if len(n.history) > n.maxHistory {
		n.history = n.history[1:]
	}
}

func (n *Navigator) Next() bool {
	if n.currentIndex < n.presentation.SlideCount()-1 {
		n.recordHistory()
		n.currentIndex++
		return true
	}
	return false
}

func (n *Navigator) Previous() bool {
	if n.currentIndex > 0 {
		n.recordHistory()
		n.currentIndex--
		return true
	}
	return false
}

func (n *Navigator) First() bool {
	if n.currentIndex != 0 {
		n.recordHistory()
		n.currentIndex = 0
		return true
	}
	return false
}

func (n *Navigator) Last() bool {
	lastIndex := n.presentation.SlideCount() - 1
	if n.currentIndex != lastIndex {
		n.recordHistory()
		n.currentIndex = lastIndex
		return true
	}
	return false
}

func (n *Navigator) GoTo(index int) error {
	if index < 0 || index >= n.presentation.SlideCount() {
		return fmt.Errorf("slide index %d out of bounds (0-%d)", index, n.presentation.SlideCount()-1)
	}

	if n.currentIndex != index {
		n.recordHistory()
		n.currentIndex = index
	}

	return nil
}

func (n *Navigator) GoToSlideNumber(slideNum int) error {
	return n.GoTo(slideNum - 1)
}

func (n *Navigator) Back() bool {
	if len(n.history) > 0 {
		// * Pop from history
		lastIndex := n.history[len(n.history)-1]
		n.history = n.history[:len(n.history)-1]

		n.currentIndex = lastIndex
		return true
	}
	return false
}

func (n *Navigator) CurrentIndex() int {
	return n.currentIndex
}

func (n *Navigator) CurrentSlideNumber() int {
	return n.currentIndex + 1
}

func (n *Navigator) CurrentSlide() (*models.Slide, error) {
	return n.presentation.GetSlide(n.currentIndex)
}

func (n *Navigator) TotalSlides() int {
	return n.presentation.SlideCount()
}

func (n *Navigator) HasNext() bool {
	return n.currentIndex < n.presentation.SlideCount()-1
}

func (n *Navigator) HasPrevious() bool {
	return n.currentIndex > 0
}

func (n *Navigator) IsFirst() bool {
	return n.currentIndex == 0
}

func (n *Navigator) IsLast() bool {
	return n.currentIndex == n.presentation.SlideCount()-1
}

func (n *Navigator) Progress() float64 {
	if n.presentation.SlideCount() == 0 {
		return 0.0
	}
	return float64(n.currentIndex+1) / float64(n.presentation.SlideCount())
}

func (n *Navigator) ProgressText() string {
	return fmt.Sprintf("%d/%d", n.CurrentSlideNumber(), n.TotalSlides())
}

func (n *Navigator) ClearHistory() {
	n.history = make([]int, 0)
}

func (n *Navigator) HistorySize() int {
	return len(n.history)
}

func (n *Navigator) Reset() {
	n.currentIndex = 0
	n.ClearHistory()
}

func (n *Navigator) GetSlideAt(index int) (*models.Slide, error) {
	return n.presentation.GetSlide(index)
}

func (n *Navigator) CanNavigate(index int) bool {
	return index >= 0 && index < n.presentation.SlideCount()
}

func (n *Navigator) JumpForward(count int) bool {
	targetIndex := n.currentIndex + count
	if targetIndex >= n.presentation.SlideCount() {
		return false
	}

	n.recordHistory()
	n.currentIndex = targetIndex
	return true
}

func (n *Navigator) JumpBackward(count int) bool {
	targetIndex := n.currentIndex - count
	if targetIndex < 0 {
		return false
	}

	n.recordHistory()
	n.currentIndex = targetIndex
	return true
}
