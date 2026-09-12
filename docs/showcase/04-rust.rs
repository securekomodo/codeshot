pub fn render(src: &str, opts: &Options) -> Result<Vec<u8>> {
    let lines = prepare(src);
    let card = layout(&highlight(&lines, &opts.theme))?;
    // rasterize at 2x for retina screens
    let png = raster::png(&svg::render(&card), opts.scale)?;
    match png.len() {
        0 => Err(Error::Empty),
        _ => Ok(png),
    }
}
