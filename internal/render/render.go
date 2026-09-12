// Package render runs the whole pipeline: prepare the text, highlight or
// colorize it, lay out the card, and serialize it as SVG or rasterize it
// to PNG.
package render

import (
	"fmt"

	"codeshot/internal/fonts"
	"codeshot/internal/highlight"
	"codeshot/internal/layout"
	"codeshot/internal/preset"
	"codeshot/internal/raster"
	"codeshot/internal/settings"
	"codeshot/internal/svg"
	"codeshot/internal/theme"
)

// Card is a laid-out image ready to be serialized.
type Card struct {
	Layout *layout.Layout
	faces  []*fonts.Face
}

// Build validates the settings and lays out src as a card.
func Build(s settings.Settings, src string) (*Card, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	th, _ := theme.Get(s.Theme)
	bd, _ := theme.GetBackdrop(s.Backdrop)

	code, err := fonts.Load(s.Font)
	if err != nil {
		return nil, err
	}
	fallbacks, err := fonts.Fallbacks()
	if err != nil {
		return nil, err
	}
	code.SetFallbacks(fallbacks...)
	faces := append([]*fonts.Face{code}, fallbacks...)
	in := layout.Input{Settings: s, Theme: th, Backdrop: bd, Code: code}
	if s.ShowChrome || s.Label != "" {
		inter, err := fonts.Inter()
		if err != nil {
			return nil, err
		}
		in.Title = inter
		faces = append(faces, inter)
	}
	if s.ShowChrome {
		if s.Badge() != settings.BadgeNone && code.ID != "jetbrains" {
			// The badge always uses the UI mono font, JetBrains Mono.
			mono, err := fonts.Load("jetbrains")
			if err != nil {
				return nil, err
			}
			in.Badge = mono
			faces = append(faces, mono)
		}
	}

	lines := layout.PrepareText(src)
	if s.Preset.Render == preset.Prism {
		in.Lines, err = highlight.Highlight(lines, s.Language, th)
		if err != nil {
			return nil, fmt.Errorf("highlight: %w", err)
		}
	} else {
		in.Lines = highlight.Colorize(s.Preset.Render, lines, s.Prompt, th)
	}

	L, err := layout.Compute(in)
	if err != nil {
		return nil, err
	}
	return &Card{Layout: L, faces: faces}, nil
}

// SVG serializes the card at 1x. With embed set, the code and title fonts
// are inlined as @font-face data URIs (the fallback faces are not: browsers
// have their own symbol and emoji fonts, and they would add 2 MB).
func (c *Card) SVG(embed bool) []byte {
	o := svg.Options{}
	if embed {
		for _, f := range c.faces {
			if f.ID != "fallback" {
				o.EmbedFonts = append(o.EmbedFonts, f)
			}
		}
	}
	return svg.Render(c.Layout, o)
}

// PNG rasterizes the card at the given pixel ratio.
func (c *Card) PNG(scale int) ([]byte, error) {
	if scale < 1 {
		scale = 1
	}
	doc := svg.Render(c.Layout, svg.Options{Scale: float64(scale), FastShadow: true})
	data := make([][]byte, len(c.faces))
	for i, f := range c.faces {
		data[i] = f.Data
	}
	w := uint32(c.Layout.W * float64(scale))
	h := uint32(c.Layout.H * float64(scale))
	return raster.PNG(doc, data, w, h)
}

// Size returns the PNG dimensions at the given scale.
func (c *Card) Size(scale int) (w, h int) {
	return int(c.Layout.W) * scale, int(c.Layout.H) * scale
}
