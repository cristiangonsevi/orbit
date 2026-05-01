package ui

import (
	"fmt"
	"strings"
	"time"
)

// Spinner frames for animation
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner represents an animated spinner
type Spinner struct {
	message string
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// NewSpinner creates a new spinner with the given message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
}

// Start begins the spinner animation in a goroutine
func (s *Spinner) Start() {
	go func() {
		frame := 0
		for {
			select {
			case <-s.stopCh:
				// Erase the line and show completion
				fmt.Printf("\r\033[K")
				close(s.doneCh)
				return
			default:
				// Print spinner with message
				fmt.Printf("\r%s %s ", ColorCyan(spinnerFrames[frame]), s.message)
				frame = (frame + 1) % len(spinnerFrames)
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

// Stop stops the spinner and optionally shows a success/failure message
func (s *Spinner) Stop() {
	close(s.stopCh)
	<-s.doneCh
}

// StopWithSuccess shows a success checkmark after stopping
func (s *Spinner) StopWithSuccess(message string) {
	s.Stop()
	fmt.Printf("\r%s %s\n", ColorGreen("✓"), message)
}

// StopWithError shows an error X after stopping
func (s *Spinner) StopWithError(message string) {
	s.Stop()
	fmt.Printf("\r%s %s\n", ColorRed("✗"), message)
}

// Color codes for terminal output
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
	colorBold    = "\033[1m"
	colorDim     = "\033[2m"
)

// ColorRed returns text wrapped in red ANSI codes
func ColorRed(s string) string {
	return colorRed + s + colorReset
}

// ColorGreen returns text wrapped in green ANSI codes
func ColorGreen(s string) string {
	return colorGreen + s + colorReset
}

// ColorYellow returns text wrapped in yellow ANSI codes
func ColorYellow(s string) string {
	return colorYellow + s + colorReset
}

// ColorBlue returns text wrapped in blue ANSI codes
func ColorBlue(s string) string {
	return colorBlue + s + colorReset
}

// ColorMagenta returns text wrapped in magenta ANSI codes
func ColorMagenta(s string) string {
	return colorMagenta + s + colorReset
}

// ColorCyan returns text wrapped in cyan ANSI codes
func ColorCyan(s string) string {
	return colorCyan + s + colorReset
}

// ColorBold returns text wrapped in bold ANSI codes
func ColorBold(s string) string {
	return colorBold + s + colorReset
}

// ColorDim returns text wrapped in dim ANSI codes
func ColorDim(s string) string {
	return colorDim + s + colorReset
}

// ProgressBar displays a progress bar
type ProgressBar struct {
	width    int
	finished int
	total    int
	prefix   string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(prefix string, total int) *ProgressBar {
	return &ProgressBar{
		width:    40,
		total:    total,
		prefix:   prefix,
		finished: 0,
	}
}

// SetTotal updates the total count
func (p *ProgressBar) SetTotal(total int) {
	p.total = total
}

// Increment advances the progress by one
func (p *ProgressBar) Increment() {
	p.finished++
	p.Draw()
}

// SetProgress sets the current progress
func (p *ProgressBar) SetProgress(current int) {
	p.finished = current
	p.Draw()
}

// Draw renders the progress bar
func (p *ProgressBar) Draw() {
	if p.total == 0 {
		return
	}

	percent := float64(p.finished) / float64(p.total)
	filled := int(float64(p.width) * percent)
	empty := p.width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	percentStr := fmt.Sprintf("%.0f%%", percent*100)

	fmt.Printf("\r%s [%s] %s (%d/%d)",
		p.prefix, bar, percentStr, p.finished, p.total)
}

// Clear erases the progress bar line
func (p *ProgressBar) Clear() {
	fmt.Printf("\r\033[K")
}

// Done prints completion message
func (p *ProgressBar) Done(message string) {
	p.Clear()
	fmt.Printf("%s %s\n", ColorGreen("✓"), message)
}

// Header prints a styled header
func Header(title string) {
	fmt.Println()
	fmt.Println(ColorBold(ColorCyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")))
	fmt.Printf("  %s\n", ColorBold(ColorCyan(title)))
	fmt.Println(ColorBold(ColorCyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")))
	fmt.Println()
}

// SubHeader prints a styled sub-header
func SubHeader(title string) {
	fmt.Println()
	fmt.Printf("%s %s\n", ColorBold("▶"), ColorBold(ColorBlue(title)))
}

// Success prints a success message
func Success(message string) {
	fmt.Printf("%s %s\n", ColorGreen("✓"), message)
}

// Info prints an info message
func Info(message string) {
	fmt.Printf("%s %s\n", ColorBlue("●"), message)
}

// Warning prints a warning message
func Warning(message string) {
	fmt.Printf("%s %s\n", ColorYellow("⚠"), message)
}

// Error prints an error message
func Error(message string) {
	fmt.Printf("%s %s\n", ColorRed("✗"), ColorRed(message))
}

// Step prints a numbered step
func Step(num int, total int, message string) {
	fmt.Printf("%s %d/%d %s\n", ColorMagenta("▸"), num, total, ColorBold(message))
}

// Separator prints a visual separator
func Separator() {
	fmt.Println(ColorDim("────────────────────────────────────────────────────────"))
}
