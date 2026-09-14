// Package layout positions every element of the card in CSS pixels at 1x:
// backdrop padding, the rounded window with its title bar, gutter and code
// rows.
package layout

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/securekomodo/codeshot/internal/fonts"
	"github.com/securekomodo/codeshot/internal/highlight"
	"github.com/securekomodo/codeshot/internal/preset"
	"github.com/securekomodo/codeshot/internal/settings"
	"github.com/securekomodo/codeshot/internal/theme"
)

// Fixed geometry of the card, in CSS px.
const (
	BarHeight       = 36   // px-4 py-3 around 12px dots
	BarWithBadge    = 50.5 // py-3 around a 26.5px badge
	DotRadius       = 6
	DotGap          = 8
	BarPadX         = 16
	TitleInset      = 80  // minimum clearance on each side of the centered title (dots on the left)
	MaxTitleRoom    = 400 // a longer title is truncated rather than widening the card
	TitleSize       = 13
	BadgeSize       = 11
	BadgePadX       = 8
	BadgeHeight     = 26.5
	BadgeRadius     = 6
	BadgeGap        = 8
	KaliTitleHeight = 33 // Kali: title bar; its contents sit at KaliTitleMid
	KaliTitleMid    = 19
	KaliMenuHeight  = 21 // Kali: "File Actions Edit View Help" bar
	KaliMenuPadX    = 19
	KaliTitleSize   = 11
	KaliMenuSize    = 13
	KaliButtonR     = 7
	KaliButtonGap   = 22 // center to center
	KaliButtonInset = 19 // close button center from the right edge
	KaliIconInset   = 21 // terminal icon center from the left edge
	KaliMenuGap     = 15

	// Windows console: icon and title on the left, three line-glyph controls
	// on the right, each in its own cell.
	WinTitleHeight = 32
	WinTitleSize   = 13
	WinIconInset   = 14
	WinIconSize    = 16
	WinTitleGap    = 10
	WinButtonW     = 46
	WinGlyph       = 10
	CodePadX       = 16
	CodePadTop     = 4
	CodePadBottom  = 20
	GutterPad      = 16
	MaxCardWidth   = 4096
	MinCardWidth   = 120
)

// DotColors are the macOS traffic lights.
var DotColors = [3]string{"#ff5f57", "#febc2e", "#28c840"}

// Kali window colors (the Kali-Dark window theme and terminal scheme).
const (
	KaliBlue       = highlight.KaliBlue // close button; user and host in the prompt
	KaliBar        = "#1e2028"          // title and menu bars, one continuous surface
	KaliButton     = "#40444f"          // minimize and maximize discs
	KaliButtonEdge = "#111318"          // thin black outline around each disc, and the x on close
	KaliText       = "#e6e6e6"

	// Windows console chrome.
	WinBar  = "#1f1f1f"
	WinText = "#ffffff"
	WinIcon = "#2f6fd0"
)

// KaliMenu is the menu bar of the Kali terminal.
var KaliMenu = []string{"File", "Actions", "Edit", "View", "Help"}

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
	Image    *WatermarkImage  // optional overlay
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

// WatermarkImage is a user-supplied image to overlay on the window.
type WatermarkImage struct {
	Data    []byte
	Mime    string // image/png, image/jpeg, image/svg+xml
	W, H    int    // pixel size, for the aspect ratio
	Opacity float64
}

// Watermark is the placed overlay: an image box inside the window.
type Watermark struct {
	Box     Rect
	Data    []byte
	Mime    string
	Opacity float64
}

// Label is the caption pill above the window.
type Label struct {
	Box  Rect
	Text Text
}

// Button is one Kali window control.
type Button struct {
	CX, CY float64
	Kind   string // "minimize", "maximize" or "close"
}

// Chrome is the window's title area: for the macOS style a single bar with
// traffic lights; for the Kali style a title bar with controls on the
// right plus a menu bar.
type Chrome struct {
	Style   string
	Bar     Rect // the whole area above the code
	Dots    [3]Dot
	Buttons []Button
	Title   Text
	Badge   *Badge
	MenuBar Rect
	Menu    []Text
	Icon    *Rect // Kali: the terminal icon at the left of the title bar
}

// Layout is the fully positioned card at 1x.
type Layout struct {
	W, H           float64
	Backdrop       theme.Backdrop
	ShowBackground bool
	Card           Rect
	Radius         float64 // backdrop corner radius
	CardRadius     float64 // window corner radius
	Window         theme.Color
	Shadow         bool
	Light          bool
	Chrome         *Chrome // nil when the window chrome is hidden
	Label          *Label  // nil when there is no caption
	Cursor         *Rect   // block cursor after the last line, when asked for
	Swirl          bool    // draw the built-in sweeping bands behind the content
	Watermark      *Watermark
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
	if s.ShowChrome && s.Chrome == settings.ChromeKali {
		// Terminal line spacing, so box-drawing characters join into
		// continuous lines (Kali's two-line prompt relies on it).
		lh = math.Round(size * 1.2)
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

	maxW, cursorW := 0.0, 0.0
	for _, r := range rows {
		w := 0.0
		for _, sp := range r.line {
			w += in.Code.Width(sp.Text, size)
		}
		maxW = math.Max(maxW, w)
		if strings.TrimSpace(r.line.Text()) != "" {
			cursorW = w + 1.5*cell // the block cursor sits in the next cell along
		}
	}
	if s.Cursor {
		maxW = math.Max(maxW, cursorW)
	}
	// The title is centered, so it needs the same clearance on both sides:
	// enough for the dots on the left and the badge on the right.
	inset := math.Max(TitleInset, BarPadX+badgeW+BadgeGap)
	kali := s.ShowChrome && s.Chrome == settings.ChromeKali
	if kali {
		inset = 3*KaliButtonGap + KaliButtonInset // the controls sit on the right, the title is centered
	}
	windows := s.ShowChrome && s.Chrome == settings.ChromeWindows
	cardW := math.Ceil(2*CodePadX + gutterW + maxW)
	if s.Width > 0 {
		cardW = float64(s.Width)
	} else {
		minW := float64(MinCardWidth)
		if s.ShowChrome && !windows {
			titleW := math.Min(titleFont.Width(s.DisplayTitle(), TitleSize), MaxTitleRoom)
			minW = math.Ceil(2*inset + titleW)
		}
		if windows {
			// icon, title and the three control cells
			titleW := math.Min(titleFont.Width(s.DisplayTitle(), WinTitleSize), MaxTitleRoom)
			minW = math.Ceil(WinIconInset + WinIconSize + WinTitleGap + titleW + WinTitleGap + 3*WinButtonW + badgeW)
		}
		if kali {
			menuW := float64(KaliMenuPadX)
			for _, item := range KaliMenu {
				menuW += titleFont.Width(item, KaliMenuSize) + KaliMenuGap
			}
			minW = math.Max(minW, math.Ceil(menuW+BarPadX+badgeW))
		}
		cardW = math.Min(math.Max(cardW, minW), capW)
	}

	barH := 0.0
	if s.ShowChrome {
		switch s.Chrome {
		case settings.ChromeWindows:
			barH = WinTitleHeight
		case settings.ChromeKali:
			barH = KaliTitleHeight + KaliMenuHeight
		default:
			barH = BarHeight
			if badge != nil {
				barH = BarWithBadge
			}
		}
	}
	// Whole pixels keep the PNG size exact at every scale (the badge bar is 50.5px).
	cardH := math.Ceil(barH + CodePadTop + float64(len(rows))*lh + CodePadBottom)
	pad := float64(s.Padding)
	// A caption takes the place of the top padding: the space above the
	// window is whatever is larger, the padding or the pill with some air.
	top := pad
	if s.Label != "" {
		top = math.Max(pad, labelBand(float64(s.LabelSize)))
	}

	L := &Layout{
		W: cardW + 2*pad, H: cardH + pad + top,
		Backdrop: in.Backdrop, ShowBackground: s.ShowBackground && !in.Backdrop.Transparent,
		Card:   Rect{pad, top, cardW, cardH},
		Radius: float64(s.Radius), CardRadius: float64(s.CardRadius),
		Window: window, Shadow: s.Shadow, Light: light,
		Font: in.Code, FontSize: size, LineHeight: lh, Plain: plain, Center: milestone,
	}

	if s.ShowChrome && s.Chrome == settings.ChromeWindows {
		L.Chrome = buildWindowsChrome(L.Card, s.DisplayTitle(), titleFont, badge)
	} else if s.ShowChrome && s.Chrome == settings.ChromeKali {
		L.Chrome = buildKaliChrome(L.Card, s.DisplayTitle(), titleFont, badge)
	} else if s.ShowChrome {
		L.Chrome = buildChrome(L.Card, barH, s.DisplayTitle(), titleFont, badge, inset, light)
	}
	if s.Label != "" {
		L.Label = buildLabel(L.Card, s.Label, titleFont, float64(s.LabelSize), top)
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
	L.Swirl = s.Watermark == settings.WatermarkSwirl
	if in.Image != nil && in.Image.W > 0 && in.Image.H > 0 {
		// Bottom-right of the window body, about two thirds of its height,
		// bleeding a little off the right edge the way a wallpaper peeks in.
		bodyY := L.Card.Y + barH
		h := math.Min((L.Card.H-barH)*0.7, L.Card.W*0.7)
		w := h * float64(in.Image.W) / float64(in.Image.H)
		L.Watermark = &Watermark{
			Box:  Rect{L.Card.X + L.Card.W - w*0.85, bodyY + (L.Card.H - barH) - h*0.95, w, h},
			Data: in.Image.Data, Mime: in.Image.Mime, Opacity: in.Image.Opacity,
		}
	}
	if s.Cursor && len(L.Rows) > 0 {
		// The cursor sits at the last row that has text, as in a terminal.
		last := L.Rows[len(L.Rows)-1]
		for i := len(L.Rows) - 1; i >= 0; i-- {
			if strings.TrimSpace(L.Rows[i].Spans.Text()) != "" {
				last = L.Rows[i]
				break
			}
		}
		x := last.X
		for _, sp := range last.Spans {
			x += in.Code.Width(sp.Text, size)
		}
		if milestone {
			x = last.X + (x-last.X)/2
		}
		// Sit in the next cell with some air, as a terminal's cursor does.
		L.Cursor = &Rect{x + cell/2, last.Y - asc, cell, asc + desc}
	}
	return L, nil
}

// buildWindowsChrome lays out a Windows console title bar: the icon and
// title run along the left, the three controls sit in equal cells on the right.
func buildWindowsChrome(card Rect, title string, font *fonts.Face, badge *Badge) *Chrome {
	c := &Chrome{Style: settings.ChromeWindows, Bar: Rect{card.X, card.Y, card.W, WinTitleHeight}}
	cy := card.Y + WinTitleHeight/2
	c.Icon = &Rect{card.X + WinIconInset, cy - WinIconSize/2, WinIconSize, WinIconSize}
	for i, kind := range []string{"minimize", "maximize", "close"} {
		c.Buttons = append(c.Buttons, Button{
			CX:   card.X + card.W - float64(3-i)*WinButtonW + WinButtonW/2,
			CY:   cy,
			Kind: kind,
		})
	}
	asc, desc := font.Metrics(WinTitleSize)
	x := card.X + WinIconInset + WinIconSize + WinTitleGap
	c.Title = Text{
		X: x, Y: cy + (asc-desc)/2, Anchor: "start", Size: WinTitleSize, Font: font,
		Text:  truncate(title, font, WinTitleSize, card.X+card.W-3*WinButtonW-x-WinTitleGap),
		Color: theme.Color{Hex: WinText, Alpha: 1},
	}
	if badge != nil {
		badge.Box.X = card.X + card.W - 3*WinButtonW - WinTitleGap - badge.Box.W
		badge.Box.Y = cy - badge.Box.H/2
		bx := badge.Box.X + 1 + BadgePadX
		basc, bdesc := badge.Texts[0].Font.Metrics(BadgeSize)
		for i := range badge.Texts {
			t := &badge.Texts[i]
			t.X, t.Y = bx, cy+(basc-bdesc)/2
			bx += t.Font.Width(t.Text, BadgeSize) + BadgeGap
		}
		c.Badge = badge
	}
	return c
}

// buildKaliChrome lays out the Kali terminal's title bar (controls on the
// right, title centered) and menu bar.
func buildKaliChrome(card Rect, title string, font *fonts.Face, badge *Badge) *Chrome {
	c := &Chrome{Style: settings.ChromeKali, Bar: Rect{card.X, card.Y, card.W, KaliTitleHeight + KaliMenuHeight}}
	c.MenuBar = Rect{card.X, card.Y + KaliTitleHeight, card.W, KaliMenuHeight}
	cy := card.Y + KaliTitleMid
	c.Icon = &Rect{card.X + KaliIconInset - 7, cy - 6, 14, 12}
	for i, kind := range []string{"minimize", "maximize", "close"} {
		c.Buttons = append(c.Buttons, Button{CX: card.X + card.W - KaliButtonInset - float64(2-i)*KaliButtonGap, CY: cy, Kind: kind})
	}
	asc, desc := font.Metrics(KaliTitleSize)
	inset := 3*KaliButtonGap + KaliButtonInset
	c.Title = Text{
		X: card.X + card.W/2, Y: cy + (asc-desc)/2, Anchor: "middle",
		Size: KaliTitleSize, Font: font, Weight: 500,
		Text: truncate(title, font, KaliTitleSize, card.W-2*float64(inset)), Color: theme.Color{Hex: KaliText, Alpha: 0.95},
	}
	masc, mdesc := font.Metrics(KaliMenuSize)
	my := c.MenuBar.Y + KaliMenuHeight/2 + (masc-mdesc)/2
	x := card.X + KaliMenuPadX
	for _, item := range KaliMenu {
		c.Menu = append(c.Menu, Text{X: x, Y: my, Anchor: "start", Size: KaliMenuSize, Font: font, Text: item,
			Color: theme.Color{Hex: KaliText, Alpha: 0.95}})
		x += font.Width(item, KaliMenuSize) + KaliMenuGap
	}
	if badge != nil {
		badge.Box.X = card.X + card.W - BarPadX - badge.Box.W
		badge.Box.Y = c.MenuBar.Y + (KaliMenuHeight-badge.Box.H)/2
		bx := badge.Box.X + 1 + BadgePadX
		basc, bdesc := badge.Texts[0].Font.Metrics(BadgeSize)
		for i := range badge.Texts {
			t := &badge.Texts[i]
			t.X, t.Y = bx, c.MenuBar.Y+KaliMenuHeight/2+(basc-bdesc)/2
			bx += t.Font.Width(t.Text, BadgeSize) + BadgeGap
		}
		c.Badge = badge
	}
	return c
}

// labelBand is the space a caption of the given font size needs above the
// window: the pill plus air above and below it.
func labelBand(size float64) float64 { return math.Round(size*2) + 2*math.Round(size*0.75) }

// buildLabel centers a caption pill in the space above the window (top px
// high). The pill and its padding scale with the font size.
func buildLabel(card Rect, text string, font *fonts.Face, size, top float64) *Label {
	padX := math.Round(size * 1.1)
	height := math.Round(size * 2)
	text = truncate(text, font, size, card.W-2*padX)
	w := math.Ceil(font.Width(text, size) + 2*padX)
	cx, cy := card.X+card.W/2, card.Y-top/2
	asc, desc := font.Metrics(size)
	return &Label{
		Box: Rect{cx - w/2, cy - height/2, w, height},
		Text: Text{X: cx, Y: cy + (asc-desc)/2, Anchor: "middle", Size: size, Font: font, Weight: 500,
			Text: text, Color: theme.Color{Hex: "#ffffff", Alpha: 0.94}},
	}
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
	c := &Chrome{Style: settings.ChromeMac, Bar: Rect{card.X, card.Y, card.W, barH}}
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
