#!/usr/bin/env bash
# Builds docs/showcase.gif: one frame per preset, in order, rendering each
# preset's built-in sample captioned with the preset's name in its own
# theme, backdrop and font. Frames are the same size, rendered by codeshot
# at 2x and downsampled by ffmpeg into a looping GIF with a fresh palette
# per frame. Needs ffmpeg.
set -euo pipefail
cd "$(dirname "$0")/.."
go build -o codeshot ./cmd/codeshot
trap 'rm -rf codeshot docs/showcase-frames' EXIT
F=docs/showcase-frames; mkdir -p "$F"

# n | label | preset | theme | backdrop | font | extra flags
# Every sample is ten lines (the milestone six, balanced by padding and width).
# Frames in the Kali window style have a taller title area, balanced the same way.
while IFS='|' read -r n label preset th bg font extra; do
  [ -z "$n" ] && continue
  eval "extra_args=($extra)" # keeps quoted values such as --title "git log" together
  ./codeshot --sample --preset "$preset" --scale 2 --width 768 --label "$label" --label-size 24 \
    --theme "$th" --bg "$bg" --font "$font" "${extra_args[@]}" -o "$F/$n.png" >/dev/null
done <<'TABLE'
01|Code Snippet|code|dracula|ember|cascadia|--title alerts.js
02|Terminal Window|terminal|oneDark|slate|fira|--chrome kali --padding 37 --width 790 --title "kali@kali: ~" --cursor --prompt "kali@kali:~$"
03|API Response|api|oneLight|paper|space|--no-badge --title reading.json
04|Error & Log|error-log|vsDark|ink|cascadia|--chrome kali --padding 37 --width 790 --title "kali@kali: ~"
05|Database Schema|db-schema|nightOwl|tide|jetbrains|--title schema.sql
06|SQL Query|sql-query|palenight|mint|geist|--title warmest.sql
07|Regex|regex|okaidia|darkroom|fira|--flags "" --title patterns.re
08|Git Commit|git-commit|gruvboxDark|rose|source|--title "git log"
09|Project Structure|project-structure|oneDark|dusk|fira|--title nimbus
10|Environment Variables|env-vars|oceanicNext|citrus|ibm|--title .env
11|Terminal Session|terminal-session|nightOwl|darkroom|jetbrains|--chrome kali --padding 37 --width 790 --title "kali@kali: ~" --cursor --prompt "kali@kali:~$"
12|Developer Milestone|dev-milestone|shadesOfPurple|dusk|space|--padding 69 --width 726
13|HTTP Request|http-request|palenight|slate|source|--title alert.http
14|Git Diff|git-diff|okaidia|rose|geist|--title notify.diff
15|Test Results|test-results|gruvboxDark|citrus|ibm|--chrome kali --padding 37 --width 790 --title "kali@kali: ~"
16|ASCII Tree|ascii-tree|oneLight|paper|space|--title pipeline
17|Performance Metrics|perf-metrics|dracula|darkroom|jetbrains|--title lighthouse
TABLE

# Every frame must be the same size for a clean GIF.
sizes=$(for f in "$F"/*.png; do file "$f" | grep -oE '[0-9]+ x [0-9]+'; done | sort | uniq -c)
[ "$(echo "$sizes" | wc -l)" -eq 1 ] || { echo "frame sizes differ:"; echo "$sizes"; exit 1; }

# 2 s per frame, downsampled to 1x, a fresh palette per frame, looping.
ffmpeg -y -loglevel error -framerate 1/2 -i "$F/%02d.png" \
  -vf "scale=iw/2:-1:flags=lanczos,split[a][b];[a]palettegen=stats_mode=single[p];[b][p]paletteuse=new=1:dither=sierra2_4a" \
  -loop 0 docs/showcase.gif
ls -la docs/showcase.gif | awk '{print $5, $9}'
