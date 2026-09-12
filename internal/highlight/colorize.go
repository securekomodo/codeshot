package highlight

import (
	"regexp"
	"strconv"
	"strings"

	"codeshot/internal/preset"
	"codeshot/internal/theme"
)

// Options tune the colorizers.
type Options struct {
	Theme       *theme.Theme
	Prompt      string // terminal modes: replaces the prompt ("" keeps the original)
	PromptColor string // terminal modes: color of the prompt ("" = teal)
	PromptStyle string // terminal modes: "" as written, or PromptStyleKali
}

// Colorize applies the per-line colorizer for a non-Prism render mode.
// Terminal output in the Kali prompt style may have more lines than the
// input, since each prompt becomes two lines plus a separator.
func Colorize(mode string, lines []string, o Options) []Line {
	pal := NewPalette(o.Theme)
	c := colorizer{prompt: o.Prompt, promptColor: o.PromptColor, kali: o.PromptStyle == PromptStyleKali, dim: pal.Dim, pal: pal}
	if c.promptColor == "" {
		c.promptColor = Teal
	}
	if mode == preset.Terminal && c.kali {
		return c.terminalKali(lines)
	}
	out := make([]Line, len(lines))
	for i, l := range lines {
		switch mode {
		case preset.Terminal:
			out[i] = c.terminal(l)
		case preset.Log:
			out[i] = c.log(l)
		case preset.Git:
			out[i] = c.git(l)
		case preset.Tree:
			out[i] = c.tree(l)
		case preset.Env:
			out[i] = c.env(l)
		case preset.HTTP:
			out[i] = c.http(l)
		case preset.Diff:
			out[i] = c.diff(l)
		case preset.Test:
			out[i] = c.test(l)
		case preset.Metrics:
			out[i] = c.metrics(l)
		default: // milestone and anything unknown: plain text
			out[i] = Line{plain(l)}
		}
	}
	return out
}

type colorizer struct {
	prompt      string
	promptColor string
	kali        bool // zsh-syntax-highlighting conventions: options stay plain
	dim         Span
	pal         Palette
}

func plain(s string) Span                { return Span{Text: s} }
func colored(s, c string) Span           { return Span{Text: s, Color: c} }
func (c colorizer) dimmed(s string) Span { d := c.dim; d.Text = s; return d }

var (
	logLevels = []struct {
		re    *regexp.Regexp
		color string
	}{
		{regexp.MustCompile(`\b(FATAL|ERROR|ERR|EXCEPTION)\b`), Red},
		{regexp.MustCompile(`\b(WARN|WARNING)\b`), Amber},
		{regexp.MustCompile(`\b(INFO|NOTICE)\b`), Teal},
		{regexp.MustCompile(`\b(DEBUG|TRACE|VERBOSE)\b`), Grey},
	}
	logTimestamp = regexp.MustCompile(`^[\d/:.\-T+Z\s]*\d`)
)

func (c colorizer) log(l string) Line {
	if strings.HasPrefix(l, "    ") || strings.HasPrefix(l, "\t") {
		return Line{c.dimmed(l)}
	}
	color := ""
	for _, lv := range logLevels {
		if lv.re.MatchString(l) {
			color = lv.color
			break
		}
	}
	ts := logTimestamp.FindString(l)
	var out Line
	if ts != "" {
		out = append(out, c.dimmed(ts))
	}
	return append(out, colored(l[len(ts):], color))
}

var gitHeader = regexp.MustCompile(`^(Author|Date|Merge):`)

func (c colorizer) git(l string) Line {
	switch {
	case strings.HasPrefix(l, "commit "):
		return Line{c.dimmed("commit "), colored(l[7:], Amber)}
	case gitHeader.MatchString(l):
		i := strings.IndexByte(l, ':')
		return Line{c.dimmed(l[:i+1]), plain(l[i+1:])}
	case strings.HasPrefix(l, "+"):
		return Line{colored(l, Teal)}
	case strings.HasPrefix(l, "-"):
		return Line{colored(l, Red)}
	}
	return Line{plain(l)}
}

var treeLine = regexp.MustCompile(`^([\s│├└─]*)(.*)$`)

func (c colorizer) tree(l string) Line {
	m := treeLine.FindStringSubmatch(l)
	prefix, name := "", l
	if m != nil {
		prefix, name = m[1], m[2]
	}
	nameColor := ""
	if strings.HasSuffix(name, "/") {
		nameColor = Amber
	}
	return Line{c.dimmed(prefix), colored(name, nameColor)}
}

func (c colorizer) env(l string) Line {
	if strings.HasPrefix(strings.TrimLeft(l, " \t\n\r"), "#") || strings.TrimSpace(l) == "" {
		return Line{c.dimmed(l)}
	}
	i := strings.IndexByte(l, '=')
	if i < 0 {
		return Line{plain(l)}
	}
	return Line{colored(l[:i], Teal), c.dimmed("="), colored(l[i+1:], Amber)}
}

var (
	httpRequest = regexp.MustCompile(`^(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)\b`)
	httpHeader  = regexp.MustCompile(`^([\w-]+):(\s.*)$`)
)

func (c colorizer) http(l string) Line {
	if httpRequest.MatchString(l) {
		sp := strings.IndexByte(l, ' ')
		if sp < 0 {
			return Line{colored(l, Amber)}
		}
		method, rest := l[:sp], l[sp:]
		out := Line{colored(method, Amber)}
		if h := strings.Index(rest, " HTTP/"); h >= 0 {
			out = append(out, plain(rest[:h]), c.dimmed(rest[h:]))
		} else {
			out = append(out, plain(rest))
		}
		return out
	}
	if m := httpHeader.FindStringSubmatch(l); m != nil {
		return Line{colored(m[1], Teal), c.dimmed(":"), plain(m[2])}
	}
	if strings.TrimSpace(l) == "" {
		return Line{c.dimmed(l)}
	}
	return Line{plain(l)}
}

var diffMeta = regexp.MustCompile(`^(diff |index |--- |\+\+\+ )`)

func (c colorizer) diff(l string) Line {
	switch {
	case diffMeta.MatchString(l):
		return Line{c.dimmed(l)}
	case strings.HasPrefix(l, "@@"):
		return Line{colored(l, Purple)}
	case strings.HasPrefix(l, "+"):
		return Line{colored(l, Teal)}
	case strings.HasPrefix(l, "-"):
		return Line{colored(l, Red)}
	}
	return Line{plain(l)}
}

var (
	testFail = regexp.MustCompile(`(?i)(✗|✘|✕|×|\bFAIL\b|\bFAILED\b|\bnot ok\b|\bfailing\b)`)
	testSkip = regexp.MustCompile(`(?i)(○|◯|⊘|\bSKIP\b|\bSKIPPED\b|\bpending\b|\btodo\b)`)
	testPass = regexp.MustCompile(`(?i)(✓|✔|√|\bPASS\b|\bPASSED\b|\bok\b|\bpassing\b)`)
)

func (c colorizer) test(l string) Line {
	switch {
	case testFail.MatchString(l):
		return Line{colored(l, Red)}
	case testSkip.MatchString(l):
		return Line{colored(l, Amber)}
	case testPass.MatchString(l):
		return Line{colored(l, Teal)}
	}
	return Line{plain(l)}
}

var metricLine = regexp.MustCompile(`^(.*?)(\d+(?:\.\d+)?)(\s?(?:%|ms|s|kB|KB|MB|mb|gb|GB|B))?\s*$`)

func (c colorizer) metrics(l string) Line {
	m := metricLine.FindStringSubmatch(l)
	if m == nil || (strings.TrimSpace(m[1]) == "" && m[3] == "") {
		return Line{plain(l)}
	}
	label, value, unit := m[1], m[2], m[3]
	color := Amber
	if unit == "" {
		n, _ := strconv.ParseFloat(value, 64)
		switch {
		case n >= 90:
			color = Teal
		case n >= 50:
			color = Amber
		default:
			color = Red
		}
	}
	return Line{plain(label), Span{Text: value + unit, Color: color, Bold: true}}
}
