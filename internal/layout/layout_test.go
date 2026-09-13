package layout

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/securekomodo/codeshot/internal/fonts"
	"github.com/securekomodo/codeshot/internal/highlight"
	"github.com/securekomodo/codeshot/internal/preset"
	"github.com/securekomodo/codeshot/internal/settings"
	"github.com/securekomodo/codeshot/internal/theme"
)

func TestPrepareText(t *testing.T) {
	got := PrepareText("a\tb\r\n\t\tc\n\n")
	want := []string{"a b", "    c", ""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q want %q", got, want)
	}
	if got := PrepareText(""); len(got) != 1 || got[0] != "" {
		t.Errorf("empty input -> %q", got)
	}
}

func TestWrapLine(t *testing.T) {
	l := highlight.Line{{Text: "hello ", Color: "#111111"}, {Text: "world foo "}, {Text: "barbazqux"}}
	rows := wrapLine(l, 10)
	texts := make([]string, len(rows))
	for i, r := range rows {
		texts[i] = r.Text()
	}
	want := []string{"hello ", "world foo ", "barbazqux"}
	if !reflect.DeepEqual(texts, want) {
		t.Errorf("rows = %q want %q", texts, want)
	}
	if rows[0][0].Color != "#111111" || len(rows[1]) != 1 {
		t.Errorf("spans not preserved: %+v", rows)
	}
	long := highlight.Line{{Text: strings.Repeat("x", 25)}}
	if r := wrapLine(long, 10); len(r) != 3 || r[2].Text() != "xxxxx" {
		t.Errorf("hard break: %+v", r)
	}
	all := wrapAll([]highlight.Line{l, {{Text: "short"}}}, 10)
	if len(all) != 4 || all[0].number != 1 || all[1].number != 0 || all[3].number != 2 {
		t.Errorf("wrapAll numbering: %+v", all)
	}
}

func fixture(t *testing.T, key string, lines ...string) (Input, settings.Settings) {
	t.Helper()
	p, _ := preset.Get(key)
	s := settings.Defaults(p)
	code, err := fonts.Load("jetbrains")
	if err != nil {
		t.Fatal(err)
	}
	inter, _ := fonts.Inter()
	th, _ := theme.Get(s.Theme)
	bd, _ := theme.GetBackdrop(s.Backdrop)
	var hl []highlight.Line
	for _, l := range lines {
		hl = append(hl, highlight.Line{{Text: l}})
	}
	return Input{Settings: s, Theme: th, Backdrop: bd, Code: code, Title: inter, Lines: hl}, s
}

func TestComputeCode(t *testing.T) {
	in, _ := fixture(t, "code", "const a = 1;", "", "x")
	L, err := Compute(in)
	if err != nil {
		t.Fatal(err)
	}
	cell := in.Code.Advance('0', 15)
	if L.FontSize != 15 || L.LineHeight != 24 {
		t.Errorf("font %v lh %v", L.FontSize, L.LineHeight)
	}
	if L.Chrome == nil || L.Chrome.Bar.H != BarHeight || L.Chrome.Badge != nil {
		t.Fatalf("chrome = %+v", L.Chrome)
	}
	// Short code: the card is widened so the centered title has room.
	titleW := in.Title.Width("snippet.js", TitleSize)
	if wantW := 2*TitleInset + titleW; L.Card.W < wantW || L.Card.W > wantW+1 || L.Chrome.Title.Text != "snippet.js" {
		t.Errorf("card width %v want ~%v, title %q", L.Card.W, wantW, L.Chrome.Title.Text)
	}
	// Long code: the card fits the content.
	long := strings.Repeat("x", 60)
	in.Lines = []highlight.Line{{{Text: long}}}
	if L2, _ := Compute(in); L2.Card.W != math.Ceil(2*CodePadX+cell+GutterPad+60*cell) {
		t.Errorf("long card width %v", L2.Card.W)
	}
	if L.Card.H != BarHeight+CodePadTop+3*24+CodePadBottom {
		t.Errorf("card height %v", L.Card.H)
	}
	if L.W != L.Card.W+96 || L.H != L.Card.H+96 || L.Card.X != 48 {
		t.Errorf("image %vx%v card %+v", L.W, L.H, L.Card)
	}
	if len(L.Gutter) != 3 || L.Gutter[2].Text != "3" || L.Gutter[0].Anchor != "end" {
		t.Errorf("gutter = %+v", L.Gutter)
	}
	if L.Rows[1].Y-L.Rows[0].Y != 24 || L.Rows[0].X != 48+CodePadX+cell+GutterPad {
		t.Errorf("rows = %+v", L.Rows[:2])
	}
	if L.Chrome.Dots[1].CX != 48+16+6+20 || L.Chrome.Title.Text != "snippet.js" || L.Chrome.Title.Anchor != "middle" {
		t.Errorf("chrome = %+v", L.Chrome)
	}
	if !L.ShowBackground || !L.Backdrop.Gradient() || L.Window.Hex != "#282a36" {
		t.Errorf("backdrop/window: %+v %+v", L.Backdrop, L.Window)
	}
}

func TestComputeBadgeWidthMilestone(t *testing.T) {
	code, _ := fonts.Load("jetbrains")
	cell := code.Advance('0', 15)
	asc, desc := code.Metrics(15)
	in, _ := fixture(t, "api", `{"a": 1}`)
	L, _ := Compute(in)
	if L.Chrome.Badge == nil || L.Chrome.Bar.H != BarWithBadge || len(L.Chrome.Badge.Texts) != 2 ||
		L.Chrome.Badge.Texts[1].Text != "200 OK" || L.Chrome.Badge.Texts[0].Color.Hex != highlight.Teal {
		t.Errorf("api badge = %+v", L.Chrome.Badge)
	}
	if right := L.Chrome.Badge.Box.X + L.Chrome.Badge.Box.W; right != L.Card.X+L.Card.W-BarPadX {
		t.Errorf("badge right edge %v", right)
	}

	in, s := fixture(t, "code", "short")
	s.Width = 768
	in.Settings = s
	L, _ = Compute(in)
	if L.Card.W != 768 || L.W != 768+96 {
		t.Errorf("fixed width: %v", L.Card.W)
	}
	s.Width, s.Wrap = 0, 3
	in.Settings = s
	in.Lines = []highlight.Line{{{Text: "ab cd ef"}}}
	L, _ = Compute(in)
	if len(L.Rows) != 3 || len(L.Gutter) != 1 {
		t.Errorf("wrap rows=%d gutter=%d", len(L.Rows), len(L.Gutter))
	}

	// A very long line wraps at the default max width instead of widening the card.
	in, s = fixture(t, "terminal", "$ curl https://example.com/"+strings.Repeat("abcdefghij", 30))
	L, _ = Compute(in)
	if L.Card.W > settings.DefaultMaxWidth || len(L.Rows) < 3 {
		t.Errorf("default max width: card %v rows %d", L.Card.W, len(L.Rows))
	}
	s.MaxWidth = 0
	in.Settings = s
	L, _ = Compute(in)
	if L.Card.W < 2000 || len(L.Rows) != 1 {
		t.Errorf("unlimited: card %v rows %d", L.Card.W, len(L.Rows))
	}
	s.MaxWidth, s.Width = 0, 500
	in.Settings = s
	L, _ = Compute(in)
	if L.Card.W != 500 || len(L.Rows) < 5 {
		t.Errorf("fixed width wins: card %v rows %d", L.Card.W, len(L.Rows))
	}

	in, s = fixture(t, "code", "x")
	s.Label = "JavaScript"
	in.Settings = s
	L, _ = Compute(in)
	top := math.Max(48, labelBand(float64(s.LabelSize)))
	if L.Label == nil || L.Card.Y != top || L.H != L.Card.H+48+top ||
		L.Label.Text.Text != "JavaScript" || L.Label.Box.X+L.Label.Box.W/2 != L.Card.X+L.Card.W/2 {
		t.Errorf("label layout: card %+v label %+v", L.Card, L.Label)
	}
	if L.Label.Box.Y+L.Label.Box.H > L.Card.Y {
		t.Error("label overlaps the card")
	}
	small := L.Label.Box
	s.LabelSize = 28
	in.Settings = s
	L, _ = Compute(in)
	if L.Label.Box.H <= small.H || L.Label.Box.W <= small.W || L.Card.Y <= top {
		t.Errorf("bigger label size should grow the pill and band: %+v", L.Label.Box)
	}

	in, s = fixture(t, "terminal", "$ ls", "a", "$")
	s.Chrome, s.Cursor, s.Title = settings.ChromeKali, true, "kali@kali: ~"
	in.Settings = s
	L, _ = Compute(in)
	c := L.Chrome
	if c == nil || c.Style != settings.ChromeKali || c.Bar.H != KaliTitleHeight+KaliMenuHeight || len(c.Buttons) != 3 || len(c.Menu) != 5 || c.Icon == nil {
		t.Fatalf("kali chrome: %+v", c)
	}
	if c.Buttons[2].Kind != "close" || c.Buttons[2].CX != L.Card.X+L.Card.W-KaliButtonInset || c.Buttons[0].CX >= c.Buttons[1].CX {
		t.Errorf("kali buttons: %+v", c.Buttons)
	}
	if c.Title.Text != "kali@kali: ~" || c.Title.Anchor != "middle" || c.Menu[0].Text != "File" || c.Menu[4].Text != "Help" {
		t.Errorf("kali title/menu: %+v %+v", c.Title, c.Menu)
	}
	if L.Cursor == nil || L.Cursor.Y >= L.Rows[2].Y || L.Cursor.X <= L.Rows[2].X || L.Cursor.W != cell {
		t.Errorf("cursor: %+v row %+v", L.Cursor, L.Rows[2])
	}
	if L.LineHeight != 18 || L.Rows[0].Y != L.Card.Y+KaliTitleHeight+KaliMenuHeight+CodePadTop+(18-(asc+desc))/2+asc {
		t.Errorf("kali line height %v, first row %v", L.LineHeight, L.Rows[0].Y)
	}

	in, _ = fixture(t, "dev-milestone", "🎉 done")
	L, _ = Compute(in)
	if L.Chrome != nil || L.FontSize != 26 || L.LineHeight != 39 || !L.Center || L.Card.X != 64 || len(L.Gutter) != 0 {
		t.Errorf("milestone: %+v", L)
	}
	if L.Rows[0].X != L.Card.X+L.Card.W/2 {
		t.Errorf("milestone row not centered")
	}

	in, s = fixture(t, "regex", "a|b")
	s.Title = strings.Repeat("long-title-", 20)
	in.Settings = s
	L, _ = Compute(in)
	if !strings.HasSuffix(L.Chrome.Title.Text, "…") || L.Chrome.Badge.Texts[0].Text != "/gi" {
		t.Errorf("title %q badge %+v", L.Chrome.Title.Text, L.Chrome.Badge.Texts)
	}
	s.ShowChrome = false
	in.Settings = s
	L, _ = Compute(in)
	if L.Chrome != nil || L.Card.H != CodePadTop+24+CodePadBottom {
		t.Errorf("no chrome: %+v", L.Card)
	}
}
