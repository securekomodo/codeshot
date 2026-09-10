package theme

import (
	"fmt"
	"strconv"
	"strings"
)

// DefaultBackdrop is the backdrop id used when none is requested.
const DefaultBackdrop = "ember"

type rawBackdrop struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	CSS         string `json:"css"`
	Transparent bool   `json:"transparent"`
}

// Stop is one color stop of a gradient backdrop.
type Stop struct {
	Color  Color
	Offset float64 // 0..1
}

// Backdrop is the fill painted behind the card: a CSS linear gradient, a
// solid color, or nothing.
type Backdrop struct {
	ID          string
	Label       string
	Transparent bool
	Solid       Color   // used when there are no Stops and not Transparent
	Angle       float64 // CSS gradient angle in degrees (0 = to top, clockwise)
	Stops       []Stop
}

// Gradient reports whether the backdrop is a linear gradient.
func (b Backdrop) Gradient() bool { return len(b.Stops) > 0 }

var backdrops []Backdrop

func initBackdrops(raws []rawBackdrop) error {
	backdrops = backdrops[:0]
	for _, r := range raws {
		b := Backdrop{ID: r.ID, Label: r.Label, Transparent: r.Transparent}
		switch {
		case r.Transparent || r.CSS == "":
			b.Transparent = true
		case strings.HasPrefix(r.CSS, "linear-gradient("):
			if err := parseGradient(&b, r.CSS); err != nil {
				return fmt.Errorf("%s: %w", r.ID, err)
			}
		default:
			c, err := ParseColor(r.CSS)
			if err != nil {
				return fmt.Errorf("%s: %w", r.ID, err)
			}
			b.Solid = c
		}
		backdrops = append(backdrops, b)
	}
	return nil
}

// parseGradient reads "linear-gradient(135deg, #f4a259 0%, #c75c3c 50%, ...)".
func parseGradient(b *Backdrop, css string) error {
	inner := strings.TrimSuffix(strings.TrimPrefix(css, "linear-gradient("), ")")
	parts := strings.Split(inner, ",")
	if len(parts) < 3 {
		return fmt.Errorf("gradient needs an angle and two stops: %q", css)
	}
	angle, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(parts[0]), "deg"), 64)
	if err != nil {
		return fmt.Errorf("bad gradient angle in %q", css)
	}
	b.Angle = angle
	for i, p := range parts[1:] {
		fields := strings.Fields(p)
		if len(fields) == 0 {
			return fmt.Errorf("empty stop in %q", css)
		}
		c, err := ParseColor(fields[0])
		if err != nil {
			return err
		}
		off := float64(i) / float64(len(parts)-2)
		if len(fields) > 1 {
			pct, err := strconv.ParseFloat(strings.TrimSuffix(fields[1], "%"), 64)
			if err != nil {
				return fmt.Errorf("bad stop offset %q", fields[1])
			}
			off = pct / 100
		}
		b.Stops = append(b.Stops, Stop{Color: c, Offset: off})
	}
	return nil
}

// Backdrops returns every backdrop in display order.
func Backdrops() []Backdrop { return backdrops }

// GetBackdrop returns the backdrop with the given id.
func GetBackdrop(id string) (Backdrop, bool) {
	for _, b := range backdrops {
		if b.ID == id {
			return b, true
		}
	}
	return Backdrop{}, false
}

// BackdropIDs returns the backdrop ids in display order.
func BackdropIDs() []string {
	ids := make([]string, len(backdrops))
	for i, b := range backdrops {
		ids[i] = b.ID
	}
	return ids
}
