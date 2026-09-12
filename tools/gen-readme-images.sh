#!/usr/bin/env bash
# Regenerates every image the README embeds, using codeshot itself.
set -euo pipefail
cd "$(dirname "$0")/.."
go build -o codeshot ./cmd/codeshot
trap 'rm -f codeshot' EXIT
S=docs/snippet.js
mkdir -p docs/themes docs/backdrops docs/presets docs/fonts

# The animated hero is built separately by tools/gen-showcase.sh (needs ffmpeg).

# No backdrop at all (transparent PNG), and rounded backdrop corners.
./codeshot --preset code --sample --transparent --title shoot.js -o docs/transparent.png >/dev/null
./codeshot --preset code --sample --radius 28 --bg dusk --title shoot.js -o docs/rounded.png >/dev/null

# One thumbnail per theme (same snippet, neutral backdrop so the theme shows).
for t in $(./codeshot --list themes | awk '{print $1}' | grep -v '^random$'); do
  ./codeshot --theme "$t" --bg slate --scale 1 --font-size 13 --padding 20 --title "$t" -o "docs/themes/$t.png" "$S" >/dev/null
done

# One thumbnail per backdrop (tiny card so the backdrop dominates).
for b in $(./codeshot --list backdrops | awk '{print $1}' | grep -v '^random$'); do
  ./codeshot --bg "$b" --scale 1 --font-size 11 --padding 36 --no-chrome --line-numbers=false -o "docs/backdrops/$b.png" "$S" >/dev/null
done

# One thumbnail per font.
for f in $(./codeshot --list fonts | awk '{print $1}'); do
  ./codeshot --font "$f" --theme nightOwl --bg darkroom --scale 1 --font-size 14 --padding 20 --title "$f" -o "docs/fonts/$f.png" "$S" >/dev/null
done

# One render per preset, with the preset's built-in sample.
for p in $(./codeshot --list presets | awk '{print $1}'); do
  ./codeshot --preset "$p" --sample --scale 1 -o "docs/presets/$p.png" >/dev/null
done
du -ch docs/*.png docs/*/*.png | tail -1
