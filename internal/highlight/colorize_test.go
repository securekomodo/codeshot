package highlight

import (
	"reflect"
	"testing"
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
		{"terminal", "$ npm test", Line{c("$", Teal), p(" npm test")}},
		{"terminal", "❯ ls", Line{c("❯", Teal), p(" ls")}},
		{"terminal", "$npm", Line{d("$npm")}},
		{"terminal", "output line", Line{d("output line")}},
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
		{"tree", "codeshot/", Line{d(""), c("codeshot/", Amber)}},
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
	for _, tc := range cases {
		got := Colorize(tc.mode, []string{tc.in}, "", false)[0]
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s %q:\n got %+v\nwant %+v", tc.mode, tc.in, got, tc.want)
		}
	}
}

func TestTerminalPromptAndLight(t *testing.T) {
	got := Colorize("terminal", []string{"$ ls", "out"}, "❯", true)
	if got[0][0].Text != "❯" || got[0][0].Color != Teal || got[0][1].Text != " ls" {
		t.Errorf("prompt replacement: %+v", got[0])
	}
	if got[1][0].Color != "#000000" || got[1][0].Opacity != 0.5 {
		t.Errorf("light dim: %+v", got[1])
	}
}
