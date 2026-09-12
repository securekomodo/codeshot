// Package settings is the studio's option set: defaults, per-preset
// overrides, validation, and the output-name slug.
package settings

import (
	"fmt"
	"regexp"
	"strings"

	"codeshot/internal/fonts"
	"codeshot/internal/preset"
	"codeshot/internal/theme"
)

// Settings holds every rendering option, plus the CLI-only Scale, Wrap
// and Width.
type Settings struct {
	Preset   preset.Preset
	Theme    string
	Backdrop string
	Font     string // bundled font id or a .ttf/.otf path
	Language string
	Title    string
	Padding  int // px around the card, 0..160
	FontSize int // px, 11..28

	Radius     int // corner radius of the backdrop (the image itself), 0 = square
	CardRadius int // corner radius of the window

	Watermark        string  // WatermarkNone, WatermarkSwirl, or a path to an image
	WatermarkOpacity float64 // for image watermarks

	Label     string // caption pill drawn above the window ("" = none)
	LabelSize int    // caption font size in px

	ShowBackground  bool
	ShowChrome      bool
	Chrome          string // window style: ChromeMac or ChromeKali
	Cursor          bool   // draw a block cursor after the last line
	ShowLineNumbers bool
	Shadow          bool

	Prompt    string // terminal renderers: replaces the prompt ("" keeps whatever the content has)
	Method    string // api preset badge
	Status    string // api preset badge
	ShowBadge bool
	Flags     string // regex preset badge

	Scale    int // PNG pixel ratio
	Wrap     int // soft-wrap at this many columns; 0 = off
	Width    int // fixed card width in px; 0 = fit content
	MaxWidth int // wrap so the card is at most this wide; 0 = unlimited
}

// Watermark choices besides an image path.
const (
	WatermarkNone  = "none"
	WatermarkSwirl = "swirl" // sweeping dark bands, like a wallpaper showing through
)

// Window styles.
const (
	ChromeMac  = "mac"  // macOS title bar with traffic lights
	ChromeKali = "kali" // Kali Linux terminal: title bar, menu bar, controls on the right
)

// Chromes lists the window styles.
var Chromes = []string{ChromeMac, ChromeKali}

// DefaultLabelSize is the caption's font size in px.
const DefaultLabelSize = 14

// DefaultCardRadius is the window's corner radius in px.
const DefaultCardRadius = 12

// DefaultMaxWidth is the widest a card gets unless --width, --wrap or
// --max-width says otherwise: long lines soft-wrap instead of producing a
// screenshot as wide as the longest URL in a log.
const DefaultMaxWidth = 768

// Allowed values for the badge and prompt options.
var (
	Methods  = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	Statuses = []string{"200", "201", "204", "400", "401", "403", "404", "422", "500"}
	Prompts  = []string{"$", "❯", "#", "~", ""}
)

// Defaults returns the studio's initial state for a preset.
func Defaults(p preset.Preset) Settings {
	s := Settings{
		Preset:           p,
		Theme:            theme.Default,
		Backdrop:         theme.DefaultBackdrop,
		Font:             fonts.Default,
		Language:         p.Language,
		Title:            p.Title,
		Padding:          48,
		FontSize:         15,
		ShowBackground:   true,
		ShowChrome:       p.Render != preset.Milestone,
		Chrome:           ChromeMac,
		Watermark:        WatermarkNone,
		WatermarkOpacity: 0.12,
		ShowLineNumbers:  p.SupportsLineNumbers(),
		Shadow:           true,
		Method:           "GET",
		Status:           "200",
		ShowBadge:        true,
		Scale:            2,
		MaxWidth:         DefaultMaxWidth,
		CardRadius:       DefaultCardRadius,
		LabelSize:        DefaultLabelSize,
	}
	if s.Language == "" {
		s.Language = "javascript"
	}
	if p.Render == preset.Milestone {
		s.Padding = 64
	}
	if p.Key == "regex" {
		s.Flags = "gi"
	}
	return s
}

// Validate checks every value against its range or registry.
// The font is validated when it is loaded.
func (s *Settings) Validate() error {
	if _, ok := theme.Get(s.Theme); !ok {
		return fmt.Errorf("unknown theme %q (one of %s)", s.Theme, strings.Join(theme.IDs(), ", "))
	}
	if _, ok := theme.GetBackdrop(s.Backdrop); !ok {
		return fmt.Errorf("unknown backdrop %q (one of %s)", s.Backdrop, strings.Join(theme.BackdropIDs(), ", "))
	}
	if s.Padding < 0 || s.Padding > 160 {
		return fmt.Errorf("padding %d out of range 0..160", s.Padding)
	}
	if s.FontSize < 11 || s.FontSize > 28 {
		return fmt.Errorf("font size %d out of range 11..28", s.FontSize)
	}
	if !contains(Chromes, s.Chrome) {
		return fmt.Errorf("unknown window style %q (one of %s)", s.Chrome, strings.Join(Chromes, ", "))
	}
	if s.WatermarkOpacity < 0 || s.WatermarkOpacity > 1 {
		return fmt.Errorf("watermark opacity %v out of range 0..1", s.WatermarkOpacity)
	}
	if s.LabelSize < 8 || s.LabelSize > 64 {
		return fmt.Errorf("label size %d out of range 8..64", s.LabelSize)
	}
	if s.Radius < 0 || s.Radius > 200 || s.CardRadius < 0 || s.CardRadius > 200 {
		return fmt.Errorf("radius values must be 0..200")
	}
	if s.Scale < 1 || s.Scale > 8 {
		return fmt.Errorf("scale %d out of range 1..8", s.Scale)
	}
	if s.Wrap < 0 || s.Width < 0 || s.MaxWidth < 0 {
		return fmt.Errorf("wrap, width and max-width must not be negative")
	}
	if s.Wrap > 0 && s.Width > 0 {
		return fmt.Errorf("use either --wrap or --width, not both")
	}
	if s.MaxWidth > 0 && s.MaxWidth < 200 {
		return fmt.Errorf("max-width %d is too small (200 or more, or 0 for unlimited)", s.MaxWidth)
	}
	if !contains(Methods, s.Method) {
		return fmt.Errorf("unknown method %q (one of %s)", s.Method, strings.Join(Methods, ", "))
	}
	if !contains(Statuses, s.Status) {
		return fmt.Errorf("unknown status %q (one of %s)", s.Status, strings.Join(Statuses, ", "))
	}
	return nil
}

// DisplayTitle is the title-bar text: the title, with " — zsh" appended for
// the terminal presets in the macOS style unless it already ends that way.
func (s Settings) DisplayTitle() string {
	if s.Chrome == ChromeMac && s.Preset.IsTerminal() && !strings.HasSuffix(s.Title, " — zsh") && s.Title != "zsh" {
		return s.Title + " — zsh"
	}
	return s.Title
}

// Badge kinds shown in the title bar's right slot.
const (
	BadgeNone  = ""
	BadgeAPI   = "api"   // method + status
	BadgeRegex = "regex" // "/" + flags
)

// Badge reports which badge, if any, the title bar shows.
func (s Settings) Badge() string {
	switch {
	case s.Preset.Key == "api" && s.ShowBadge:
		return BadgeAPI
	case s.Preset.Key == "regex" && s.Flags != "":
		return BadgeRegex
	}
	return BadgeNone
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Slug derives the output file stem from the title: lower-case,
// non-alphanumeric runs become "-", trimmed, at most 40
// characters, falling back to the preset key and then to "codeshot".
func Slug(title, presetKey string) string {
	s := title
	if s == "" {
		s = presetKey
	}
	s = nonAlnum.ReplaceAllString(strings.ToLower(s), "-")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = s[:40]
	}
	if s == "" {
		return "codeshot"
	}
	return s
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
