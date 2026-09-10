package highlight

import (
	"testing"

	"codeshot/internal/preset"
	"codeshot/internal/theme"
)

func find(lines []Line, text string) (Span, bool) {
	for _, l := range lines {
		for _, sp := range l {
			if sp.Text == text {
				return sp, true
			}
		}
	}
	return Span{}, false
}

func TestHighlightJavaScript(t *testing.T) {
	dr, _ := theme.Get("dracula")
	lines, err := Highlight([]string{`const x = "s"; // note`, "", "return true;"}, "javascript", dr)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 3 || len(lines[1]) != 0 {
		t.Fatalf("lines = %+v", lines)
	}
	if kw, ok := find(lines, "const"); !ok || kw.Color != "#bd93f9" || !kw.Italic {
		t.Errorf("const = %+v", kw)
	}
	if s, ok := find(lines, `"s"`); !ok || s.Color != "#ff79c6" {
		t.Errorf("string = %+v", s)
	}
	if c, ok := find(lines, "// note"); !ok || c.Color != "#6272a4" {
		t.Errorf("comment = %+v", c)
	}
	if b, ok := find(lines, "true"); !ok || b.Color != "" {
		t.Errorf("dracula leaves booleans plain, got %+v", b)
	}
	// Prism colors any identifier followed by "(" as a function.
	lines, _ = Highlight([]string{"fetchSnippet(id).then(console.log)", "foo (1)", "notCall"}, "javascript", dr)
	for _, name := range []string{"fetchSnippet", "then", "foo"} {
		if f, ok := find(lines, name); !ok || f.Color != "#50fa7b" {
			t.Errorf("%s should be function-colored: %+v", name, f)
		}
	}
	if n, _ := find(lines, "notCall"); n.Color != "" {
		t.Errorf("notCall should stay plain: %+v", n)
	}
	gh, _ := theme.Get("github")
	lines, _ = Highlight([]string{"return true;"}, "javascript", gh)
	if b, _ := find(lines, "true"); b.Color != "#36acaa" {
		t.Errorf("github boolean = %+v", b)
	}
}

func TestHighlightLanguages(t *testing.T) {
	gh, _ := theme.Get("github")
	lines, _ := Highlight([]string{`{"id": 1}`}, "json", gh)
	if k, ok := find(lines, `"id"`); !ok || k.Color != "#36acaa" {
		t.Errorf("json key should be property-colored: %+v", k)
	}
	vs, _ := theme.Get("vsDark")
	lines, _ = Highlight([]string{`<div class="a">`}, "markup", vs)
	if tag, ok := find(lines, "div"); !ok || tag.Color != "#569cd6" {
		t.Errorf("markup tag uses the language-specific vsDark color: %+v", tag)
	}
	if p, ok := find(lines, "<"); !ok || p.Color != "#808080" {
		t.Errorf("markup punctuation: %+v", p)
	}
	oc, _ := theme.Get("oceanicNext")
	lines, _ = Highlight([]string{"import os"}, "python", oc)
	if ns, ok := find(lines, "os"); !ok || ns.Opacity != 0.7 {
		t.Errorf("namespace opacity: %+v", ns)
	}
	for _, id := range LanguageIDs() {
		if _, err := Highlight([]string{"x = 1"}, id, gh); err != nil {
			t.Errorf("%s: %v", id, err)
		}
		if !KnownLanguage(id) {
			t.Errorf("%s unknown", id)
		}
	}
	if !KnownLanguage("bash") || KnownLanguage("nope-lang") {
		t.Error("KnownLanguage")
	}
}

func TestRegexLexer(t *testing.T) {
	dr, _ := theme.Get("dracula")
	p, _ := preset.Get("regex")
	sample, _ := p.Sample()
	lines, err := Highlight([]string{sample, `^(?<a>x|y)[^0-9]+$`}, "regex", dr)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 {
		t.Fatalf("lines = %d", len(lines))
	}
	if alt, ok := find(lines, "|"); !ok || alt.Color != "#bd93f9" {
		t.Errorf("alternation should be keyword-colored: %+v", alt)
	}
	if anchor, ok := find(lines, "^"); !ok || anchor.Color != "#50fa7b" {
		t.Errorf("anchor should be function-colored: %+v", anchor)
	}
	if name, ok := find(lines, "a"); !ok || name.Color != "#bd93f9" {
		t.Errorf("group name should be variable-colored: %+v", name)
	}
	if q, ok := find(lines, "+"); !ok || q.Color != "" {
		t.Errorf("dracula leaves numbers (quantifiers) plain: %+v", q)
	}
}

func TestDetectLanguage(t *testing.T) {
	cases := map[string]string{
		"main.go": "go", "app.tsx": "typescript", "x.jsx": "jsx", "index.html": "markup",
		"README.md": "markdown", "q.sql": "sql", "conf.yml": "yaml", "unknown.zzz": "",
	}
	for f, want := range cases {
		if got := DetectLanguage(f); got != want {
			t.Errorf("%s = %q want %q", f, got, want)
		}
	}
}
