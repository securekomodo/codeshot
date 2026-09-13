// Package preset holds the seventeen "tool" presets: each fixes a render
// mode, a default title, and for the syntax-highlighted ones a language,
// plus a built-in sample snippet.
package preset

import (
	"embed"
	"fmt"
	"strings"
)

// Render modes. Prism runs the syntax highlighter; the others are the
// site's per-line colorizers, and Milestone only changes the layout.
const (
	Prism     = "prism"
	Terminal  = "terminal"
	Log       = "log"
	Git       = "git"
	Tree      = "tree"
	Env       = "env"
	HTTP      = "http"
	Diff      = "diff"
	Test      = "test"
	Metrics   = "metrics"
	Milestone = "milestone"
)

// Preset is one tool preset.
type Preset struct {
	Key      string
	Render   string
	Title    string // default window title
	Language string // Prism language for Render == Prism
	Prompt   string // default prompt for Render == Terminal
	Desc     string // one line, shown by --help and --list presets
}

// Default is the preset used when none is requested.
const Default = "code"

// All lists the presets in display order.
var All = []Preset{
	{"code", Prism, "snippet.js", "javascript", "", "a code snippet, highlighted by --lang"},
	{"terminal", Terminal, "Terminal", "", "$", "a shell session"},
	{"api", Prism, "response.json", "json", "", "a JSON response, with a status badge"},
	{"error-log", Log, "server.log", "", "", "server logs, colored by level"},
	{"db-schema", Prism, "schema.sql", "sql", "", "SQL schema definitions"},
	{"sql-query", Prism, "query.sql", "sql", "", "a SQL query"},
	{"regex", Prism, "pattern.re", "regex", "", "regular expressions"},
	{"git-commit", Git, "git log", "", "", "git log output"},
	{"project-structure", Tree, "structure", "", "", "a directory tree"},
	{"env-vars", Env, ".env", "", "", "KEY=VALUE lines"},
	{"terminal-session", Terminal, "Session", "", "$", "a longer shell session"},
	{"dev-milestone", Milestone, "milestone", "", "", "large centered text, no window"},
	{"http-request", HTTP, "request.http", "", "", "an HTTP request"},
	{"git-diff", Diff, "changes.diff", "", "", "a unified diff"},
	{"test-results", Test, "npm test", "", "", "test output, colored by pass and fail"},
	{"ascii-tree", Tree, "tree", "", "", "any box-drawing tree"},
	{"perf-metrics", Metrics, "lighthouse", "", "", "labelled scores and timings"},
}

//go:embed samples
var samples embed.FS

// Get returns the preset with the given key.
func Get(key string) (Preset, bool) {
	for _, p := range All {
		if p.Key == key {
			return p, true
		}
	}
	return Preset{}, false
}

// Keys returns the preset keys in tab order.
func Keys() []string {
	keys := make([]string, len(All))
	for i, p := range All {
		keys[i] = p.Key
	}
	return keys
}

// Sample returns the preset's built-in demo snippet.
func (p Preset) Sample() (string, error) {
	b, err := samples.ReadFile("samples/" + p.Key + ".txt")
	if err != nil {
		return "", fmt.Errorf("no sample for preset %q: %w", p.Key, err)
	}
	return string(b), nil
}

// IsTerminal reports whether the preset is one of the two terminal tabs,
// whose window title gets " — zsh" appended.
func (p Preset) IsTerminal() bool {
	return p.Key == "terminal" || p.Key == "terminal-session"
}

// SupportsLineNumbers reports whether the preset can show a line-number
// gutter (only the Prism and log renderers do).
func (p Preset) SupportsLineNumbers() bool {
	return p.Render == Prism || p.Render == Log
}

// String lists the keys, for error messages.
func String() string { return strings.Join(Keys(), ", ") }
