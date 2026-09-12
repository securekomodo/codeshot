#!/usr/bin/env bash
# Builds docs/showcase.gif: one frame per preset, in order, each captioned
# with its name and given its own theme, backdrop and font. Frames are the
# same size (ten lines each; the milestone's bigger type is balanced with
# its padding and width), rendered by codeshot at 2x and downsampled by
# ffmpeg into a looping GIF with a fresh palette per frame. Needs ffmpeg.
set -euo pipefail
cd "$(dirname "$0")/.."
go build -o codeshot ./cmd/codeshot
trap 'rm -rf codeshot docs/showcase/frames' EXIT
F=docs/showcase/frames; mkdir -p "$F"
S=docs/showcase

# n | label | file | theme | backdrop | font | extra flags
while IFS='|' read -r n label file th bg font extra; do
  [ -z "$n" ] && continue
  eval "extra_args=($extra)" # keeps quoted values such as --title "git log" together
  ./codeshot --scale 2 --width 768 --label "$label" --theme "$th" --bg "$bg" --font "$font" "${extra_args[@]}" \
    -o "$F/$n.png" "$S/$file" >/dev/null
done <<'TABLE'
01|Code Snippet|01-code-snippet.js|dracula|ember|cascadia|--title app.js
02|Terminal Window|02-terminal-window.txt|shadesOfPurple|slate|ibm|--preset terminal --title deploy
03|API Response|03-api-response.json|oneLight|paper|space|--preset api --no-badge --title response.json
04|Error & Log|04-error-log.log|vsDark|ink|cascadia|--preset error-log --title server.log
05|Database Schema|05-database-schema.sql|nightOwl|tide|jetbrains|--preset db-schema --title schema.sql
06|SQL Query|06-sql-query.sql|palenight|mint|geist|--preset sql-query --title query.sql
07|Regex|07-regex.txt|okaidia|darkroom|fira|--preset regex --flags "" --title patterns.re
08|Git Commit|08-git-commit.txt|gruvboxDark|rose|source|--preset git-commit --title "git log"
09|Project Structure|09-project-structure.txt|oneDark|dusk|fira|--preset project-structure --title structure
10|Environment Variables|10-environment-variables.txt|oceanicNext|citrus|ibm|--preset env-vars --title .env
11|Terminal Session|11-terminal-session.txt|github|mint|jetbrains|--preset terminal-session --prompt ❯ --title codeshot
12|Developer Milestone|12-developer-milestone.txt|shadesOfPurple|dusk|space|--preset dev-milestone --padding 69 --width 726
13|HTTP Request|13-http-request.txt|palenight|slate|source|--preset http-request --title request.http
14|Git Diff|14-git-diff.diff|okaidia|rose|geist|--preset git-diff --title changes.diff
15|Test Results|15-test-results.txt|nightOwl|citrus|cascadia|--preset test-results --title "npm test"
16|ASCII Tree|16-ascii-tree.txt|oneLight|paper|space|--preset ascii-tree --title tree
17|Performance Metrics|17-performance-metrics.txt|dracula|darkroom|jetbrains|--preset perf-metrics --title lighthouse
TABLE

# Every frame must be the same size for a clean GIF.
sizes=$(for f in "$F"/*.png; do file "$f" | grep -oE '[0-9]+ x [0-9]+'; done | sort | uniq -c)
[ "$(echo "$sizes" | wc -l)" -eq 1 ] || { echo "frame sizes differ:"; echo "$sizes"; exit 1; }

# 2 s per frame, downsampled to 1x, a fresh palette per frame, looping.
ffmpeg -y -loglevel error -framerate 1/2 -i "$F/%02d.png" \
  -vf "scale=iw/2:-1:flags=lanczos,split[a][b];[a]palettegen=stats_mode=single[p];[b][p]paletteuse=new=1:dither=sierra2_4a" \
  -loop 0 docs/showcase.gif
ls -la docs/showcase.gif | awk '{print $5, $9}'
