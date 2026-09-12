package highlight

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The terminal colorizer treats a session the way a shell with syntax
// highlighting would: prompt lines are tokenized as commands, and output
// lines get light-touch inline highlights (status words, numbers, hashes,
// URLs, bracket tags, timestamps) on top of the theme's plain color.

var (
	promptChars = []rune{'$', '❯', '#', '~'}
	// user@host:~$ and [user@host dir]$ style prompts, with or without a command.
	userHostPrompts = []*regexp.Regexp{
		regexp.MustCompile(`^([\w.-]+@[\w.-]+)([^\s$#]*[$#])( |$)`),
		regexp.MustCompile(`^(\[[\w.-]+@[\w.-]+)([^\]\n]*\][$#])( |$)`),
	}
	userHost = regexp.MustCompile(`^(\[?[\w.-]+@[\w.-]+)(.*)$`)
)

func (c colorizer) terminal(l string) Line {
	first, size := utf8.DecodeRuneInString(l)
	for _, p := range promptChars {
		if first == p && (len(l) == size || l[size] == ' ') {
			shown := c.prompt
			if shown == "" {
				shown = string(first)
			}
			return append(c.promptSpans(shown), c.command(l[size:])...)
		}
	}
	for _, re := range userHostPrompts {
		if m := re.FindStringSubmatchIndex(l); m != nil {
			shown := l[:m[5]]
			if c.prompt != "" {
				shown = c.prompt
			}
			return append(c.promptSpans(shown), c.command(l[m[5]:])...)
		}
	}
	return c.output(l)
}

// promptSpans colors a prompt: a user@host prefix gets the prompt color and
// the rest (":~$") stays plain, as in a bash prompt; anything else is
// colored whole.
func (c colorizer) promptSpans(shown string) Line {
	if m := userHost.FindStringSubmatch(shown); m != nil && m[2] != "" {
		return Line{colored(m[1], c.promptColor), plain(m[2])}
	}
	return Line{colored(shown, c.promptColor)}
}

// Shell words that introduce another command after them.
var commandPrefixes = map[string]bool{"sudo": true, "time": true, "exec": true, "nohup": true, "env": true, "doas": true, "xargs": true}

var (
	shellOperator = regexp.MustCompile(`^(\|\||&&|;;|[|;&]|\d?>>?&?\d?|<<?|\(|\))`)
	shellNumber   = regexp.MustCompile(`^\d+(\.\d+)?[a-zA-Z%]{0,3}$`)
	shellAssign   = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)
)

// command highlights one shell command line (everything after the prompt).
func (c colorizer) command(s string) Line {
	var out Line
	expectCommand := true
	i := 0
	for i < len(s) {
		// Whitespace is passed through untouched.
		j := i
		for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
			j++
		}
		if j > i {
			out = append(out, plain(s[i:j]))
			i = j
			continue
		}
		if s[i] == '#' {
			out = append(out, c.pal.span(c.pal.Comment, s[i:]))
			break
		}
		if m := shellOperator.FindString(s[i:]); m != "" {
			out = append(out, c.pal.span(c.pal.Operator, m))
			i += len(m)
			expectCommand = true
			continue
		}
		tok := shellWord(s[i:])
		i += len(tok)
		switch {
		case tok[0] == '"' || tok[0] == '\'' || tok[0] == '`':
			out = append(out, c.pal.span(c.pal.String, tok))
		case tok[0] == '$':
			out = append(out, c.pal.span(c.pal.Variable, tok))
		case strings.HasPrefix(tok, "http://") || strings.HasPrefix(tok, "https://"):
			out = append(out, c.pal.span(c.pal.URL, tok))
		case expectCommand && shellAssign.MatchString(tok):
			out = append(out, c.pal.span(c.pal.Variable, tok)) // FOO=bar prefix; the command follows
			continue
		case expectCommand:
			out = append(out, c.pal.span(c.pal.Function, tok))
			expectCommand = commandPrefixes[tok]
			continue
		case len(tok) > 1 && tok[0] == '-' && !c.kali:
			out = append(out, c.pal.span(c.pal.Keyword, tok))
		case shellNumber.MatchString(tok):
			out = append(out, c.pal.span(c.pal.Number, tok))
		default:
			out = append(out, plain(tok))
		}
		expectCommand = false
	}
	return out
}

// shellWord returns the next word of s, keeping quoted strings together.
func shellWord(s string) string {
	if q := s[0]; q == '"' || q == '\'' || q == '`' {
		for i := 1; i < len(s); i++ {
			if s[i] == '\\' {
				i++
				continue
			}
			if s[i] == q {
				return s[:i+1]
			}
		}
		return s
	}
	for i, r := range s {
		if unicode.IsSpace(r) || strings.ContainsRune("|;&()<>", r) {
			if i == 0 {
				_, n := utf8.DecodeRuneInString(s)
				return s[:n]
			}
			return s[:i]
		}
	}
	return s
}

// Inline patterns for output lines, tried at every position; the earliest
// match wins, ties go to the first pattern.
type inlineRule struct {
	re   *regexp.Regexp
	kind string
}

var inlineRules = []inlineRule{
	{regexp.MustCompile(`https?://[^\s'"<>)\]]+`), "url"},
	{regexp.MustCompile(`"[^"\n]*"|'[^'\n]*'`), "string"},
	{regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}(?:[T ]\d{2}:\d{2}(?::\d{2}(?:\.\d+)?)?Z?)?\b|\b\d{1,2}:\d{2}:\d{2}(?:\.\d+)?\b`), "dim"},
	{regexp.MustCompile(`\[[^\[\]\n]{1,40}\]`), "tag"},
	{regexp.MustCompile(`(?i)\b(?:ok|success|successful|successfully|done|passed|pass|complete|completed|ready|healthy|up-to-date|created|added|installed|resolved)\b|[✔✓√]`), "good"},
	{regexp.MustCompile(`(?i)\b(?:error|errors|err|fail|failed|failing|failure|fatal|panic|denied|invalid|missing|refused|rejected|abort|aborted)\b|[✗✘×]`), "bad"},
	{regexp.MustCompile(`(?i)\b(?:warn|warning|warnings|deprecated|skipped|skip|pending|todo|retry|retrying)\b|[○◯⊘]`), "warn"},
	{regexp.MustCompile(`\b[0-9a-f]{7,40}(?:\.\.[0-9a-f]{7,40})?\b`), "hash"},
	{regexp.MustCompile(`\b\d+(?:[.,]\d+)*(?:%|\s?(?:[KMGTkmgt]i?[Bb](?:/s)?|ms|ns|µs|us|s|min|h|px|x)\b)?`), "number"},
	{regexp.MustCompile(`->|=>|→|←|\.\.\.|…`), "operator"},
}

func hasDigit(s string) bool {
	return strings.ContainsAny(s, "0123456789")
}

// output highlights one non-prompt line.
func (c colorizer) output(l string) Line {
	trimmed := strings.TrimLeft(l, " \t")
	switch {
	case trimmed == "":
		return Line{plain(l)}
	case strings.HasPrefix(trimmed, "> ") || trimmed == ">":
		return Line{c.dimmed(l)} // a script runner echoing the command it runs
	case strings.HasPrefix(trimmed, "#"):
		return Line{c.pal.span(c.pal.Comment, l)}
	}
	return c.inline(l)
}

// inline applies the inline rules to a line of output.
func (c colorizer) inline(l string) Line {
	var out Line
	pos := 0
	for pos < len(l) {
		best, bestRule := -1, inlineRule{}
		var bestLoc []int
		for _, r := range inlineRules {
			loc := r.re.FindStringIndex(l[pos:])
			if loc == nil || loc[1] == loc[0] {
				continue
			}
			if best < 0 || loc[0] < best {
				best, bestRule, bestLoc = loc[0], r, loc
			}
		}
		if best < 0 {
			out = append(out, plain(l[pos:]))
			break
		}
		start, end := pos+bestLoc[0], pos+bestLoc[1]
		if start > pos {
			out = append(out, plain(l[pos:start]))
		}
		text := l[start:end]
		var sp Span
		switch bestRule.kind {
		case "url":
			sp = c.pal.span(c.pal.URL, text)
		case "string":
			sp = c.pal.span(c.pal.String, text)
		case "dim":
			sp = c.dimmed(text)
		case "tag":
			sp = c.pal.span(c.pal.Keyword, text)
		case "good":
			sp = colored(text, Teal)
		case "bad":
			sp = colored(text, Red)
		case "warn":
			sp = colored(text, Amber)
		case "hash":
			if hasDigit(text) {
				sp = c.pal.span(c.pal.Number, text)
			} else {
				sp = plain(text)
			}
		case "number":
			sp = c.pal.span(c.pal.Number, text)
		case "operator":
			sp = c.pal.span(c.pal.Operator, text)
		}
		out = append(out, sp)
		pos = end
	}
	return mergePlain(out)
}

// mergePlain joins adjacent unstyled spans.
func mergePlain(l Line) Line {
	var out Line
	for _, sp := range l {
		if n := len(out); n > 0 && sp.Color == "" && sp.Opacity == 0 && out[n-1].Color == "" && out[n-1].Opacity == 0 {
			out[n-1].Text += sp.Text
			continue
		}
		out = append(out, sp)
	}
	return out
}
