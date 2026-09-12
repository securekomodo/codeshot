// Package sniff guesses which preset fits a piece of text, so piped input
// gets diff, log, terminal or JSON treatment without being told.
package sniff

import (
	"encoding/json"
	"regexp"
	"strings"

	"codeshot/internal/preset"
)

var (
	diffHeader  = regexp.MustCompile(`^(diff --git |@@ -\d+(,\d+)? \+\d+)`)
	diffFile    = regexp.MustCompile(`^(\+\+\+ |--- )`)
	gitCommit   = regexp.MustCompile(`^commit [0-9a-f]{7,40}\b`)
	promptLine  = regexp.MustCompile(`^[$❯] \S`)
	httpRequest = regexp.MustCompile(`^(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS) \S+ HTTP/\d`)
	envLine     = regexp.MustCompile(`^(export )?[A-Za-z_][A-Za-z0-9_]*=`)
	treeLine    = regexp.MustCompile(`(├──|└──|\|--|` + "`--" + `)`)
	logLine     = regexp.MustCompile(`^\[?\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}|^\[?\d{2}:\d{2}:\d{2}|\b(TRACE|DEBUG|INFO|NOTICE|WARN|WARNING|ERROR|FATAL)\b`)
	testLine    = regexp.MustCompile(`(?i)^\s*(✓|✔|✗|✘|○|PASS\b|FAIL\b|ok\b|not ok\b)|\b(Tests?|Test Suites):\s+\d+`)
)

// Preset returns the preset key that best fits text, or "" if it looks like
// ordinary source code.
func Preset(text string) string {
	lines := nonEmpty(strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n"))
	if len(lines) == 0 {
		return ""
	}
	first := lines[0]
	count := func(re *regexp.Regexp) int {
		n := 0
		for _, l := range lines {
			if re.MatchString(l) {
				n++
			}
		}
		return n
	}
	switch {
	case count(diffHeader) > 0 || count(diffFile) >= 2:
		return "git-diff"
	case gitCommit.MatchString(first):
		return "git-commit"
	case httpRequest.MatchString(first):
		return "http-request"
	case count(promptLine) > 0:
		return "terminal"
	case looksJSON(text):
		return "api"
	case count(treeLine) >= 2:
		return "project-structure"
	case count(envLine) >= max(2, (len(lines)-comments(lines))*3/5) && count(envLine) > 0:
		return "env-vars"
	case count(testLine) >= 2:
		return "test-results"
	case count(logLine)*2 >= len(lines):
		return "error-log"
	}
	return ""
}

func looksJSON(text string) bool {
	t := strings.TrimSpace(text)
	if t == "" || (t[0] != '{' && t[0] != '[') {
		return false
	}
	return json.Valid([]byte(t))
}

func nonEmpty(lines []string) []string {
	var out []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

func comments(lines []string) int {
	n := 0
	for _, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "#") {
			n++
		}
	}
	return n
}

// Lookup returns the preset for a sniffed key, or the fallback when the text
// looked like source code.
func Lookup(text string, fallback preset.Preset) preset.Preset {
	if key := Preset(text); key != "" {
		if p, ok := preset.Get(key); ok {
			return p
		}
	}
	return fallback
}
