// Package raster turns the SVG card into a PNG with resvg, compiled to
// WebAssembly and run in-process through wazero, so the binary needs no
// browser, CGO, or system libraries.
package raster

import (
	"context"
	"fmt"

	resvg "github.com/kanrichan/resvg-go"
)

// PNG rasterizes an SVG document to a width×height PNG. Only the given
// TTF/OTF font files are available to the renderer; system fonts are not
// loaded, so output is the same on every machine.
func PNG(svg []byte, fontData [][]byte, width, height uint32) ([]byte, error) {
	ctx, err := resvg.NewContext(context.Background())
	if err != nil {
		return nil, fmt.Errorf("resvg: %w", err)
	}
	defer ctx.Close()
	r, err := ctx.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("resvg: %w", err)
	}
	defer r.Close()
	if err := r.SetDpi(96); err != nil {
		return nil, fmt.Errorf("resvg: %w", err)
	}
	for _, d := range fontData {
		if err := r.LoadFontData(d); err != nil {
			return nil, fmt.Errorf("resvg: load font: %w", err)
		}
	}
	out, err := r.RenderWithSize(svg, width, height)
	if err != nil {
		return nil, fmt.Errorf("resvg: render: %w", err)
	}
	return out, nil
}
