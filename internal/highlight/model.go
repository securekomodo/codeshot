// Package highlight turns source text into styled lines, either through the
// chroma syntax highlighter mapped onto the bundled Prism-format themes, or
// through the per-line colorizers used by the non-code presets.
package highlight

// Span is a run of text with a resolved style.
type Span struct {
	Text      string
	Color     string  // "#rrggbb"; "" means the theme's plain text color
	Opacity   float64 // 0 or 1 = opaque; < 1 emits fill-opacity
	Italic    bool
	Bold      bool
	Underline bool
	Strike    bool
}

// Line is one visual row of spans.
type Line []Span

// Text returns the line's raw text.
func (l Line) Text() string {
	var s string
	for _, sp := range l {
		s += sp.Text
	}
	return s
}

// The accent palette shared by the badges and the line colorizers.
const (
	Teal   = "#5fb3a1"
	Amber  = "#f4a259"
	Red    = "#e5705a"
	Purple = "#b08af9"
	Grey   = "#8a929e"
)

// Dim returns the muted color used for secondary text:
// rgba(230,232,235,.55) on dark themes, rgba(0,0,0,.5) on light ones.
func Dim(light bool) (color string, opacity float64) {
	if light {
		return "#000000", 0.5
	}
	return "#e6e8eb", 0.55
}

// Gutter returns the line-number color: rgba(255,255,255,.30) on dark
// themes, rgba(0,0,0,.32) on light ones.
func Gutter(light bool) (color string, opacity float64) {
	if light {
		return "#000000", 0.32
	}
	return "#ffffff", 0.30
}
