#!/usr/bin/env bash
# Builds docs/showcase.gif: ten frames of the same size, each a different
# language or preset with its own theme, backdrop and font, rendered by
# codeshot at 2x and downsampled for crisp text. Needs ffmpeg.
set -euo pipefail
cd "$(dirname "$0")/.."
go build -o codeshot ./cmd/codeshot
trap 'rm -rf codeshot docs/showcase/frames' EXIT
F=docs/showcase/frames; mkdir -p "$F"
S=docs/showcase
common=(--width 768 --scale 2)

./codeshot "${common[@]}" --theme dracula        --bg ember    --font cascadia  --title app.js         -o "$F/01.png" "$S/01-javascript.js" >/dev/null
./codeshot "${common[@]}" --theme nightOwl       --bg tide     --font jetbrains --title shot.py        -o "$F/02.png" "$S/02-python.py"     >/dev/null
./codeshot "${common[@]}" --theme oneDark        --bg dusk     --font fira      --title render.go      -o "$F/03.png" "$S/03-go.go"         >/dev/null
./codeshot "${common[@]}" --theme okaidia        --bg darkroom --font geist     --title render.rs      -o "$F/04.png" "$S/04-rust.rs"       >/dev/null
./codeshot "${common[@]}" --theme shadesOfPurple --bg slate    --font ibm       --preset terminal --title deploy -o "$F/05.png" "$S/05-terminal.txt" >/dev/null
./codeshot "${common[@]}" --theme gruvboxDark    --bg rose     --font source    --preset git-diff -o "$F/06.png" "$S/06-diff.diff"    >/dev/null
./codeshot "${common[@]}" --theme oneLight       --bg paper    --font space     --preset api --no-badge --title response.json -o "$F/07.png" "$S/07-json.json" >/dev/null
./codeshot "${common[@]}" --theme palenight      --bg mint     --font jetbrains --preset sql-query -o "$F/08.png" "$S/08-sql.sql"    >/dev/null
./codeshot "${common[@]}" --theme vsDark         --bg ink      --font cascadia  --preset error-log -o "$F/09.png" "$S/09-log.log"    >/dev/null
./codeshot "${common[@]}" --theme github         --bg citrus   --font fira      --preset test-results --title "npm test" -o "$F/10.png" "$S/10-tests.txt" >/dev/null

# Every frame must be the same size for a clean GIF.
sizes=$(for f in "$F"/*.png; do file "$f" | grep -oE '[0-9]+ x [0-9]+'; done | sort -u)
[ "$(echo "$sizes" | wc -l)" -eq 1 ] || { echo "frame sizes differ:"; echo "$sizes"; exit 1; }

# 2.2 s per frame, downsampled to 1x, a fresh palette per frame, looping.
ffmpeg -y -loglevel error -framerate 1/2.2 -i "$F/%02d.png" \
  -vf "scale=iw/2:-1:flags=lanczos,split[a][b];[a]palettegen=stats_mode=single[p];[b][p]paletteuse=new=1:dither=sierra2_4a" \
  -loop 0 docs/showcase.gif
ls -la docs/showcase.gif | awk '{print $5, $9}'
