package main

import (
	"bytes"
	"flag"
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
	if !strings.Contains(string(svg), "snippet.js") {
		t.Error("--lang keeps the code preset")
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
