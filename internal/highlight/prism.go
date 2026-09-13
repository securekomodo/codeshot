package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"

	"github.com/securekomodo/codeshot/internal/theme"
)

// Language is one entry of the built-in language list.
type Language struct{ ID, Label string }

// Languages lists the built-in languages, keyed by Prism id.
var Languages = []Language{
	{"javascript", "JavaScript"}, {"typescript", "TypeScript"}, {"jsx", "JSX"}, {"tsx", "TSX"},
	{"python", "Python"}, {"go", "Go"}, {"rust", "Rust"}, {"c", "C"}, {"cpp", "C++"},
	{"swift", "Swift"}, {"kotlin", "Kotlin"}, {"json", "JSON"}, {"yaml", "YAML"}, {"sql", "SQL"},
	{"graphql", "GraphQL"}, {"css", "CSS"}, {"markup", "HTML"}, {"markdown", "Markdown"},
}

// LanguageIDs returns the language ids in dropdown order.
func LanguageIDs() []string {
	ids := make([]string, len(Languages))
	for i, l := range Languages {
		ids[i] = l.ID
	}
	return ids
}

// lexerAlias maps language ids to chroma lexer names where they differ.
var lexerAlias = map[string]string{"markup": "html"}

// lexerFor returns the chroma lexer for a language id, or for any chroma
// lexer name, falling back to plain text.
func lexerFor(lang string) chroma.Lexer {
	if lang == "regex" {
		return chroma.Coalesce(regexLexer)
	}
	name := lang
	if a, ok := lexerAlias[lang]; ok {
		name = a
	}
	l := lexers.Get(name)
	if l == nil {
		l = lexers.Fallback
	}
	return chroma.Coalesce(l)
}

// KnownLanguage reports whether lang is a built-in language, "regex", or a
// chroma lexer name.
func KnownLanguage(lang string) bool {
	if lang == "regex" {
		return true
	}
	if a, ok := lexerAlias[lang]; ok {
		lang = a
	}
	return lexers.Get(lang) != nil
}

// chromaToSite maps chroma lexer names back to the built-in ids.
var chromaToSite = map[string]string{
	"javascript": "javascript", "typescript": "typescript", "react": "jsx", "python": "python",
	"go": "go", "rust": "rust", "c": "c", "c++": "cpp", "swift": "swift", "kotlin": "kotlin",
	"json": "json", "yaml": "yaml", "sql": "sql", "mysql": "sql", "postgresql sql dialect": "sql",
	"transact-sql": "sql", "graphql": "graphql", "css": "css",
	"html": "markup", "markdown": "markdown",
}

// DetectLanguage guesses a language id from a filename ("" if unknown).
func DetectLanguage(filename string) string {
	l := lexers.Match(filename)
	if l == nil {
		return ""
	}
	name := strings.ToLower(l.Config().Name)
	if id, ok := chromaToSite[name]; ok {
		return id
	}
	return name
}

// Highlight tokenizes the prepared lines with the lexer for lang and styles
// them with the theme. The result has exactly one Line per input line.
func Highlight(lines []string, lang string, th *theme.Theme) ([]Line, error) {
	it, err := lexerFor(lang).Tokenise(&chroma.TokeniseOptions{State: "root", EnsureLF: true}, strings.Join(lines, "\n"))
	if err != nil {
		return nil, err
	}
	resolved := th.Resolve(lang)
	tokens := markCalls(it.Tokens(), lang)
	out := make([]Line, 0, len(lines))
	for _, toks := range chroma.SplitTokensIntoLines(tokens) {
		var line Line
		for _, tk := range toks {
			text := strings.TrimSuffix(tk.Value, "\n")
			if text == "" {
				continue
			}
			line = append(line, styled(text, resolved.StyleFor(prismTypes(tk, lang)...)))
		}
		out = append(out, line)
	}
	for len(out) < len(lines) {
		out = append(out, nil)
	}
	return out[:len(lines)], nil
}

// callLanguages are the languages whose Prism grammar tags any identifier
// followed by "(" as a function (the C-like family, Python, CSS...). chroma
// leaves such call sites as plain names, which loses the most visible color
// in themes like Dracula, so markCalls re-tags them.
var callLanguages = map[string]bool{
	"javascript": true, "typescript": true, "jsx": true, "tsx": true, "python": true, "go": true,
	"rust": true, "c": true, "cpp": true, "swift": true, "kotlin": true, "css": true,
}

// markCalls turns a plain name token directly followed by "(" (optionally
// after whitespace) into a NameFunction token.
func markCalls(tokens []chroma.Token, lang string) []chroma.Token {
	if !callLanguages[lang] {
		return tokens
	}
	for i, tk := range tokens {
		if tk.Type != chroma.Name && tk.Type != chroma.NameOther {
			continue
		}
		j := i + 1
		if j < len(tokens) && strings.TrimSpace(tokens[j].Value) == "" {
			j++
		}
		if j < len(tokens) && strings.HasPrefix(tokens[j].Value, "(") {
			tokens[i].Type = chroma.NameFunction
		}
	}
	return tokens
}

// styled builds a Span from a resolved theme style.
func styled(text string, st theme.Style) Span {
	sp := Span{Text: text, Italic: st.Italic(), Bold: st.Bold(),
		Underline: st.TextDecoration == "underline", Strike: st.TextDecoration == "line-through"}
	if st.Color != "" {
		if c, err := theme.ParseColor(st.Color); err == nil {
			sp.Color, sp.Opacity = c.Hex, c.Alpha
		}
	}
	if st.HasOpacity {
		if sp.Opacity == 0 {
			sp.Opacity = 1
		}
		sp.Opacity *= st.Opacity
	}
	return sp
}
