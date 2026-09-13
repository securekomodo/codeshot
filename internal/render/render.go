// Package render runs the whole pipeline: prepare the text, highlight or
// colorize it, lay out the card, and serialize it as SVG or rasterize it
// to PNG.
package render

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/securekomodo/codeshot/internal/fonts"
	"github.com/securekomodo/codeshot/internal/highlight"
	"github.com/securekomodo/codeshot/internal/layout"
	"github.com/securekomodo/codeshot/internal/preset"
	"github.com/securekomodo/codeshot/internal/raster"
	"github.com/securekomodo/codeshot/internal/settings"
	"github.com/securekomodo/codeshot/internal/svg"
	"github.com/securekomodo/codeshot/internal/theme"
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
	if bold, err := fonts.Bold(code.ID); err != nil {
		return nil, err
	} else if bold != nil {
		faces = append(faces, bold)
	}
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
		opts := highlight.Options{Theme: th, Prompt: s.Prompt}
		if s.Chrome == settings.ChromeKali {
			opts.PromptColor, opts.PromptStyle = highlight.KaliBlue, highlight.PromptStyleKali
		}
		in.Lines = highlight.Colorize(s.Preset.Render, lines, opts)
	}

	if s.Watermark != "" && s.Watermark != settings.WatermarkNone && s.Watermark != settings.WatermarkSwirl {
		img, err := loadImage(s.Watermark, s.WatermarkOpacity)
		if err != nil {
			return nil, err
		}
		in.Image = img
	}

	L, err := layout.Compute(in)
	if err != nil {
		return nil, err
	}
	return &Card{Layout: L, faces: faces}, nil
}

// loadImage reads a PNG, JPEG or SVG file for use as a watermark.
func loadImage(path string, opacity float64) (*layout.WatermarkImage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("watermark: %w", err)
	}
	img := &layout.WatermarkImage{Data: data, Opacity: opacity}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".svg":
		img.Mime = "image/svg+xml"
		img.W, img.H = svgSize(data)
	default:
		cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("watermark %s: %w (PNG, JPEG or SVG)", path, err)
		}
		img.Mime, img.W, img.H = "image/"+format, cfg.Width, cfg.Height
	}
	return img, nil
}

var svgDims = regexp.MustCompile(`(?s)<svg[^>]*?\s(width|height)="([0-9.]+)[a-z]*"[^>]*?\s(width|height)="([0-9.]+)[a-z]*"`)

// svgSize reads an SVG's width and height attributes (1:1 if absent).
func svgSize(data []byte) (int, int) {
	if m := svgDims.FindSubmatch(data); m != nil {
		a, _ := strconv.ParseFloat(string(m[2]), 64)
		b, _ := strconv.ParseFloat(string(m[4]), 64)
		if string(m[1]) == "width" {
			return int(a), int(b)
		}
		return int(b), int(a)
	}
	return 1, 1
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
