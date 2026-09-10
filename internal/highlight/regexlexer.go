package highlight

import "github.com/alecthomas/chroma/v2"

// regexLexer approximates Prism's regex grammar for the regex preset, with
// chroma token types chosen so the generic mapping lands
// on the Prism types Prism's grammar aliases to: alternation and
// backreferences -> keyword, anchors -> function, quantifiers -> number,
// character sets -> class-name, groups and brackets -> punctuation, group
// names -> variable, negation and ranges -> operator.
func rule(pattern string, typ chroma.Emitter, mutator chroma.Mutator) chroma.Rule {
	return chroma.Rule{Pattern: pattern, Type: typ, Mutator: mutator}
}

var regexLexer = chroma.MustNewLexer(&chroma.Config{
	Name:    "regex",
	Aliases: []string{"regex", "regexp"},
}, func() chroma.Rules {
	return chroma.Rules{
		"root": {
			rule(`\\[1-9]\d*|\\k<[^>]+>`, chroma.Keyword, nil),
			rule(`\\[bBAZzG]|[\^$]`, chroma.NameFunction, nil),
			rule(`\\[wWsSdD]|\\[pP]\{[^}]*\}|\.`, chroma.NameClass, nil),
			rule(`\\.`, chroma.Text, nil),
			rule(`\|`, chroma.Keyword, nil),
			rule(`(\(\?<)([A-Za-z_][A-Za-z0-9_]*)(>)`, chroma.ByGroups(chroma.Punctuation, chroma.NameVariable, chroma.Punctuation), nil),
			rule(`(\(\?')([A-Za-z_][A-Za-z0-9_]*)(')`, chroma.ByGroups(chroma.Punctuation, chroma.NameVariable, chroma.Punctuation), nil),
			rule(`\((?:\?(?:[:=!>]|<[=!]))?|\)`, chroma.Punctuation, nil),
			rule(`(?:[+*?]|\{\d+(?:,\d*)?\})[?+]?`, chroma.LiteralNumber, nil),
			rule(`(\[)(\^?)`, chroma.ByGroups(chroma.Punctuation, chroma.Operator), chroma.Push("class")),
			rule(`\s+`, chroma.Text, nil),
			rule(`.`, chroma.Text, nil),
		},
		"class": {
			rule(`\]`, chroma.Punctuation, chroma.Pop(1)),
			rule(`\\[wWsSdD]|\\[pP]\{[^}]*\}`, chroma.NameClass, nil),
			rule(`\\.`, chroma.Text, nil),
			rule(`-`, chroma.Operator, nil),
			rule(`\s+`, chroma.Text, nil),
			rule(`.`, chroma.Text, nil),
		},
	}
})
