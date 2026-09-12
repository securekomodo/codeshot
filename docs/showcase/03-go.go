func Render(src string, opts Options) ([]byte, error) {
	lines := PrepareText(src)
	card, err := layout.Compute(highlight(lines, opts.Theme))
	if err != nil {
		return nil, fmt.Errorf("layout: %w", err)
	}
	// rasterize at 2x for retina screens
	png, err := raster.PNG(svg.Render(card), opts.Fonts, 2)
	return png, err
}
