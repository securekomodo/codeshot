package settings

import (
	"testing"

	"codeshot/internal/preset"
)

func TestDefaults(t *testing.T) {
	code, _ := preset.Get("code")
	s := Defaults(code)
	if s.MaxWidth != DefaultMaxWidth {
		t.Errorf("max width default = %d", s.MaxWidth)
	}
	if s.Theme != "dracula" || s.Backdrop != "ember" || s.Font != "cascadia" || s.Language != "javascript" ||
		s.Title != "snippet.js" || s.Padding != 48 || s.FontSize != 15 || !s.ShowChrome || !s.ShowLineNumbers ||
		!s.Shadow || !s.ShowBackground || s.Prompt != "$" || s.Scale != 2 || s.Flags != "" {
		t.Errorf("code defaults = %+v", s)
	}
	if err := s.Validate(); err != nil {
		t.Error(err)
	}
	ms, _ := preset.Get("dev-milestone")
	if m := Defaults(ms); m.Padding != 64 || m.ShowChrome || m.ShowLineNumbers {
		t.Errorf("milestone defaults = %+v", m)
	}
	rx, _ := preset.Get("regex")
	if r := Defaults(rx); r.Flags != "gi" || r.Language != "regex" || r.Badge() != BadgeRegex {
		t.Errorf("regex defaults = %+v", r)
	}
	term, _ := preset.Get("terminal")
	if x := Defaults(term); x.DisplayTitle() != "Terminal — zsh" || x.ShowLineNumbers || x.Language != "javascript" {
		t.Errorf("terminal defaults = %+v title=%q", x, x.DisplayTitle())
	}
	x := Defaults(term)
	x.Title = "deploy — zsh"
	if x.DisplayTitle() != "deploy — zsh" {
		t.Errorf("suffix doubled: %q", x.DisplayTitle())
	}
	api, _ := preset.Get("api")
	a := Defaults(api)
	if a.Badge() != BadgeAPI {
		t.Error("api badge")
	}
	a.ShowBadge = false
	if a.Badge() != BadgeNone {
		t.Error("api badge off")
	}
}

func TestValidate(t *testing.T) {
	code, _ := preset.Get("code")
	bad := []func(*Settings){
		func(s *Settings) { s.Theme = "x" },
		func(s *Settings) { s.Backdrop = "x" },
		func(s *Settings) { s.Padding = 161 },
		func(s *Settings) { s.FontSize = 10 },
		func(s *Settings) { s.Scale = 0 },
		func(s *Settings) { s.Wrap, s.Width = 80, 768 },
		func(s *Settings) { s.Method = "TRACE" },
		func(s *Settings) { s.Status = "418" },
		func(s *Settings) { s.MaxWidth = 100 },
		func(s *Settings) { s.MaxWidth = -1 },
	}
	for i, f := range bad {
		s := Defaults(code)
		f(&s)
		if err := s.Validate(); err == nil {
			t.Errorf("case %d should fail: %+v", i, s)
		}
	}
}

func TestSlug(t *testing.T) {
	cases := [][3]string{
		{"snippet.js", "code", "snippet-js"},
		{"git log", "git-commit", "git-log"},
		{".env", "env-vars", "env"},
		{"", "api", "api"},
		{"!!!", "", "codeshot"},
		{"Hello World — Ünïcode", "x", "hello-world-n-code"},
		{"aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeee", "x", "aaaaaaaaaabbbbbbbbbbccccccccccdddddddddd"},
	}
	for _, c := range cases {
		if got := Slug(c[0], c[1]); got != c[2] {
			t.Errorf("Slug(%q,%q) = %q want %q", c[0], c[1], got, c[2])
		}
	}
}
