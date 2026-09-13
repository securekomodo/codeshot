package highlight

import (
	"reflect"
	"testing"

	"github.com/securekomodo/codeshot/internal/theme"
)

func TestColorizers(t *testing.T) {
	dim := Span{Color: "#e6e8eb", Opacity: 0.55}
	d := func(s string) Span { x := dim; x.Text = s; return x }
	p := func(s string) Span { return Span{Text: s} }
	c := func(s, col string) Span { return Span{Text: s, Color: col} }

	cases := []struct {
		mode, in string
		want     Line
	}{
		{"log", "2026-06-25 10:04:11  INFO   Server listening", Line{d("2026-06-25 10:04:11"), c("  INFO   Server listening", Teal)}},
		{"log", "10:04:21  ERROR  Unhandled rejection", Line{d("10:04:21"), c("  ERROR  Unhandled rejection", Red)}},
		{"log", "    at fetchSnippet (app/lib/data.ts:14:11)", Line{d("    at fetchSnippet (app/lib/data.ts:14:11)")}},
		{"log", "WARN slow", Line{c("WARN slow", Amber)}},
		{"log", "DEBUG x", Line{c("DEBUG x", Grey)}},
		{"log", "plain text", Line{c("plain text", "")}},
		{"git", "commit 8Kq2f4a9", Line{d("commit "), c("8Kq2f4a9", Amber)}},
		{"git", "Author: Jo <jo@x.io>", Line{d("Author:"), p(" Jo <jo@x.io>")}},
		{"git", "+added", Line{c("+added", Teal)}},
		{"git", "-removed", Line{c("-removed", Red)}},
		{"git", "    fix: thing", Line{p("    fix: thing")}},
		{"tree", "├── app/", Line{d("├── "), c("app/", Amber)}},
		{"tree", "│   └── page.tsx", Line{d("│   └── "), c("page.tsx", "")}},
		{"tree", "github.com/securekomodo/codeshot/", Line{d(""), c("github.com/securekomodo/codeshot/", Amber)}},
		{"env", "# Database", Line{d("# Database")}},
		{"env", "", Line{d("")}},
		{"env", "DATABASE_URL=postgres://x", Line{c("DATABASE_URL", Teal), d("="), c("postgres://x", Amber)}},
		{"env", "NOEQUALS", Line{p("NOEQUALS")}},
		{"http", "POST /api/snippets HTTP/1.1", Line{c("POST", Amber), p(" /api/snippets"), d(" HTTP/1.1")}},
		{"http", "GET /x", Line{c("GET", Amber), p(" /x")}},
		{"http", "Content-Type: application/json", Line{c("Content-Type", Teal), d(":"), p(" application/json")}},
		{"http", "", Line{d("")}},
		{"http", `{"language": "typescript"}`, Line{p(`{"language": "typescript"}`)}},
		{"diff", "diff --git a/x b/x", Line{d("diff --git a/x b/x")}},
		{"diff", "--- a/x", Line{d("--- a/x")}},
		{"diff", "+++ b/x", Line{d("+++ b/x")}},
		{"diff", "@@ -1,3 +1,4 @@", Line{c("@@ -1,3 +1,4 @@", Purple)}},
		{"diff", "+new", Line{c("+new", Teal)}},
		{"diff", "-old", Line{c("-old", Red)}},
		{"diff", " ctx", Line{p(" ctx")}},
		{"test", "PASS  app/lib/export.test.ts", Line{c("PASS  app/lib/export.test.ts", Teal)}},
		{"test", "  ✓ renders (12 ms)", Line{c("  ✓ renders (12 ms)", Teal)}},
		{"test", "  ✗ fails", Line{c("  ✗ fails", Red)}},
		{"test", "  ○ skipped", Line{c("  ○ skipped", Amber)}},
		{"test", "Tests: 1 failed, 2 passed", Line{c("Tests: 1 failed, 2 passed", Red)}},
		{"test", "Time: 1.2s", Line{p("Time: 1.2s")}},
		{"metrics", "Performance              98", Line{p("Performance              "), {Text: "98", Color: Teal, Bold: true}}},
		{"metrics", "SEO 72", Line{p("SEO "), {Text: "72", Color: Amber, Bold: true}}},
		{"metrics", "PWA 30", Line{p("PWA "), {Text: "30", Color: Red, Bold: true}}},
		{"metrics", "First Contentful Paint 0.8 s", Line{p("First Contentful Paint "), {Text: "0.8 s", Color: Amber, Bold: true}}},
		{"metrics", "Bundle 142kB", Line{p("Bundle "), {Text: "142kB", Color: Amber, Bold: true}}},
		{"metrics", "Lighthouse · example.com", Line{p("Lighthouse · example.com")}},
		{"metrics", "42", Line{p("42")}},
		{"milestone", "🎉 10,000 snippets shared", Line{p("🎉 10,000 snippets shared")}},
	}
	dr, _ := theme.Get("dracula")
	for _, tc := range cases {
		got := Colorize(tc.mode, []string{tc.in}, Options{Theme: dr})[0]
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s %q:\n got %+v\nwant %+v", tc.mode, tc.in, got, tc.want)
		}
	}
}

func TestTerminalCommands(t *testing.T) {
	dr, _ := theme.Get("dracula")
	pal := NewPalette(dr)
	got := Colorize("terminal", []string{
		`$ git commit -m "Add post" && git push`,
		"❯ FOO=1 sudo make -j4 # build",
		"$npm",
	}, Options{Theme: dr})
	want0 := Line{
		{Text: "$", Color: Teal}, {Text: " "}, pal.span(pal.Function, "git"), {Text: " "}, {Text: "commit"}, {Text: " "},
		pal.span(pal.Keyword, "-m"), {Text: " "}, pal.span(pal.String, `"Add post"`), {Text: " "},
		pal.span(pal.Operator, "&&"), {Text: " "}, pal.span(pal.Function, "git"), {Text: " "}, {Text: "push"},
	}
	if !reflect.DeepEqual(got[0], want0) {
		t.Errorf("command line:\n got %+v\nwant %+v", got[0], want0)
	}
	want1 := Line{
		{Text: "❯", Color: Teal}, {Text: " "}, pal.span(pal.Variable, "FOO=1"), {Text: " "}, pal.span(pal.Function, "sudo"),
		{Text: " "}, pal.span(pal.Function, "make"), {Text: " "}, pal.span(pal.Keyword, "-j4"), {Text: " "}, pal.span(pal.Comment, "# build"),
	}
	if !reflect.DeepEqual(got[1], want1) {
		t.Errorf("prefix/assign/comment:\n got %+v\nwant %+v", got[1], want1)
	}
	if got[2][0].Text != "$npm" || got[2][0].Color != "" {
		t.Errorf("no prompt space: %+v", got[2])
	}
	if pal.Function.Color != "#50fa7b" || pal.String.Color != "#ff79c6" || pal.Number.Color != Amber {
		t.Errorf("palette from theme: %+v", pal)
	}
}

func TestUserHostPrompts(t *testing.T) {
	dr, _ := theme.Get("dracula")
	pal := NewPalette(dr)
	got := Colorize("terminal", []string{
		"kali@kali:~$ echo hello world",
		"hello world",
		"kali@kali:~$",
		"[root@box /etc]# ls -la",
		"$ ls",
	}, Options{Theme: dr, PromptColor: "#3c82f6"})
	if got[0][0].Text != "kali@kali" || got[0][0].Color != "#3c82f6" || got[0][1].Text != ":~$" || got[0][1].Color != "" ||
		got[0][3].Text != "echo" || got[0][3].Color != pal.Function.Color {
		t.Errorf("kali prompt: %+v", got[0])
	}
	if got[1][0].Color != "" {
		t.Errorf("output should be plain: %+v", got[1])
	}
	if len(got[2]) != 2 || got[2][0].Text != "kali@kali" || got[2][1].Text != ":~$" {
		t.Errorf("bare prompt: %+v", got[2])
	}
	if got[3][0].Text != "[root@box" || got[3][0].Color != "#3c82f6" || got[3][1].Text != " /etc]#" {
		t.Errorf("bracket prompt: %+v", got[3])
	}
	if got[4][0].Text != "$" || got[4][0].Color != "#3c82f6" {
		t.Errorf("prompt color applies to $ too: %+v", got[4])
	}
	// --prompt can replace "$" with a full user@host prompt.
	rep := Colorize("terminal", []string{"$ ls", "$"}, Options{Theme: dr, Prompt: "kali@kali:~$"})
	if rep[0][0].Text != "kali@kali" || rep[0][0].Color != Teal || rep[0][1].Text != ":~$" || rep[1][1].Text != ":~$" {
		t.Errorf("prompt replacement: %+v %+v", rep[0], rep[1])
	}
}

func TestTerminalOutput(t *testing.T) {
	dr, _ := theme.Get("dracula")
	pal := NewPalette(dr)
	out := Colorize("terminal", []string{
		"[main 8c41f2e] Add post",
		" 3 files changed, 142 insertions(+)",
		"Writing objects: 100% (7/7), 1.21 MiB | 9.80 MiB/s, done.",
		"To https://github.com/example/site.git",
		"   623c873..8c41f2e  main -> main",
		"> astro build",
		"23:12:40 [build] 90 page(s) built in 6.31s",
		"✔ Project name … my-shots",
		"error: something failed",
		"",
	}, Options{Theme: dr})
	find := func(i int, text string) (Span, bool) {
		for _, sp := range out[i] {
			if sp.Text == text {
				return sp, true
			}
		}
		return Span{}, false
	}
	checks := []struct {
		line  int
		text  string
		color string
	}{
		{0, "[main 8c41f2e]", pal.Keyword.Color}, {0, " Add post", ""},
		{1, "3", pal.Number.Color}, {1, "142", pal.Number.Color},
		{2, "100%", pal.Number.Color}, {2, "1.21 MiB", pal.Number.Color}, {2, "9.80 MiB/s", pal.Number.Color}, {2, "done", Teal},
		{3, "https://github.com/example/site.git", pal.URL.Color},
		{4, "623c873..8c41f2e", pal.Number.Color}, {4, "->", pal.Operator.Color},
		{6, "23:12:40", pal.Dim.Color}, {6, "[build]", pal.Keyword.Color}, {6, "6.31s", pal.Number.Color},
		{7, "✔", Teal}, {8, "error", Red}, {8, "failed", Red},
	}
	for _, ck := range checks {
		sp, ok := find(ck.line, ck.text)
		if !ok || sp.Color != ck.color {
			t.Errorf("line %d %q: got %+v want color %q", ck.line, ck.text, sp, ck.color)
		}
	}
	if out[5][0].Color != pal.Dim.Color || out[5][0].Text != "> astro build" {
		t.Errorf("script echo should be dim: %+v", out[5])
	}
	if len(out[9]) != 1 || out[9][0].Text != "" {
		t.Errorf("empty line: %+v", out[9])
	}
	gh, _ := theme.Get("github")
	light := Colorize("terminal", []string{"12:00:00 x"}, Options{Theme: gh})
	if light[0][0].Color != "#000000" || light[0][0].Opacity != 0.5 {
		t.Errorf("light dim: %+v", light[0])
	}
}

func TestKaliPromptStyle(t *testing.T) {
	th, _ := theme.Get("kali")
	out := Colorize("terminal", []string{
		"$ echo hello world",
		"hello world",
		"$",
	}, Options{Theme: th, PromptStyle: PromptStyleKali})
	texts := make([]string, len(out))
	for i, l := range out {
		texts[i] = l.Text()
	}
	want := []string{"┌──(kali㉿kali)-[~]", "└─$ echo hello world", "hello world", "", "┌──(kali㉿kali)-[~]", "└─$ "}
	if !reflect.DeepEqual(texts, want) {
		t.Fatalf("expanded lines:\n got %q\nwant %q", texts, want)
	}
	top, bottom := out[0], out[1]
	if top[0].Color != KaliGreen || top[1].Text != "kali㉿kali" || top[1].Color != KaliBlue || top[3].Text != "~" || top[3].Color != "" {
		t.Errorf("top line: %+v", top)
	}
	if bottom[0].Text != "└─$" || bottom[0].Color != KaliGreen || bottom[2].Text != "echo" || bottom[2].Color != "#5ebdab" {
		t.Errorf("bottom line: %+v", bottom)
	}

	// Identity from --prompt, and options stay plain in the Kali style.
	root := Colorize("terminal", []string{"$ ls -la /etc"}, Options{Theme: th, PromptStyle: PromptStyleKali, Prompt: "root@box:/etc#"})
	if root[0].Text() != "┌──(root㉿box)-[/etc]" || root[1].Text() != "└─# ls -la /etc" {
		t.Errorf("root identity: %q %q", root[0].Text(), root[1].Text())
	}
	for _, sp := range root[1] {
		if sp.Text == "-la" && sp.Color != "" {
			t.Errorf("options should be plain in the Kali style: %+v", sp)
		}
	}

	// Identity from the content itself, and already-Kali lines pass through.
	own := Colorize("terminal", []string{
		"mira@nimbus:~/api$ make test",
		"┌──(kali㉿kali)-[~/src]",
		"└─$ id",
	}, Options{Theme: th, PromptStyle: PromptStyleKali})
	if own[0].Text() != "┌──(mira㉿nimbus)-[~/api]" || own[1].Text() != "└─$ make test" || len(own) != 4 ||
		own[2].Text() != "┌──(kali㉿kali)-[~/src]" || own[3].Text() != "└─$ id" {
		t.Errorf("content identity / passthrough: %q", func() []string {
			var s []string
			for _, l := range own {
				s = append(s, l.Text())
			}
			return s
		}())
	}
	if id, ok := ParseIdentity("[user@host dir]$"); !ok || id.User != "user" || id.Host != "host" || id.Path != "dir" || id.Symbol != "$" {
		t.Errorf("bracket identity: %+v", id)
	}
	if id, ok := ParseIdentity("nope"); ok || id != DefaultIdentity {
		t.Errorf("bad identity: %+v %v", id, ok)
	}
}
