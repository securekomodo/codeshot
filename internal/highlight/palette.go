package highlight

import "github.com/securekomodo/codeshot/internal/theme"

// Palette is the set of colors the line colorizers may use: the theme's
// own token colors where the theme defines them, and the accent palette
// otherwise, so terminal sessions and logs match the code cards next to them.
type Palette struct {
	Light bool
	Dim   Span // secondary text (timestamps, script echo, decorations)

	String, Keyword, Function, Number, Operator, Comment, Builtin, Variable, URL Span
}

// NewPalette derives a palette from a theme.
func NewPalette(th *theme.Theme) Palette {
	light := th.Light()
	dimColor, dimOp := Dim(light)
	r := th.Resolve("bash")
	pick := func(fallback string, types ...string) Span {
		sp := styled("", r.StyleFor(types...))
		if sp.Color == "" {
			sp.Color, sp.Opacity = fallback, 0
		}
		sp.Italic, sp.Bold, sp.Underline, sp.Strike = false, false, false, false
		return sp
	}
	return Palette{
		Light:    light,
		Dim:      Span{Color: dimColor, Opacity: dimOp},
		String:   pick(Amber, "string"),
		Keyword:  pick(Purple, "keyword"),
		Function: pick(Teal, "function"),
		Number:   pick(Amber, "number"),
		Operator: pick(Grey, "operator"),
		Comment:  pick(Grey, "comment"),
		Builtin:  pick(Purple, "builtin"),
		Variable: pick(Amber, "variable"),
		URL:      pick(Teal, "url", "string"),
	}
}

func (p Palette) span(base Span, text string) Span {
	base.Text = text
	return base
}
