package highlight

import (
	"regexp"
	"strings"
)

// Kali's terminal scheme, used by its two-line zsh prompt: the frame and
// commands in green, the user and host in blue.
const (
	KaliBlue  = "#367bf0"
	KaliGreen = "#5ebdab"
)

// PromptStyleKali expands prompts into Kali Linux's two-line zsh prompt:
//
//	┌──(kali㉿kali)-[~]
//	└─$ command
//
// with a blank line before every prompt block after the first.
const PromptStyleKali = "kali"

// Identity is who and where a prompt claims to be.
type Identity struct{ User, Host, Path, Symbol string }

// DefaultIdentity is Kali's out-of-the-box prompt.
var DefaultIdentity = Identity{"kali", "kali", "~", "$"}

var identityRe = regexp.MustCompile(`^\[?([\w.-]+)[@㉿]([\w.-]+)\]?(?::([^\s$#\]]*)|\s+([^\s\]$#]*)\]?)?\s*([$#])?\s*$`)

// ParseIdentity reads "kali@kali:~$", "root@box:/etc#", "[user@host dir]$"
// or plain "user@host" into an Identity, filling gaps from the default.
func ParseIdentity(s string) (Identity, bool) {
	m := identityRe.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return DefaultIdentity, false
	}
	id := Identity{User: m[1], Host: m[2], Path: m[3], Symbol: m[5]}
	if id.Path == "" {
		id.Path = m[4]
	}
	if id.Path == "" {
		id.Path = DefaultIdentity.Path
	}
	if id.Symbol == "" {
		id.Symbol = DefaultIdentity.Symbol
	}
	return id, true
}

var (
	kaliTopLine    = regexp.MustCompile(`^(\s*┌──\()([^)]+)(\)-\[)([^\]]*)(\])(.*)$`)
	kaliBottomLine = regexp.MustCompile(`^(\s*└─)([$#])( |$)`)
)

// terminalKali renders a session with Kali's two-line prompt. Prompt lines
// written as "$ cmd", "user@host:~$ cmd" or "[user@host dir]$ cmd" are
// expanded; lines that already carry the Kali prompt are colored as they
// are; everything else is output.
func (c colorizer) terminalKali(lines []string) []Line {
	var out []Line
	first := true // no blank line before the first prompt block
	emitPrompt := func(id Identity, cmd string) {
		if !first {
			out = append(out, Line{plain("")})
		}
		out = append(out, c.kaliTop(id.User+"㉿"+id.Host, id.Path))
		out = append(out, c.kaliBottom(id.Symbol, cmd))
		first = false
	}
	for _, l := range lines {
		if m := kaliTopLine.FindStringSubmatch(l); m != nil {
			out = append(out, Line{colored(m[1], KaliGreen), colored(m[2], KaliBlue), colored(m[3], KaliGreen), plain(m[4]), colored(m[5], KaliGreen), plain(m[6])})
			first = false
			continue
		}
		if m := kaliBottomLine.FindStringSubmatchIndex(l); m != nil {
			out = append(out, c.kaliBottom(l[m[4]:m[5]], l[m[5]:]))
			first = false
			continue
		}
		if id, cmd, ok := c.promptOf(l); ok {
			emitPrompt(id, cmd)
			continue
		}
		out = append(out, c.output(l))
		first = false
	}
	return out
}

// promptOf recognizes a prompt line and returns who it belongs to and the
// command after it. --prompt, when given, overrides the identity.
func (c colorizer) promptOf(l string) (Identity, string, bool) {
	id := DefaultIdentity
	if c.prompt != "" {
		if parsed, ok := ParseIdentity(c.prompt); ok {
			id = parsed
		}
	}
	for _, p := range promptChars {
		if strings.HasPrefix(l, string(p)) {
			rest := l[len(string(p)):]
			if rest == "" || rest[0] == ' ' {
				if p == '#' && c.prompt == "" {
					id.Symbol = "#"
				}
				return id, strings.TrimPrefix(rest, " "), true
			}
		}
	}
	for _, re := range userHostPrompts {
		if m := re.FindStringSubmatchIndex(l); m != nil {
			if c.prompt == "" {
				if parsed, ok := ParseIdentity(l[:m[5]]); ok {
					id = parsed
				}
			}
			return id, strings.TrimPrefix(l[m[5]:], " "), true
		}
	}
	return id, "", false
}

func (c colorizer) kaliTop(who, path string) Line {
	return Line{colored("┌──(", KaliGreen), colored(who, KaliBlue), colored(")-[", KaliGreen), plain(path), colored("]", KaliGreen)}
}

func (c colorizer) kaliBottom(symbol, cmd string) Line {
	line := Line{colored("└─"+symbol, KaliGreen)}
	cmd = strings.TrimPrefix(cmd, " ")
	if cmd == "" {
		return append(line, plain(" ")) // the space the cursor sits after
	}
	return append(line, c.command(" "+cmd)...)
}
