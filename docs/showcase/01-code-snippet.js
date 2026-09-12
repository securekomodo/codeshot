export async function shoot(code, opts = {}) {
  const theme = opts.theme ?? "dracula";
  const card = layout(highlight(code, theme), {
    font: opts.font ?? "cascadia",
    padding: 48,
  });
  // one command, one image
  const png = await rasterize(card, { scale: 2 });
  return png.length > 0 ? png : null;
}
