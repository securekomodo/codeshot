package sniff

import "testing"

func TestPreset(t *testing.T) {
	cases := map[string]string{
		"diff --git a/x b/x\nindex 1..2\n--- a/x\n+++ b/x\n@@ -1 +1 @@\n-a\n+b": "git-diff",
		"commit 8c41f2e0aa\nAuthor: A <a@example.com>\nDate:   Mon\n\n    fix":  "git-commit",
		"POST /api/things HTTP/1.1\nHost: api.example.com\n":                    "http-request",
		"$ git status\nOn branch main\nnothing to commit":                       "terminal",
		"❯ ls\nREADME.md": "terminal",
		// A pasted session keeps whatever prompt it was captured with.
		"tunnel@box2:~$ ssh -p 2222 user@127.0.0.1\nssh: connect to host 127.0.0.1 port 2222: Connection refused": "terminal",
		"[root@box /etc]# ls -la\ntotal 4":                                                          "terminal",
		`PS C:\Windows\system32> Get-Process` + "\nId   CPU":                                        "terminal",
		`C:\Users\dev> dir` + "\n Volume in drive C":                                                "terminal",
		"┌──(kali㉿kali)-[~]\n└─$ id\nuid=0(root)":                                                   "terminal",
		`{"id": 1, "tags": ["a"]}`:                                                                  "api",
		"[1, 2, 3]":                                                                                 "api",
		"app/\n├── lib/\n│   └── x.ts\n└── page.tsx":                                                "project-structure",
		"# db\nDATABASE_URL=postgres://x\nREDIS_URL=redis://y\n\nDEBUG=true":                        "env-vars",
		"PASS app/x.test.ts\n  ✓ works (2 ms)\n  ✓ also (1 ms)\nTests: 2 passed":                    "test-results",
		"2026-06-25 10:04:11  INFO   listening\n2026-06-25 10:04:12  WARN   slow\n10:04:13 ERROR x": "error-log",
		"func main() {\n\tfmt.Println(1)\n}":                                                        "",
		"const x = 1;\n// $ not a prompt":                                                           "",
		"":                                                                                          "",
		"SELECT * FROM users WHERE id = 1":                                                          "",
	}
	for in, want := range cases {
		if got := Preset(in); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
}
