package theme

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Color is an opaque hex color plus an alpha: the form SVG attributes need,
// since resvg does not accept rgba() literals.
type Color struct {
	Hex   string  // "#rrggbb"
	Alpha float64 // 0..1
}

// RGBA returns a Color from 8-bit channels and an alpha in 0..1.
func RGBA(r, g, b uint8, a float64) Color {
	return Color{Hex: fmt.Sprintf("#%02x%02x%02x", r, g, b), Alpha: a}
}

// ParseColor parses the CSS color forms used by the themes: #rgb, #rrggbb,
// #rrggbbaa, rgb(), rgba(), hsl(), hsla() and "transparent".
func ParseColor(s string) (Color, error) {
	s = strings.TrimSpace(s)
	switch {
	case s == "transparent":
		return Color{Hex: "#000000", Alpha: 0}, nil
	case strings.HasPrefix(s, "#"):
		return parseHex(s)
	case strings.HasPrefix(s, "rgb"):
		return parseFunc(s, false)
	case strings.HasPrefix(s, "hsl"):
		return parseFunc(s, true)
	}
	return Color{}, fmt.Errorf("unsupported color %q", s)
}

func parseHex(s string) (Color, error) {
	h := s[1:]
	if len(h) == 3 || len(h) == 4 {
		var b strings.Builder
		for _, c := range h {
			b.WriteRune(c)
			b.WriteRune(c)
		}
		h = b.String()
	}
	if len(h) != 6 && len(h) != 8 {
		return Color{}, fmt.Errorf("bad hex color %q", s)
	}
	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return Color{}, fmt.Errorf("bad hex color %q", s)
	}
	if len(h) == 8 {
		return RGBA(uint8(v>>24), uint8(v>>16), uint8(v>>8), float64(v&0xff)/255), nil
	}
	return RGBA(uint8(v>>16), uint8(v>>8), uint8(v), 1), nil
}

func parseFunc(s string, hsl bool) (Color, error) {
	open, close := strings.IndexByte(s, '('), strings.LastIndexByte(s, ')')
	if open < 0 || close < open {
		return Color{}, fmt.Errorf("bad color %q", s)
	}
	parts := strings.FieldsFunc(s[open+1:close], func(r rune) bool { return r == ',' || r == ' ' || r == '/' })
	if len(parts) < 3 {
		return Color{}, fmt.Errorf("bad color %q", s)
	}
	alpha := 1.0
	if len(parts) > 3 {
		a, err := strconv.ParseFloat(strings.TrimSuffix(parts[3], "%"), 64)
		if err != nil {
			return Color{}, fmt.Errorf("bad alpha in %q", s)
		}
		if strings.HasSuffix(parts[3], "%") {
			a /= 100
		}
		alpha = a
	}
	nums := make([]float64, 3)
	for i := 0; i < 3; i++ {
		p := parts[i]
		pct := strings.HasSuffix(p, "%")
		n, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSuffix(p, "%"), "deg"), 64)
		if err != nil {
			return Color{}, fmt.Errorf("bad channel %q in %q", p, s)
		}
		if pct {
			n /= 100
			if !hsl {
				n *= 255
			}
		} else if hsl && i > 0 {
			n /= 100
		}
		nums[i] = n
	}
	if !hsl {
		return RGBA(clamp8(nums[0]), clamp8(nums[1]), clamp8(nums[2]), alpha), nil
	}
	r, g, b := hslToRGB(nums[0], nums[1], nums[2])
	return RGBA(r, g, b, alpha), nil
}

func clamp8(v float64) uint8 {
	return uint8(math.Max(0, math.Min(255, math.Round(v))))
}

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	h = math.Mod(math.Mod(h, 360)+360, 360) / 360
	if s == 0 {
		v := clamp8(l * 255)
		return v, v, v
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	hue := func(t float64) float64 {
		if t < 0 {
			t++
		}
		if t > 1 {
			t--
		}
		switch {
		case t < 1.0/6:
			return p + (q-p)*6*t
		case t < 0.5:
			return q
		case t < 2.0/3:
			return p + (q-p)*(2.0/3-t)*6
		}
		return p
	}
	return clamp8(hue(h+1.0/3) * 255), clamp8(hue(h) * 255), clamp8(hue(h-1.0/3) * 255)
}
