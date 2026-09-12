package svg

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"codeshot/internal/fonts"
	"codeshot/internal/highlight"
	"codeshot/internal/layout"
	"codeshot/internal/preset"
	"codeshot/internal/settings"
	"codeshot/internal/theme"
)

func TestGradientVector(t *testing.T) {
	cases := []struct {
		angle, w, h    float64
		x1, y1, x2, y2 float64
	}{
		{90, 800, 600, 0, 300, 800, 300},
		{0, 800, 600, 400, 600, 400, 0},
		{180, 800, 600, 400, 0, 400, 600},
		{135, 800, 600, 50, -50, 750, 650},
	}
	for _, c := range cases {
		x1, y1, x2, y2 := GradientVector(c.angle, c.w, c.h)
		for i, p := range [][2]float64{{x1, c.x1}, {y1, c.y1}, {x2, c.x2}, {y2, c.y2}} {
			if math.Abs(p[0]-p[1]) > 1e-6 {
				t.Errorf("angle %v: coord %d = %v want %v", c.angle, i, p[0], p[1])
			}
		}
	}
}

func render(t *testing.T, key string, o Options, lines ...highlight.Line) string {
	t.Helper()
	p, _ := preset.Get(key)
	s := settings.Defaults(p)
	code, _ := fonts.Load("jetbrains")
	inter, _ := fonts.Inter()
	th, _ := theme.Get(s.Theme)
	bd, _ := theme.GetBackdrop(s.Backdrop)
	L, err := layout.Compute(layout.Input{Settings: s, Theme: th, Backdrop: bd, Code: code, Title: inter, Lines: lines})
	if err != nil {
		t.Fatal(err)
	}
	return string(Render(L, o))
}

func TestRender(t *testing.T) {
	out := render(t, "code", Options{}, highlight.Line{
		{Text: "a < b && c", Color: "#ff79c6"},
		{Text: "  x", Opacity: 0.55, Italic: true, Underline: true},
	})
	for _, want := range []string{
		`viewBox="0 0 `, `<linearGradient id="bg"`, `stop-color="#c75c3c"`, `offset="0.5"`,
		`<clipPath id="card">`, `<filter id="shadow"`, `filter="url(#shadow)"`, `rx="12"`,
		`fill="#282a36"`, `<circle`, `fill="#28c840"`, `text-anchor="middle"`, `font-weight="500"`,
		`snippet.js`, `text-anchor="end"`, `>1</text>`, `xml:space="preserve"`,
		`<tspan fill="#ff79c6">a &lt; b &amp;&amp; c</tspan>`,
		`<tspan fill-opacity="0.55" font-style="italic" text-decoration="underline">  x</tspan>`,
		`font-family="'JetBrains Mono'"`, `font-family="'Inter'`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "@font-face") || strings.Contains(out, "rgba(") {
		t.Error("unexpected @font-face or rgba()")
	}

	scaled := render(t, "api", Options{Scale: 2}, highlight.Line{{Text: "{}"}})
	if !strings.Contains(scaled, `<rect x=`) || !strings.Contains(scaled, `stroke-opacity="0.12"`) || !strings.Contains(scaled, `200 OK`) {
		t.Errorf("badge missing:\n%s", scaled)
	}
	head := scaled[:strings.Index(scaled, ">")]
	if !strings.Contains(head, `width="`) || strings.Contains(head, `width="0"`) {
		t.Errorf("bad header %s", head)
	}

	code, _ := fonts.Load("jetbrains")
	embedded := render(t, "code", Options{EmbedFonts: []*fonts.Face{code}}, highlight.Line{{Text: "x"}})
	if !strings.Contains(embedded, "@font-face{font-family:'JetBrains Mono';font-weight:400;src:url(data:font/ttf;base64,") {
		t.Error("font not embedded")
	}
}

func TestFastShadow(t *testing.T) {
	out := render(t, "code", Options{FastShadow: true}, highlight.Line{{Text: "x"}})
	if strings.Contains(out, "<filter") || strings.Contains(out, "filter=") {
		t.Error("fast shadow should not use a filter")
	}
	bands := strings.Count(out, `fill-rule="evenodd"`)
	if bands < 30 || bands > 41 {
		t.Errorf("bands = %d", bands)
	}
	// Opacities rise from the outermost band inward and stay under the CSS alpha.
	last := -1.0
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, `fill-rule="evenodd"`) {
			continue
		}
		i := strings.Index(line, `fill-opacity="`) + len(`fill-opacity="`)
		v, _ := strconv.ParseFloat(line[i:strings.Index(line[i:], `"`)+i], 64)
		if v < last || v > 0.55 {
			t.Errorf("band opacity %v after %v", v, last)
		}
		last = v
	}
	if p := roundedRectPath(0, 0, 10, 10, 0); p != "M0,0 h10 v10 h-10 z" {
		t.Errorf("sharp path %q", p)
	}
	if p := roundedRectPath(0, 0, 10, 10, 2); !strings.HasPrefix(p, "M2,0 h6 a2,2 0 0 1 2,2 v6") {
		t.Errorf("rounded path %q", p)
	}
}

func TestKaliChromeAndCursor(t *testing.T) {
	p, _ := preset.Get("terminal")
	s := settings.Defaults(p)
	s.Chrome, s.Cursor, s.Title = settings.ChromeKali, true, "kali@kali: ~"
	code, _ := fonts.Load("jetbrains")
	inter, _ := fonts.Inter()
	th, _ := theme.Get(s.Theme)
	bd, _ := theme.GetBackdrop(s.Backdrop)
	L, _ := layout.Compute(layout.Input{Settings: s, Theme: th, Backdrop: bd, Code: code, Title: inter, Lines: []highlight.Line{{{Text: "$"}}}})
	out := string(Render(L, Options{}))
	for _, want := range []string{`fill="#ffffff" fill-opacity="0.07"/>`, `fill="` + layout.KaliBlue + `"`, `stroke="` + layout.KaliButtonEdge + `" stroke-width="2"`,
		`fill="` + layout.KaliButton + `" stroke="` + layout.KaliButtonEdge + `" stroke-width="1"/>`, `rx="1.5" fill="none"`, `>Actions</text>`, `>kali@kali: ~</text>`, `fill-opacity="0.95"/>`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(out, `fill="#ff5f57"`) {
		t.Error("traffic lights should not be drawn in the kali style")
	}
	if strings.Contains(out, `fill="`+layout.KaliBar+`"`) {
		t.Error("kali bars share the window color")
	}
	if strings.Contains(out, `fill-opacity="0.11"`) {
		t.Error("no swirl unless asked")
	}
	s.Watermark = settings.WatermarkSwirl
	L, _ = layout.Compute(layout.Input{Settings: s, Theme: th, Backdrop: bd, Code: code, Title: inter, Lines: []highlight.Line{{{Text: "$"}}}})
	out = string(Render(L, Options{}))
	if strings.Count(out, `fill="#000000" fill-opacity="0.11" stroke="#000000"`) != 3 {
		t.Error("swirl should place the three traced strokes")
	}
	if strings.Index(out, `fill-opacity="0.11"`) > strings.Index(out, `>Actions</text>`) {
		t.Error("swirl should be drawn under the menu text")
	}

	png1x1 := []byte("\x89PNG\r\n\x1a\n")
	L, _ = layout.Compute(layout.Input{Settings: s, Theme: th, Backdrop: bd, Code: code, Title: inter, Lines: []highlight.Line{{{Text: "$"}}},
		Image: &layout.WatermarkImage{Data: png1x1, Mime: "image/png", W: 2, H: 1, Opacity: 0.2}})
	out = string(Render(L, Options{}))
	if !strings.Contains(out, `<image x=`) || !strings.Contains(out, `opacity="0.2"`) || !strings.Contains(out, `href="data:image/png;base64,`) {
		t.Errorf("image watermark missing: %s", out[strings.Index(out, "<g clip-path"):][:200])
	}
	if L.Watermark.Box.W != 2*L.Watermark.Box.H {
		t.Errorf("watermark aspect: %+v", L.Watermark.Box)
	}
}

func TestRadii(t *testing.T) {
	p, _ := preset.Get("code")
	s := settings.Defaults(p)
	s.Radius, s.CardRadius = 24, 0
	code, _ := fonts.Load("jetbrains")
	th, _ := theme.Get(s.Theme)
	bd, _ := theme.GetBackdrop(s.Backdrop)
	L, _ := layout.Compute(layout.Input{Settings: s, Theme: th, Backdrop: bd, Code: code, Lines: []highlight.Line{{{Text: "x"}}}})
	out := string(Render(L, Options{FastShadow: true}))
	for _, want := range []string{`rx="24" fill="url(#bg)"`, `<clipPath id="backdrop"><rect width="`, `clip-path="url(#backdrop)"`, `rx="0"/></clipPath>`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
	s.Backdrop = "none"
	bd, _ = theme.GetBackdrop(s.Backdrop)
	L, _ = layout.Compute(layout.Input{Settings: s, Theme: th, Backdrop: bd, Code: code, Lines: []highlight.Line{{{Text: "x"}}}})
	if out := string(Render(L, Options{})); strings.Contains(out, "backdrop") {
		t.Error("transparent backdrop should not be clipped")
	}
}

func TestNoBackgroundNoShadow(t *testing.T) {
	p, _ := preset.Get("dev-milestone")
	s := settings.Defaults(p)
	s.Backdrop, s.Shadow = "none", false
	code, _ := fonts.Load("jetbrains")
	th, _ := theme.Get(s.Theme)
	bd, _ := theme.GetBackdrop(s.Backdrop)
	L, _ := layout.Compute(layout.Input{Settings: s, Theme: th, Backdrop: bd, Code: code, Lines: []highlight.Line{{{Text: "hi"}}}})
	out := string(Render(L, Options{}))
	if strings.Contains(out, "linearGradient") || strings.Contains(out, "filter=") || strings.Contains(out, "<circle") {
		t.Errorf("unexpected backdrop/shadow/chrome:\n%s", out)
	}
	if !strings.Contains(out, `text-anchor="middle"`) || !strings.Contains(out, `font-size="26"`) {
		t.Errorf("milestone text not centered/scaled:\n%s", out)
	}
}
