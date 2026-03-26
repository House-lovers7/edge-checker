package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/House-lovers7/edge-checker/internal/judge"
)

const markdownTemplate = `# WAF Defense Verification Report

## Scenario
- **Name:** {{ .ScenarioName }}
{{- if .Description }}
- **Description:** {{ .Description }}
{{- end }}
- **Started:** {{ .StartedAt.Format "2006-01-02 15:04:05 MST" }}
- **Ended:** {{ .EndedAt.Format "2006-01-02 15:04:05 MST" }}
- **Duration:** {{ .Duration }}
{{- if .Interrupted }}
- **Status:** Interrupted
{{- end }}

## Target
- **URL:** {{ .Target.Method }} {{ .Target.BaseURL }}{{ .Target.Path }}
- **Profile:** {{ .Target.Profile }}

## Execution
- **Mode:** {{ .Execution.Mode }}
- **Duration:** {{ .Execution.Duration }}
- **Concurrency:** {{ .Execution.Concurrency }}
- **RPS:** {{ .Execution.RPS }}
- **Environment:** {{ .Execution.Environment }}

## Summary
| Metric | Value |
|--------|-------|
| Total requests | {{ .Summary.TotalRequests }} |
| Success (2xx) | {{ .Summary.SuccessCount }} |
| Errors | {{ .Summary.ErrorCount }} |
| Avg latency | {{ printf "%.1f" .Summary.AvgLatencyMs }} ms |
| p50 latency | {{ printf "%.1f" .Summary.P50LatencyMs }} ms |
| p95 latency | {{ printf "%.1f" .Summary.P95LatencyMs }} ms |
| p99 latency | {{ printf "%.1f" .Summary.P99LatencyMs }} ms |

## Status Code Distribution
| Status | Count | Percentage |
|--------|-------|------------|
{{ range $code, $count := .StatusCodesSorted -}}
| {{ $code.Code }} | {{ $code.Count }} | {{ $code.Pct }}% |
{{ end -}}

{{- if .Verdict }}

## Verdict: {{ .VerdictIcon }} {{ .Verdict.Overall }}

| Rule | Status | Expected | Actual |
|------|--------|----------|--------|
{{ range .Verdict.Rules -}}
| {{ .Name }} | {{ verdictIcon .Status }} {{ .Status }} | {{ .Expected }} | {{ .Actual }} |
{{ end -}}
{{- end }}

## Timeline (per second)
| Second | Requests | {{ range $.StatusCodesHeader }}{{ . }} | {{ end }}Errors | Avg Latency |
|--------|----------|{{ range $.StatusCodesHeader }}--------|{{ end }}--------|-------------|
{{ range .Timeline -}}
| {{ .Second }}s | {{ .RequestCount }} | {{ range $.StatusCodesForBucket . }}{{ . }} | {{ end }}{{ .ErrorCount }} | {{ printf "%.0f" .AvgLatency }}ms |
{{ end -}}
`

// StatusCodeEntry is used for sorted status code display.
type StatusCodeEntry struct {
	Code  int
	Count int
	Pct   int
}

// MarkdownData extends Result with template helper data.
type MarkdownData struct {
	*Result
	StatusCodesSorted []StatusCodeEntry
}

// VerdictIcon returns an icon for the overall verdict.
func (d *MarkdownData) VerdictIcon() string {
	if d.Verdict != nil && d.Verdict.Overall == judge.StatusPass {
		return "PASS"
	}
	return "FAIL"
}

// StatusCodesHeader returns sorted unique status codes for the table header.
func (d *MarkdownData) StatusCodesHeader() []string {
	codes := d.sortedCodes()
	headers := make([]string, len(codes))
	for i, c := range codes {
		headers[i] = fmt.Sprintf("%d", c)
	}
	return headers
}

// StatusCodesForBucket returns counts for each status code in a timeline bucket.
func (d *MarkdownData) StatusCodesForBucket(bucket interface{}) []string {
	// Type assert from the template
	type bucketLike interface {
		GetStatusCounts() map[int]int
	}

	codes := d.sortedCodes()
	result := make([]string, len(codes))

	// Get status counts from bucket via JSON roundtrip (template-safe)
	data, _ := json.Marshal(bucket)
	var m map[string]interface{}
	json.Unmarshal(data, &m)

	statusCounts := make(map[int]int)
	if sc, ok := m["status_counts"].(map[string]interface{}); ok {
		for k, v := range sc {
			var code int
			fmt.Sscanf(k, "%d", &code)
			if n, ok := v.(float64); ok {
				statusCounts[code] = int(n)
			}
		}
	}

	for i, code := range codes {
		result[i] = fmt.Sprintf("%d", statusCounts[code])
	}
	return result
}

func (d *MarkdownData) sortedCodes() []int {
	codeSet := make(map[int]bool)
	for code := range d.Summary.StatusCounts {
		codeSet[code] = true
	}
	codes := make([]int, 0, len(codeSet))
	for c := range codeSet {
		codes = append(codes, c)
	}
	sort.Ints(codes)
	return codes
}

// WriteMarkdown generates a Markdown report from a Result.
func WriteMarkdown(result *Result, pathPattern string) (string, error) {
	path := expandPattern(pathPattern, result)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory %q: %w", dir, err)
	}

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create report file %q: %w", path, err)
	}
	defer f.Close()

	if err := renderMarkdown(f, result); err != nil {
		return "", err
	}

	return path, nil
}

// RenderMarkdownToString renders the markdown to a string (for show command).
func RenderMarkdownToString(result *Result) (string, error) {
	var sb strings.Builder
	if err := renderMarkdown(&sb, result); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func renderMarkdown(w io.Writer, result *Result) error {
	funcMap := template.FuncMap{
		"verdictIcon": func(s judge.VerdictStatus) string {
			switch s {
			case judge.StatusPass:
				return "✓"
			case judge.StatusFail:
				return "✗"
			default:
				return "-"
			}
		},
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(markdownTemplate)
	if err != nil {
		return fmt.Errorf("template parse error: %w", err)
	}

	data := &MarkdownData{
		Result: result,
	}

	// Build sorted status code entries
	total := result.Summary.TotalRequests
	codes := make([]int, 0, len(result.Summary.StatusCounts))
	for c := range result.Summary.StatusCounts {
		codes = append(codes, c)
	}
	sort.Ints(codes)
	for _, c := range codes {
		count := result.Summary.StatusCounts[c]
		pct := 0
		if total > 0 {
			pct = count * 100 / total
		}
		data.StatusCodesSorted = append(data.StatusCodesSorted, StatusCodeEntry{
			Code:  c,
			Count: count,
			Pct:   pct,
		})
	}

	return tmpl.Execute(w, data)
}
