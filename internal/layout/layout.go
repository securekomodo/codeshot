// Package layout positions every element of the card in CSS pixels at 1x:
// backdrop padding, the rounded window with its title bar, gutter and code
// rows.
package layout

import (
	"fmt"
	"math"
	"strconv"

	"codeshot/internal/fonts"
	"codeshot/internal/highlight"
	"codeshot/internal/preset"
	"codeshot/internal/settings"
	"codeshot/internal/theme"
)

// Fixed geometry of the card, in CSS px.
const (
	Radius        = 12   // rounded-xl
	BarHeight     = 36   // px-4 py-3 around 12px dots
	BarWithBadge  = 50.5 // py-3 around a 26.5px badge
	DotRadius     = 6
	DotGap        = 8
	BarPadX       = 16
	TitleInset    = 80  // minimum clearance on each side of the centered title (dots on the left)
	MaxTitleRoom  = 400 // a longer title is truncated rather than widening the card
	TitleSize     = 13
	BadgeSize     = 11
	BadgePadX     = 8
	BadgeHeight   = 26.5
	BadgeRadius   = 6
	BadgeGap      = 8
	CodePadX      = 16
	CodePadTop    = 4
	CodePadBottom = 20
	GutterPad     = 16
	MaxCardWidth  = 4096
	MinCardWidth  = 120
)

// DotColors are the macOS traffic lights.
var DotColors = [3]string{"#ff5f57", "#febc2e", "#28c840"}

// StatusLabels are the badge texts for the api preset's status codes.
var StatusLabels = map[string]string{
	"200": "OK", "201": "Created", "204": "No Content", "301": "Moved",
	"400": "Bad Request", "401": "Unauthorized", "403": "Forbidden",
	"404": "Not Found", "422": "Unprocessable", "500": "Server Error",
}

// MethodColors color the api badge's method.
var MethodColors = map[string]string{
	"GET": highlight.Teal, "POST": highlight.Amber, "PUT": "#6a9bf4",
	"PATCH": highlight.Purple, "DELETE": highlight.Red,
}

// Input is what the layout needs, already resolved by the caller.
type Input struct {
	Settings settings.Settings
	Theme    *theme.Theme
	Backdrop theme.Backdrop
	Code     *fonts.Face      // code area and gutter
	Title    *fonts.Face      // title bar text (Inter); nil falls back to Code
	Badge    *fonts.Face      // badge text (JetBrains Mono); nil falls back to Code
	Lines    []highlight.Line // one per logical line, tabs already expanded
}

// Rect is an axis-aligned box.
type Rect struct{ X, Y, W, H float64 }

// Text is one positioned run of single-style text.
type Text struct {
	X, Y   float64 // baseline origin
	Anchor string  // "start", "middle" or "end"
	Size   float64
	Font   *fonts.Face
	Weight int // 0 = normal
	Text   string
	Color  theme.Color
}

// Row is one visual line of code.
type Row struct {
	X, Y  float64 // baseline origin (X is the center when Layout.Center)
	Spans highlight.Line
}

// Dot is one traffic light.
type Dot struct {
	CX, CY float64
	Color  string
}

// Badge is the pill in the title bar's right slot.
type Badge struct {
	Box    Rect
	Border theme.Color
	Texts  []Text
}

// Chrome is the title bar.
type Chrome struct {
	Bar   Rect
	Dots  [3]Dot
	Title Text
	Badge *Badge
}

// Layout is the fully positioned card at 1x.
type Layout struct {
	W, H           float64
	Backdrop       theme.Backdrop
	ShowBackground bool
	Card           Rect
	Window         theme.Color
	Shadow         bool
	Light          bool
	Chrome         *Chrome // nil when the window chrome is hidden
	Font           *fonts.Face
	FontSize       float64
	LineHeight     float64
	Plain          theme.Color
	Center         bool // milestone: rows are centered
	Gutter         []Text
	Rows           []Row
}

// Compute lays out the card.
func Compute(in Input) (*Layout, error) {
	s := in.Settings
	milestone := s.Preset.Render == preset.Milestone
	size := float64(s.FontSize)
	lh := math.Round(size * 1.6)
	if milestone {
		size = math.Round(1.7 * size)
		lh = math.Round(size * 1.5)
	}
	light := in.Theme.Light()

	plain, err := theme.ParseColor(in.Theme.Plain.Color)
	if err != nil {
		if light {
			plain = theme.Color{Hex: "#1a1a1a", Alpha: 1}
		} else {
			plain = theme.Color{Hex: "#e6e8eb", Alpha: 1}
		}
	}
	window, err := theme.ParseColor(in.Theme.Window)
	if err != nil {
		return nil, fmt.Errorf("theme %s: %w", in.Theme.ID, err)
	}
	titleFont, badgeFont := in.Title, in.Badge
	if titleFont == nil {
		titleFont = in.Code
	}
	if badgeFont == nil {
		badgeFont = in.Code
	}

	cell := in.Code.Advance('0', size)
	showGutter := s.ShowLineNumbers && s.Preset.SupportsLineNumbers() && !milestone
	digits := len(strconv.Itoa(max(1, len(in.Lines))))
	gutterW := 0.0
	if showGutter {
		gutterW = float64(digits)*cell + GutterPad
	}

	var badge *Badge
	if s.ShowChrome {
		badge = buildBadge(s, badgeFont, light)
	}
	badgeW := 0.0
	if badge != nil {
		badgeW = badge.Box.W
	}

	// Columns available at a given card width.
	colsAt := func(width float64) int {
		return max(1, int((width-2*CodePadX-gutterW)/cell))
	}
	cols := 0
	capW := float64(MaxCardWidth)
	switch {
	case s.Wrap > 0:
		cols = s.Wrap
	case s.Width > 0:
		cols = colsAt(float64(s.Width))
	case s.MaxWidth > 0:
		cols = colsAt(float64(s.MaxWidth))
		capW = float64(s.MaxWidth)
	}
	rows := wrapAll(in.Lines, cols)

	maxW := 0.0
	for _, r := range rows {
		w := 0.0
		for _, sp := range r.line {
			w += in.Code.Width(sp.Text, size)
		}
		maxW = math.Max(maxW, w)
	}
	// The title is centered, so it needs the same clearance on both sides:
	// enough for the dots on the left and the badge on the right.
	inset := math.Max(TitleInset, BarPadX+badgeW+BadgeGap)
	cardW := math.Ceil(2*CodePadX + gutterW + maxW)
	if s.Width > 0 {
		cardW = float64(s.Width)
	} else {
		minW := float64(MinCardWidth)
		if s.ShowChrome {
			titleW := math.Min(titleFont.Width(s.DisplayTitle(), TitleSize), MaxTitleRoom)
			minW = math.Ceil(2*inset + titleW)
		}
		cardW = math.Min(math.Max(cardW, minW), capW)
	}

	barH := 0.0
	if s.ShowChrome {
		barH = BarHeight
		if badge != nil {
			barH = BarWithBadge
		}
	}
	// Whole pixels keep the PNG size exact at every scale (the badge bar is 50.5px).
	cardH := math.Ceil(barH + CodePadTop + float64(len(rows))*lh + CodePadBottom)
	pad := float64(s.Padding)

	L := &Layout{
		W: cardW + 2*pad, H: cardH + 2*pad,
		Backdrop: in.Backdrop, ShowBackground: s.ShowBackground && !in.Backdrop.Transparent,
		Card:   Rect{pad, pad, cardW, cardH},
		Window: window, Shadow: s.Shadow, Light: light,
		Font: in.Code, FontSize: size, LineHeight: lh, Plain: plain, Center: milestone,
	}

	if s.ShowChrome {
		L.Chrome = buildChrome(L.Card, barH, s.DisplayTitle(), titleFont, badge, inset, light)
	}

	asc, desc := in.Code.Metrics(size)
	codeTop := L.Card.Y + barH + CodePadTop
	codeX := L.Card.X + CodePadX + gutterW
	gutterHex, gutterOp := highlight.Gutter(light)
	for i, r := range rows {
		y := codeTop + float64(i)*lh + (lh-(asc+desc))/2 + asc
		x := codeX
		if milestone {
			x = L.Card.X + cardW/2
		}
		L.Rows = append(L.Rows, Row{X: x, Y: y, Spans: r.line})
		if showGutter && r.number > 0 {
			L.Gutter = append(L.Gutter, Text{
				X: L.Card.X + CodePadX + float64(digits)*cell, Y: y, Anchor: "end",
				Size: size, Font: in.Code, Text: strconv.Itoa(r.number),
				Color: theme.Color{Hex: gutterHex, Alpha: gutterOp},
			})
		}
	}
	return L, nil
}

func buildBadge(s settings.Settings, font *fonts.Face, light bool) *Badge {
	border := theme.Color{Hex: "#ffffff", Alpha: 0.12}
	if light {
		border = theme.Color{Hex: "#000000", Alpha: 0.1}
	}
	var texts []Text
	switch s.Badge() {
	case settings.BadgeAPI:
		mc, ok := MethodColors[s.Method]
		if !ok {
			mc = highlight.Grey
		}
		code, _ := strconv.Atoi(s.Status)
		sc := highlight.Red
		switch {
		case code < 300:
			sc = highlight.Teal
		case code < 400:
			sc = highlight.Amber
		}
		status := s.Status
		if label, ok := StatusLabels[s.Status]; ok {
			status += " " + label
		}
		texts = []Text{
			{Text: s.Method, Color: theme.Color{Hex: mc, Alpha: 1}},
			{Text: status, Color: theme.Color{Hex: sc, Alpha: 1}},
		}
	case settings.BadgeRegex:
		texts = []Text{{Text: "/" + s.Flags, Color: theme.Color{Hex: highlight.Amber, Alpha: 1}}}
	default:
		return nil
	}
	w := 2.0 + 2*BadgePadX // borders + padding
	for i := range texts {
		texts[i].Size, texts[i].Font, texts[i].Weight, texts[i].Anchor = BadgeSize, font, 500, "start"
		if i > 0 {
			w += BadgeGap
		}
		w += font.Width(texts[i].Text, BadgeSize)
	}
	return &Badge{Box: Rect{W: math.Ceil(w), H: BadgeHeight}, Border: border, Texts: texts}
}

func buildChrome(card Rect, barH float64, title string, font *fonts.Face, badge *Badge, inset float64, light bool) *Chrome {
	c := &Chrome{Bar: Rect{card.X, card.Y, card.W, barH}}
	cy := card.Y + barH/2
	for i := range c.Dots {
		c.Dots[i] = Dot{CX: card.X + BarPadX + DotRadius + float64(i)*(2*DotRadius+DotGap), CY: cy, Color: DotColors[i]}
	}
	asc, desc := font.Metrics(TitleSize)
	color := theme.Color{Hex: "#ffffff", Alpha: 0.45}
	if light {
		color = theme.Color{Hex: "#000000", Alpha: 0.5}
	}
	c.Title = Text{
		X: card.X + card.W/2, Y: cy + (asc-desc)/2, Anchor: "middle",
		Size: TitleSize, Font: font, Weight: 500,
		Text: truncate(title, font, TitleSize, card.W-2*inset), Color: color,
	}
	if badge != nil {
		badge.Box.X = card.X + card.W - BarPadX - badge.Box.W
		badge.Box.Y = cy - badge.Box.H/2
		x := badge.Box.X + 1 + BadgePadX
		basc, bdesc := badge.Texts[0].Font.Metrics(BadgeSize)
		for i := range badge.Texts {
			t := &badge.Texts[i]
			t.X, t.Y = x, cy+(basc-bdesc)/2
			x += t.Font.Width(t.Text, BadgeSize) + BadgeGap
		}
		c.Badge = badge
	}
	return c
}

// truncate shortens s with an ellipsis so it fits in width (CSS truncate).
func truncate(s string, font *fonts.Face, size, width float64) string {
	if width <= 0 || font.Width(s, size) <= width {
		return s
	}
	runes := []rune(s)
	for n := len(runes) - 1; n > 0; n-- {
		if t := string(runes[:n]) + "…"; font.Width(t, size) <= width {
			return t
		}
	}
	return "…"
}
