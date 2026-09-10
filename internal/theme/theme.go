// Package theme holds the twelve editor themes (in Prism's theme format)
// and the eleven backdrops, loaded from themes.json.
package theme

import (
	_ "embed"
	"encoding/json"
	"strconv"
	"strings"
)

//go:embed themes.json
var rawJSON []byte

// Default is the theme id used when none is requested.
const Default = "dracula"

// Style is a partial text style in Prism's theme shape. Empty fields
// are "unset" and inherit when styles are merged.
type Style struct {
	Color          string  // CSS color as written in the theme ("" = inherit)
	FontStyle      string  // "italic" or ""
	FontWeight     string  // "bold", "400", "600" or ""
	Opacity        float64 // valid when HasOpacity
	HasOpacity     bool
	TextDecoration string // "underline", "line-through" or ""
}

// UnmarshalJSON accepts the theme's style objects, whose fontWeight may be a
// string or a number. backgroundColor and textShadow are ignored: the card
// paints the theme's window color behind transparent text, and SVG text has
// no shadow.
func (s *Style) UnmarshalJSON(b []byte) error {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	for k, v := range m {
		switch k {
		case "color":
			s.Color, _ = v.(string)
		case "fontStyle":
			s.FontStyle, _ = v.(string)
		case "fontWeight":
			switch w := v.(type) {
			case string:
				s.FontWeight = w
			case float64:
				s.FontWeight = strconv.Itoa(int(w))
			}
		case "opacity":
			if o, ok := v.(float64); ok {
				s.Opacity, s.HasOpacity = o, true
			}
		case "textDecorationLine":
			s.TextDecoration, _ = v.(string)
		}
	}
	return nil
}

// Merge returns s overridden by every field that is set in o.
func (s Style) Merge(o Style) Style {
	if o.Color != "" {
		s.Color = o.Color
	}
	if o.FontStyle != "" {
		s.FontStyle = o.FontStyle
	}
	if o.FontWeight != "" {
		s.FontWeight = o.FontWeight
	}
	if o.HasOpacity {
		s.Opacity, s.HasOpacity = o.Opacity, true
	}
	if o.TextDecoration != "" {
		s.TextDecoration = o.TextDecoration
	}
	return s
}

// Italic reports whether the style asks for an italic face.
func (s Style) Italic() bool { return s.FontStyle == "italic" }

// Bold reports whether the style asks for weight 600 or heavier.
func (s Style) Bold() bool {
	switch s.FontWeight {
	case "bold", "bolder":
		return true
	}
	w, err := strconv.Atoi(s.FontWeight)
	return err == nil && w >= 600
}

type entry struct {
	Types     []string `json:"types"`
	Languages []string `json:"languages"`
	Style     Style    `json:"style"`
}

// Theme is one syntax theme plus the window color the card is painted with.
type Theme struct {
	ID     string  `json:"id"`
	Label  string  `json:"label"`
	Mode   string  `json:"mode"` // "dark" or "light"
	Window string  `json:"window"`
	Plain  Style   `json:"plain"`
	Styles []entry `json:"styles"`
}

// Light reports whether the theme is a light theme; the window chrome,
// gutter and dim colors follow this.
func (t *Theme) Light() bool { return t.Mode == "light" }

// Resolved is a theme flattened for one language: prism token type -> style.
type Resolved struct {
	theme *Theme
	dict  map[string]Style
}

// Resolve flattens the theme for lang (case-insensitive), applying the style
// entries in order with shallow merges, as Prism-based renderers do.
func (t *Theme) Resolve(lang string) *Resolved {
	lang = strings.ToLower(lang)
	r := &Resolved{theme: t, dict: map[string]Style{}}
	for _, e := range t.Styles {
		if len(e.Languages) > 0 && !contains(e.Languages, lang) {
			continue
		}
		for _, ty := range e.Types {
			r.dict[ty] = r.dict[ty].Merge(e.Style)
		}
	}
	return r
}

// StyleFor merges the styles of the given prism token types in order (a
// token's own type first, then its aliases); later types win.
func (r *Resolved) StyleFor(types ...string) Style {
	var s Style
	for _, t := range types {
		if d, ok := r.dict[t]; ok {
			s = s.Merge(d)
		}
	}
	return s
}

// Theme returns the theme this was resolved from.
func (r *Resolved) Theme() *Theme { return r.theme }

var (
	themes []Theme
	byID   map[string]*Theme
)

func init() {
	var f struct {
		Themes    []Theme       `json:"themes"`
		Backdrops []rawBackdrop `json:"backdrops"`
	}
	if err := json.Unmarshal(rawJSON, &f); err != nil {
		panic("theme: bad themes.json: " + err.Error())
	}
	themes = f.Themes
	byID = make(map[string]*Theme, len(themes))
	for i := range themes {
		byID[themes[i].ID] = &themes[i]
	}
	if err := initBackdrops(f.Backdrops); err != nil {
		panic("theme: bad backdrop: " + err.Error())
	}
}

// All returns every theme in display order.
func All() []Theme { return themes }

// Get returns the theme with the given id.
func Get(id string) (*Theme, bool) {
	t, ok := byID[id]
	return t, ok
}

// IDs returns the theme ids in display order.
func IDs() []string {
	ids := make([]string, len(themes))
	for i, t := range themes {
		ids[i] = t.ID
	}
	return ids
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
