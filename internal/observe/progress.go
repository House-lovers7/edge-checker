package observe

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// Progress handles terminal progress display during a test run.
type Progress struct {
	collector    *Collector
	totalSeconds int
	writer       io.Writer
	verbose      bool
	lastLine     string
}

// NewProgress creates a progress display.
func NewProgress(collector *Collector, totalSeconds int, writer io.Writer, verbose bool) *Progress {
	return &Progress{
		collector:    collector,
		totalSeconds: totalSeconds,
		writer:       writer,
		verbose:      verbose,
	}
}

// Update refreshes the progress display.
func (p *Progress) Update() {
	m := p.collector.Snapshot()
	elapsed := time.Since(m.StartTime).Seconds()
	if elapsed < 0 {
		elapsed = 0
	}

	elapsedInt := int(elapsed)
	if elapsedInt > p.totalSeconds {
		elapsedInt = p.totalSeconds
	}

	// Build status codes summary
	statusParts := p.formatStatusCounts(m.StatusCounts)

	// Progress bar
	barWidth := 20
	filled := 0
	if p.totalSeconds > 0 {
		filled = barWidth * elapsedInt / p.totalSeconds
	}
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled)

	line := fmt.Sprintf("\r[%s] %d/%ds | %d reqs | %s | p95: %dms",
		bar,
		elapsedInt, p.totalSeconds,
		m.TotalRequests,
		statusParts,
		m.P95Latency.Milliseconds(),
	)

	if p.verbose {
		fmt.Fprintln(p.writer, line)
	} else {
		fmt.Fprint(p.writer, line)
		p.lastLine = line
	}
}

// Finish prints a final newline after the progress bar.
func (p *Progress) Finish() {
	if !p.verbose && p.lastLine != "" {
		fmt.Fprintln(p.writer)
	}
}

func (p *Progress) formatStatusCounts(counts map[int]int) string {
	if len(counts) == 0 {
		return "no responses"
	}

	// Sort status codes for consistent display
	codes := make([]int, 0, len(counts))
	for code := range counts {
		codes = append(codes, code)
	}
	sort.Ints(codes)

	parts := make([]string, 0, len(codes))
	for _, code := range codes {
		parts = append(parts, fmt.Sprintf("%d:%d", code, counts[code]))
	}
	return strings.Join(parts, " ")
}
