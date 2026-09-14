package main

import (
	"bytes"
	"flag"

	"github.com/securekomodo/codeshot/internal/highlight"
	"github.com/securekomodo/codeshot/internal/preset"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseInterspersed(t *testing.T) {
	cases := []struct {
		args []string
		pos  []string
		out  string
	}{
		{[]string{"-o", "x.png", "main.go"}, []string{"main.go"}, "x.png"},
		{[]string{"main.go", "-o", "x.png"}, []string{"main.go"}, "x.png"},
		{[]string{"a", "-o", "x.png", "b"}, []string{"a", "b"}, "x.png"},
		{[]string{"--", "-weird.go"}, []string{"-weird.go"}, ""},
		{[]string{"-"}, []string{"-"}, ""},
		{nil, nil, ""},
	}
	for _, c := range cases {
		fs := flag.NewFlagSet("t", flag.ContinueOnError)
		out := fs.String("o", "", "")
		pos, err := parseInterspersed(fs, c.args)
		if err != nil {
			t.Fatalf("%v: %v", c.args, err)
		}
		if !reflect.DeepEqual(pos, c.pos) || *out != c.out {
			t.Errorf("%v: positional %v (want %v), -o %q (want %q)", c.args, pos, c.pos, *out, c.out)
		}
	}
}

func TestSniffedPresetFromStdin(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "d.svg")
	diff := "diff --git a/x b/x\n--- a/x\n+++ b/x\n@@ -1 +1 @@\n-a\n+b\n"
	var stdout bytes.Buffer
	if err := run([]string{"-o", out}, strings.NewReader(diff), &stdout); err != nil {
		t.Fatal(err)
	}
	svg, _ := os.ReadFile(out)
	if !strings.Contains(string(svg), "changes.diff") {
		t.Error("piped diff should pick the git-diff preset (title changes.diff)")
	}
	// An explicit preset or language is left alone.
	if err := run([]string{"--lang", "go", "-o", out}, strings.NewReader(diff), &stdout); err != nil {
		t.Fatal(err)
	}
	svg, _ = os.ReadFile(out)
	if !strings.Contains(string(svg), "snippet.go") {
		t.Error("--lang keeps the code preset, titled after the language")
	}
	// A file gets its name as the title.
	src := filepath.Join(dir, "main.go")
	os.WriteFile(src, []byte("package main\n"), 0o644)
	if err := run([]string{src, "-o", out}, strings.NewReader(""), &stdout); err != nil {
		t.Fatal(err)
	}
	svg, _ = os.ReadFile(out)
	if !strings.Contains(string(svg), ">main.go<") {
		t.Error("file name should be the title")
	}
}

func TestPickRandom(t *testing.T) {
	ids := []string{"a", "b", "none"}
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		id := pickRandom(ids, "none")
		if id == "none" {
			t.Fatal("excluded id picked")
		}
		seen[id] = true
	}
	if len(seen) != 2 {
		t.Errorf("only saw %v", seen)
	}
	out := filepath.Join(t.TempDir(), "r.svg")
	var stdout bytes.Buffer
	if err := run([]string{"--bg", "random", "--theme", "random", "--sample", "-o", out}, strings.NewReader(""), &stdout); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatal(err)
	}
}

// A snippet that arrives without a file name is named after its language,
// rather than always claiming to be JavaScript.
func TestUntitledSnippetTakesTheLanguageExtension(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "a.svg")
	var stdout bytes.Buffer
	cases := map[string]string{
		"powershell": "snippet.ps1",
		"python":     "snippet.py",
		"bash":       "snippet.bash", // not in the built-in list
		"":           "snippet.js",   // nothing chosen, so the preset default stands
	}
	for lang, want := range cases {
		args := []string{"-o", out}
		if lang != "" {
			args = append(args, "--lang", lang)
		}
		if err := run(args, strings.NewReader("x = 1\n"), &stdout); err != nil {
			t.Fatalf("%s: %v", lang, err)
		}
		svg, _ := os.ReadFile(out)
		if !strings.Contains(string(svg), ">"+want+"<") {
			t.Errorf("--lang %q should title the window %q", lang, want)
		}
	}
	// A file name and an explicit title both outrank the language.
	src := filepath.Join(dir, "audit.ps1")
	os.WriteFile(src, []byte("Write-Host 1\n"), 0o644)
	if err := run([]string{src, "-o", out}, strings.NewReader(""), &stdout); err != nil {
		t.Fatal(err)
	}
	if svg, _ := os.ReadFile(out); !strings.Contains(string(svg), ">audit.ps1<") {
		t.Error("a file should be titled after its name")
	}
	if err := run([]string{"--lang", "python", "--title", "chosen", "-o", out}, strings.NewReader("x=1\n"), &stdout); err != nil {
		t.Fatal(err)
	}
	if svg, _ := os.ReadFile(out); !strings.Contains(string(svg), ">chosen<") {
		t.Error("--title should win")
	}
}

func TestListAndUsageErrors(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"ignored.go", "--list", "presets"}, strings.NewReader(""), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "terminal-session") {
		t.Errorf("list presets output: %q", out.String())
	}
	if err := run([]string{"a.go", "b.go"}, strings.NewReader(""), &out); err == nil {
		t.Error("two files should be a usage error")
	}
	if err := run([]string{"--preset", "nope", "--sample"}, strings.NewReader(""), &out); err == nil {
		t.Error("unknown preset should fail")
	}
	if err := run([]string{"--lang", "nope-lang", "--sample"}, strings.NewReader(""), &out); err == nil {
		t.Error("unknown language should fail")
	}
}

// The grouped help is written by hand, so make sure it documents every flag
// the program actually registers.
func TestHelpDocumentsEveryFlag(t *testing.T) {
	help := usage("1.0.0")
	var o options
	n := 0
	newFlagSet(&o).VisitAll(func(f *flag.Flag) {
		n++
		if !strings.Contains(help, "--"+f.Name) && !strings.Contains(help, "-"+f.Name+",") {
			t.Errorf("flag --%s is not documented in the help text", f.Name)
		}
	})
	if n < 30 {
		t.Fatalf("expected the full flag set, got %d flags", n)
	}
	for _, p := range preset.All {
		if !strings.Contains(help, p.Key) || !strings.Contains(help, p.Desc) {
			t.Errorf("preset %s is not listed in the help text", p.Key)
		}
	}
	for _, id := range highlight.LanguageIDs() {
		if !strings.Contains(help, id) {
			t.Errorf("language %s is not listed in the help text", id)
		}
	}
	for _, want := range []string{author, projectURL, tagline, "Usage:", "Examples:", "Author:"} {
		if !strings.Contains(help, want) {
			t.Errorf("help is missing %q", want)
		}
	}
}

func TestVersionShowsAuthor(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"--version"}, strings.NewReader(""), &out); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 || !strings.HasPrefix(lines[0], "codeshot ") || lines[1] != author || lines[2] != projectURL {
		t.Errorf("--version output:\n%s", out.String())
	}
}
