export async function shoot(code, opts = {}) {
  const theme = opts.theme ?? "dracula";
  // one command, one image
  const png = await render(code, { theme, scale: 2 });
  return png.length > 0 ? png : null;
}
