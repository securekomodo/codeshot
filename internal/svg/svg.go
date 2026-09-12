// Package svg serializes a positioned layout as an SVG document that both
// resvg and browsers render the same way: gradients and shadows as filters,
// the rounded card as a clip path, one <text> per row with a <tspan> per
// styled span, and every color as hex plus opacity (resvg rejects rgba()).
package svg

import (
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"strings"

	"codeshot/internal/fonts"
	"codeshot/internal/highlight"
	"codeshot/internal/layout"
	"codeshot/internal/theme"
)

// Options control serialization.
type Options struct {
	// Scale multiplies the declared width and height (the viewBox stays at
	// 1x), so resvg rasterizes at that pixel ratio. 0 means 1.
	Scale float64
	// EmbedFonts are written as @font-face data URIs so a standalone .svg
	// renders with the right fonts. Leave empty for the PNG path, where the
	// rasterizer loads the font data directly.
	EmbedFonts []*fonts.Face
	// FastShadow draws the drop shadow as a stack of translucent rounded
	// rectangles that follow the Gaussian falloff instead of a blur filter.
	// Visually equivalent, and it removes the one expensive step for the
	// rasterizer (a 60px blur triples the render time).
	FastShadow bool
}

// Render serializes the layout.
func Render(L *layout.Layout, o Options) []byte {
	scale := o.Scale
	if scale <= 0 {
		scale = 1
	}
	w := &writer{}
	w.printf(`<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s">`,
		num(math.Round(L.W*scale)), num(math.Round(L.H*scale)), num(L.W), num(L.H))
	w.defs(L, o)
	w.backdrop(L)
	// With rounded backdrop corners, nothing may spill into the cut-off corners.
	clipped := L.ShowBackground && L.Radius > 0
	if clipped {
		w.printf(`<g clip-path="url(#backdrop)">`)
	}
	if o.FastShadow {
		w.layeredShadow(L)
	} else {
		w.shadow(L)
	}
	// A 1px hairline that stays even when the shadow is off.
	w.printf(`<rect x="%s" y="%s" width="%s" height="%s" rx="%s" fill="#ffffff" fill-opacity="0.04"/>`,
		num(L.Card.X-1), num(L.Card.Y-1), num(L.Card.W+2), num(L.Card.H+2), num(L.CardRadius+1))
	w.printf(`<g clip-path="url(#card)">`)
	w.printf(`<rect x="%s" y="%s" width="%s" height="%s"%s/>`,
		num(L.Card.X), num(L.Card.Y), num(L.Card.W), num(L.Card.H), fill(L.Window))
	if L.Chrome != nil {
		w.chrome(L.Chrome)
	}
	for _, g := range L.Gutter {
		w.text(g)
	}
	for _, r := range L.Rows {
		w.row(L, r)
	}
	w.printf(`</g>`)
	if clipped {
		w.printf(`</g>`)
	}
	w.printf(`</svg>`)
	return []byte(w.String())
}

type writer struct{ strings.Builder }

func (w *writer) printf(format string, args ...any) {
	fmt.Fprintf(w, format, args...)
	w.WriteByte('\n')
}

func (w *writer) defs(L *layout.Layout, o Options) {
	embed := o.EmbedFonts
	w.printf(`<defs>`)
	if L.ShowBackground && L.Backdrop.Gradient() {
		x1, y1, x2, y2 := GradientVector(L.Backdrop.Angle, L.W, L.H)
		w.printf(`<linearGradient id="bg" gradientUnits="userSpaceOnUse" x1="%s" y1="%s" x2="%s" y2="%s">`,
			num(x1), num(y1), num(x2), num(y2))
		for _, s := range L.Backdrop.Stops {
			w.printf(`<stop offset="%s" stop-color="%s"%s/>`, num(s.Offset), s.Color.Hex, opacityAttr("stop-opacity", s.Color.Alpha))
		}
		w.printf(`</linearGradient>`)
	}
	w.printf(`<clipPath id="card"><rect x="%s" y="%s" width="%s" height="%s" rx="%s"/></clipPath>`,
		num(L.Card.X), num(L.Card.Y), num(L.Card.W), num(L.Card.H), num(L.CardRadius))
	if L.ShowBackground && L.Radius > 0 {
		w.printf(`<clipPath id="backdrop"><rect width="%s" height="%s" rx="%s"/></clipPath>`, num(L.W), num(L.H), num(L.Radius))
	}
	if L.Shadow && !o.FastShadow {
		// CSS 0 24px 60px -12px rgba(0,0,0,.55): a 60px blur is a Gaussian with
		// sigma 30. The filter region is the shadow box plus three sigmas,
		// clipped to the image, so the blur is not computed over the whole card.
		const margin = 3 * shadowSigma
		x := math.Max(0, L.Card.X+shadowSpread-margin)
		y := math.Max(0, L.Card.Y+shadowSpread+shadowOffsetY-margin)
		x2 := math.Min(L.W, L.Card.X+L.Card.W-shadowSpread+margin)
		y2 := math.Min(L.H, L.Card.Y+L.Card.H-shadowSpread+shadowOffsetY+margin)
		w.printf(`<filter id="shadow" filterUnits="userSpaceOnUse" x="%s" y="%s" width="%s" height="%s"><feGaussianBlur stdDeviation="%s"/></filter>`,
			num(x), num(y), num(x2-x), num(y2-y), num(shadowSigma))
	}
	if len(embed) > 0 {
		w.printf(`<style>`)
		for _, f := range embed {
			weight := 400
			if f.ID == "inter" {
				weight = 500
			}
			w.printf(`@font-face{font-family:%s;font-weight:%d;src:url(data:font/ttf;base64,%s) format('truetype');}`,
				cssFamily(f.Family), weight, base64.StdEncoding.EncodeToString(f.Data))
		}
		w.printf(`</style>`)
	}
	w.printf(`</defs>`)
}

func (w *writer) backdrop(L *layout.Layout) {
	if !L.ShowBackground {
		return
	}
	rx := ""
	if L.Radius > 0 {
		rx = ` rx="` + num(L.Radius) + `"`
	}
	switch {
	case L.Backdrop.Gradient():
		w.printf(`<rect width="%s" height="%s"%s fill="url(#bg)"/>`, num(L.W), num(L.H), rx)
	default:
		w.printf(`<rect width="%s" height="%s"%s%s/>`, num(L.W), num(L.H), rx, fill(L.Backdrop.Solid))
	}
}

// The card's CSS box-shadow: 0 24px 60px -12px rgba(0,0,0,.55).
const (
	shadowOffsetY = 24
	shadowSigma   = 30 // blur radius 60px
	shadowSpread  = 12 // negative spread shrinks the shadow box
)

func (w *writer) shadow(L *layout.Layout) {
	if !L.Shadow {
		return
	}
	w.printf(`<rect x="%s" y="%s" width="%s" height="%s" rx="%s" fill="#000000" fill-opacity="%s" filter="url(#shadow)"/>`,
		num(L.Card.X+shadowSpread), num(L.Card.Y+shadowSpread+shadowOffsetY),
		num(L.Card.W-2*shadowSpread), num(L.Card.H-2*shadowSpread), num(shadowRadius(L)), num(shadowAlpha))
}

const shadowAlpha = 0.55

// shadowRadius is the corner radius of the shadow box: the window's radius
// shrunk by the negative spread.
func shadowRadius(L *layout.Layout) float64 { return math.Max(0, L.CardRadius-shadowSpread) }

// layeredShadow approximates the blurred shadow box with concentric bands.
// A Gaussian blur of a box has coverage A(d) = alpha * Q(d/sigma) at signed
// distance d from the edge (Q the upper normal tail), so each band between
// two offsets is filled with A at its midpoint. Bands are even-odd ring
// paths, so each pixel is painted once; outward bands are rounded by their
// offset, exactly as the blurred corner would be.
func (w *writer) layeredShadow(L *layout.Layout) {
	if !L.Shadow {
		return
	}
	bx, by := L.Card.X+shadowSpread, L.Card.Y+shadowSpread+shadowOffsetY
	bw, bh := L.Card.W-2*shadowSpread, L.Card.H-2*shadowSpread
	const outer, inner = 3.0, 2.0 // sigmas covered outside and inside the box
	const step = 0.125            // sigmas per band
	coverage := func(d float64) float64 { return shadowAlpha * 0.5 * math.Erfc(d/shadowSigma/math.Sqrt2) }
	for e := outer * shadowSigma; e > -inner*shadowSigma+1e-9; e -= step * shadowSigma {
		in := e - step*shadowSigma
		alpha := coverage((e + in) / 2)
		if alpha < 0.0005 || bw+2*in <= 0 || bh+2*in <= 0 {
			continue
		}
		r0 := shadowRadius(L)
		w.printf(`<path d="%s %s" fill-rule="evenodd" fill="#000000" fill-opacity="%s"/>`,
			roundedRectPath(bx-e, by-e, bw+2*e, bh+2*e, math.Max(0, r0+e)),
			roundedRectPath(bx-in, by-in, bw+2*in, bh+2*in, math.Max(0, r0+in)), num(alpha))
	}
	// The solid core, hidden under the card except at the bottom edge.
	core := inner * shadowSigma
	if bw-2*core > 0 && bh-2*core > 0 {
		w.printf(`<rect x="%s" y="%s" width="%s" height="%s" fill="#000000" fill-opacity="%s"/>`,
			num(bx+core), num(by+core), num(bw-2*core), num(bh-2*core), num(coverage(-core)))
	}
}

// roundedRectPath returns path data for a rectangle with corner radius r.
func roundedRectPath(x, y, w, h, r float64) string {
	r = math.Min(r, math.Min(w, h)/2)
	if r <= 0 {
		return fmt.Sprintf("M%s,%s h%s v%s h%s z", num(x), num(y), num(w), num(h), num(-w))
	}
	a := func(dx, dy float64) string {
		return fmt.Sprintf("a%s,%s 0 0 1 %s,%s", num(r), num(r), num(dx), num(dy))
	}
	return fmt.Sprintf("M%s,%s h%s %s v%s %s h%s %s v%s %s z",
		num(x+r), num(y), num(w-2*r), a(r, r), num(h-2*r), a(-r, r), num(-(w - 2*r)), a(-r, -r), num(-(h - 2*r)), a(r, -r))
}

func (w *writer) chrome(c *layout.Chrome) {
	for _, d := range c.Dots {
		w.printf(`<circle cx="%s" cy="%s" r="%s" fill="%s"/>`, num(d.CX), num(d.CY), num(layout.DotRadius), d.Color)
		w.printf(`<circle cx="%s" cy="%s" r="%s" fill="none" stroke="#000000" stroke-opacity="0.18" stroke-width="0.5"/>`,
			num(d.CX), num(d.CY), num(layout.DotRadius-0.25))
	}
	w.text(c.Title)
	if b := c.Badge; b != nil {
		// A CSS border sits inside the box; an SVG stroke straddles the edge.
		w.printf(`<rect x="%s" y="%s" width="%s" height="%s" rx="%s" fill="none" stroke="%s"%s stroke-width="1"/>`,
			num(b.Box.X+0.5), num(b.Box.Y+0.5), num(b.Box.W-1), num(b.Box.H-1), num(layout.BadgeRadius),
			b.Border.Hex, opacityAttr("stroke-opacity", b.Border.Alpha))
		for _, t := range b.Texts {
			w.text(t)
		}
	}
}

func (w *writer) text(t layout.Text) {
	if t.Text == "" {
		return
	}
	weight := ""
	if t.Weight > 0 {
		weight = fmt.Sprintf(` font-weight="%d"`, t.Weight)
	}
	w.printf(`<text x="%s" y="%s" text-anchor="%s" font-family=%s font-size="%s"%s%s xml:space="preserve">%s</text>`,
		num(t.X), num(t.Y), t.Anchor, family(t.Font), num(t.Size), weight, fill(t.Color), escape(t.Text))
}

func (w *writer) row(L *layout.Layout, r layout.Row) {
	if r.Spans.Text() == "" {
		return
	}
	anchor := "start"
	if L.Center {
		anchor = "middle"
	}
	w.printf(`<text x="%s" y="%s" text-anchor="%s" font-family=%s font-size="%s"%s xml:space="preserve">`,
		num(r.X), num(r.Y), anchor, family(L.Font), num(L.FontSize), fill(L.Plain))
	for _, sp := range r.Spans {
		if sp.Text == "" {
			continue
		}
		w.WriteString(`<tspan`)
		if sp.Color != "" {
			w.WriteString(` fill="` + sp.Color + `"`)
		}
		w.WriteString(opacityAttr("fill-opacity", sp.Opacity))
		if sp.Italic {
			w.WriteString(` font-style="italic"`)
		}
		if sp.Bold {
			w.WriteString(` font-weight="600"`)
		}
		if deco := decoration(sp); deco != "" {
			w.WriteString(` text-decoration="` + deco + `"`)
		}
		w.WriteString(`>` + escape(sp.Text) + `</tspan>`)
	}
	w.printf(`</text>`)
}

func decoration(sp highlight.Span) string {
	switch {
	case sp.Underline && sp.Strike:
		return "underline line-through"
	case sp.Underline:
		return "underline"
	case sp.Strike:
		return "line-through"
	}
	return ""
}

// GradientVector converts a CSS linear-gradient angle (0deg = to top,
// clockwise) on a w×h box into SVG userSpaceOnUse endpoints, using the CSS
// "magic corners" gradient-line length so the first and last stops land
// exactly on the box corners.
func GradientVector(angleDeg, w, h float64) (x1, y1, x2, y2 float64) {
	rad := angleDeg * math.Pi / 180
	dx, dy := math.Sin(rad), -math.Cos(rad)
	length := math.Abs(w*math.Sin(rad)) + math.Abs(h*math.Cos(rad))
	cx, cy := w/2, h/2
	return cx - dx*length/2, cy - dy*length/2, cx + dx*length/2, cy + dy*length/2
}

// num formats a coordinate compactly, with at most three decimals.
func num(v float64) string {
	v = math.Round(v*1000) / 1000
	if v == 0 {
		return "0"
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func fill(c theme.Color) string {
	return ` fill="` + c.Hex + `"` + opacityAttr("fill-opacity", c.Alpha)
}

// opacityAttr emits the attribute only when the alpha is meaningful.
func opacityAttr(attr string, alpha float64) string {
	if alpha <= 0 && attr == "fill-opacity" {
		return "" // an unset span opacity; the theme's plain fill is opaque
	}
	if alpha >= 1 || (alpha == 0 && attr != "stop-opacity") {
		return ""
	}
	return ` ` + attr + `="` + num(alpha) + `"`
}

func family(f *fonts.Face) string {
	names := make([]string, len(f.Families))
	for i, n := range f.Families {
		names[i] = "'" + strings.ReplaceAll(n, "'", "") + "'"
	}
	return `"` + strings.Join(names, ", ") + `"`
}

func cssFamily(name string) string { return "'" + strings.ReplaceAll(name, "'", "") + "'" }

var escaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")

// escape makes text safe for XML content, dropping control characters
// XML 1.0 forbids.
func escape(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			return -1
		}
		return r
	}, s)
	return escaper.Replace(s)
}
