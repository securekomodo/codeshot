package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
)

// The themes are keyed by Prism token types. chroma emits its own token
// types, so each is mapped to the ordered
// list of Prism types a comparable Prism token would carry (its type, then
// aliases); theme.StyleFor merges them in that order.
var typeMap = map[chroma.TokenType][]string{
	chroma.Keyword:            {"keyword"},
	chroma.KeywordDeclaration: {"keyword"},
	chroma.KeywordNamespace:   {"keyword"},
	chroma.KeywordPseudo:      {"keyword"},
	chroma.KeywordReserved:    {"keyword"},
	chroma.KeywordType:        {"builtin"},
	chroma.NameKeyword:        {"keyword"},
	chroma.OperatorWord:       {"keyword"},

	chroma.NameAttribute:     {"attr-name"},
	chroma.NameTag:           {"tag"},
	chroma.NameClass:         {"class-name"},
	chroma.NameException:     {"class-name"},
	chroma.NameConstant:      {"constant"},
	chroma.NameDecorator:     {"decorator", "annotation"},
	chroma.NameEntity:        {"entity"},
	chroma.NameNamespace:     {"namespace"},
	chroma.NameOperator:      {"operator"},
	chroma.NameProperty:      {"property"},
	chroma.NamePseudo:        {"selector"},
	chroma.NameBuiltin:       {"builtin"},
	chroma.NameBuiltinPseudo: {"builtin"},
	chroma.NameVariable:      {"variable"},
	chroma.NameFunction:      {"function"},

	chroma.LiteralString:         {"string"},
	chroma.LiteralStringChar:     {"string", "char"},
	chroma.LiteralStringRegex:    {"regex"},
	chroma.LiteralStringInterpol: {"interpolation-punctuation", "punctuation"},
	chroma.LiteralStringSymbol:   {"symbol"},
	chroma.LiteralStringEscape:   {"string"},
	chroma.LiteralNumber:         {"number"},

	chroma.Operator:        {"operator"},
	chroma.Punctuation:     {"punctuation"},
	chroma.TextPunctuation: {"punctuation"},

	chroma.Comment:            {"comment"},
	chroma.CommentPreproc:     {"keyword"},
	chroma.CommentPreprocFile: {"string"},

	chroma.GenericDeleted:    {"deleted"},
	chroma.GenericInserted:   {"inserted"},
	chroma.GenericEmph:       {"italic"},
	chroma.GenericStrong:     {"bold"},
	chroma.GenericHeading:    {"title", "important"},
	chroma.GenericSubheading: {"title", "important"},
}

// langOverrides refine the mapping where Prism's grammar for a language
// names things differently from chroma's lexer.
var langOverrides = map[string]map[chroma.TokenType][]string{
	"json":     {chroma.NameTag: {"property"}},
	"yaml":     {chroma.NameTag: {"key", "atrule"}, chroma.CommentPreproc: {"punctuation"}},
	"css":      {chroma.NameTag: {"selector"}, chroma.NameClass: {"selector"}, chroma.NameDecorator: {"selector"}, chroma.LiteralStringOther: {"url"}},
	"markup":   {chroma.LiteralString: {"attr-value"}},
	"markdown": {chroma.LiteralStringBacktick: {"code", "keyword"}},
	"c":        {chroma.KeywordType: {"keyword"}},
	"cpp":      {chroma.KeywordType: {"keyword"}},
	"rust":     {chroma.KeywordType: {"keyword"}},
	"swift":    {chroma.KeywordType: {"keyword"}},
	"kotlin":   {chroma.KeywordType: {"keyword"}},
	"sql":      {chroma.KeywordType: {"keyword"}},
}

// prismTypes returns the Prism token types for a chroma token.
func prismTypes(tk chroma.Token, lang string) []string {
	if ov, ok := langOverrides[lang]; ok {
		if ts, ok := lookup(ov, tk.Type); ok {
			return ts
		}
	}
	switch tk.Type {
	case chroma.KeywordConstant:
		switch strings.ToLower(tk.Value) {
		case "true", "false":
			return []string{"boolean"}
		case "null", "nil", "none", "undefined":
			return []string{"keyword"}
		}
		return []string{"constant"}
	case chroma.CommentPreproc:
		if lang == "markup" {
			v := strings.ToUpper(tk.Value)
			switch {
			case strings.HasPrefix(v, "<!DOCTYPE"):
				return []string{"doctype"}
			case strings.HasPrefix(v, "<![CDATA["):
				return []string{"cdata"}
			case strings.HasPrefix(v, "<?"):
				return []string{"prolog"}
			}
		}
	}
	if ts, ok := lookup(typeMap, tk.Type); ok {
		return ts
	}
	return nil
}

// lookup tries the exact token type, then its sub-category, then its category.
func lookup(m map[chroma.TokenType][]string, t chroma.TokenType) ([]string, bool) {
	for _, k := range []chroma.TokenType{t, t.SubCategory(), t.Category()} {
		if ts, ok := m[k]; ok {
			return ts, true
		}
	}
	return nil, false
}
