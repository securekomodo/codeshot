<p align="center">
  <img src="docs/banner.png" width="960" alt="codeshot: turn code into a beautiful image in seconds">
</p>

<p align="center">
  <b>Point it at code. Get a picture.</b> One binary, no browser, no server, nothing leaves your machine.
</p>

<p align="center">
  <img alt="Go 1.25" src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white">
  <img alt="single binary" src="https://img.shields.io/badge/deps-single%20binary-f4a259">
  <img alt="output" src="https://img.shields.io/badge/output-PNG%20%C2%B7%20SVG%20%C2%B7%20clipboard-5fb3a1">
  <img alt="offline" src="https://img.shields.io/badge/network-none-0b0d10">
</p>

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/hero-dark.png">
    <img src="docs/hero-light.png" width="720" alt="A code snippet rendered as a framed image. This picture follows your GitHub theme: dark shows Dracula on the Ember backdrop, light shows GitHub Light on Paper.">
  </picture>
  <br>
  <sub>This hero follows your GitHub theme. Both versions were rendered by <code>codeshot</code> itself.</sub>
</p>

<table align="center">
  <tr>
    <td align="center"><sub>INPUT</sub><br><b>source code</b></td>
    <td align="center"><sub>OUTPUT</sub><br><b>PNG · SVG · clipboard</b></td>
    <td align="center"><sub>COST</sub><br><b>free</b></td>
    <td align="center"><sub>ACCOUNT</sub><br><b>not required</b></td>
    <td align="center"><sub>PROCESSING</sub><br><b>on your machine</b></td>
  </tr>
</table>

## The thirty-second tour

```sh
go build -o codeshot ./cmd/codeshot           # that's the whole install

codeshot main.go                              # → main-go.png, language guessed from the name
git diff | codeshot                           # piped input picks its own preset: a diff here
cat query.sql | codeshot --lang sql           # or say what it is
codeshot --theme nightOwl --bg tide app.tsx   # pick a theme and a backdrop
codeshot --copy notes.md                      # straight to the clipboard, ready to paste
codeshot --preset git-diff --sample -o d.svg  # every preset ships a sample; .svg gets you vectors
codeshot --list themes                        # themes, backdrops, fonts, languages, presets
```

## Why you'll like it

<table>
  <tr>
    <td valign="top" width="33%"><b>🎨 Real editor themes</b><br>Dracula, Night Owl, One Dark, GitHub and eight more, with the exact token colors you see in your editor, not a lookalike palette.</td>
    <td valign="top" width="33%"><b>🔤 Ligature-ready fonts</b><br>Cascadia Code, JetBrains Mono, Fira Code and four more are built into the binary. Arrows and <code>!==</code> render the way they do in your terminal.</td>
    <td valign="top" width="33%"><b>🖼️ A frame that looks finished</b><br>macOS-style window, soft shadow, gradient backdrops, line numbers, a title bar with badges. Tune every knob or keep the defaults.</td>
  </tr>
  <tr>
    <td valign="top"><b>🧰 Seventeen presets, picked for you</b><br>Terminal sessions, git logs and diffs, server logs, HTTP requests, test output, .env files, ASCII trees, Lighthouse-style metrics. Pipe something in and the right one is chosen from the content. Terminal sessions get real shell highlighting: commands, flags, strings, pipes, and status words, numbers, hashes and URLs in the output.</td>
    <td valign="top"><b>📐 Crisp at any size</b><br>Retina PNG at 2x by default (or 1x to 8x), or a standalone SVG with the fonts embedded, so it scales forever.</td>
    <td valign="top"><b>🔒 Nothing leaves your machine</b><br>No browser, no server, no network code at all. Paste production logs with a clear conscience.</td>
  </tr>
</table>

## Pick a theme

<table align="center"><tr>
<td align="center" valign="top"><a href="docs/themes/dracula.png"><img src="docs/themes/dracula.png" width="220" alt="dracula"></a><br><sub><code>--theme dracula</code></sub></td>
<td align="center" valign="top"><a href="docs/themes/nightOwl.png"><img src="docs/themes/nightOwl.png" width="220" alt="nightOwl"></a><br><sub><code>--theme nightOwl</code></sub></td>
<td align="center" valign="top"><a href="docs/themes/oneDark.png"><img src="docs/themes/oneDark.png" width="220" alt="oneDark"></a><br><sub><code>--theme oneDark</code></sub></td>
<td align="center" valign="top"><a href="docs/themes/palenight.png"><img src="docs/themes/palenight.png" width="220" alt="palenight"></a><br><sub><code>--theme palenight</code></sub></td>
</tr><tr>
<td align="center" valign="top"><a href="docs/themes/oceanicNext.png"><img src="docs/themes/oceanicNext.png" width="220" alt="oceanicNext"></a><br><sub><code>--theme oceanicNext</code></sub></td>
<td align="center" valign="top"><a href="docs/themes/shadesOfPurple.png"><img src="docs/themes/shadesOfPurple.png" width="220" alt="shadesOfPurple"></a><br><sub><code>--theme shadesOfPurple</code></sub></td>
<td align="center" valign="top"><a href="docs/themes/vsDark.png"><img src="docs/themes/vsDark.png" width="220" alt="vsDark"></a><br><sub><code>--theme vsDark</code></sub></td>
<td align="center" valign="top"><a href="docs/themes/okaidia.png"><img src="docs/themes/okaidia.png" width="220" alt="okaidia"></a><br><sub><code>--theme okaidia</code></sub></td>
</tr><tr>
<td align="center" valign="top"><a href="docs/themes/gruvboxDark.png"><img src="docs/themes/gruvboxDark.png" width="220" alt="gruvboxDark"></a><br><sub><code>--theme gruvboxDark</code></sub></td>
<td align="center" valign="top"><a href="docs/themes/github.png"><img src="docs/themes/github.png" width="220" alt="github"></a><br><sub><code>--theme github</code></sub></td>
<td align="center" valign="top"><a href="docs/themes/oneLight.png"><img src="docs/themes/oneLight.png" width="220" alt="oneLight"></a><br><sub><code>--theme oneLight</code></sub></td>
<td align="center" valign="top"><a href="docs/themes/nightOwlLight.png"><img src="docs/themes/nightOwlLight.png" width="220" alt="nightOwlLight"></a><br><sub><code>--theme nightOwlLight</code></sub></td>
</tr></table>

## Pick a backdrop

<table align="center"><tr>
<td align="center" valign="top"><a href="docs/backdrops/ember.png"><img src="docs/backdrops/ember.png" width="220" alt="ember"></a><br><sub><code>--bg ember</code></sub></td>
<td align="center" valign="top"><a href="docs/backdrops/darkroom.png"><img src="docs/backdrops/darkroom.png" width="220" alt="darkroom"></a><br><sub><code>--bg darkroom</code></sub></td>
<td align="center" valign="top"><a href="docs/backdrops/tide.png"><img src="docs/backdrops/tide.png" width="220" alt="tide"></a><br><sub><code>--bg tide</code></sub></td>
<td align="center" valign="top"><a href="docs/backdrops/dusk.png"><img src="docs/backdrops/dusk.png" width="220" alt="dusk"></a><br><sub><code>--bg dusk</code></sub></td>
</tr><tr>
<td align="center" valign="top"><a href="docs/backdrops/citrus.png"><img src="docs/backdrops/citrus.png" width="220" alt="citrus"></a><br><sub><code>--bg citrus</code></sub></td>
<td align="center" valign="top"><a href="docs/backdrops/slate.png"><img src="docs/backdrops/slate.png" width="220" alt="slate"></a><br><sub><code>--bg slate</code></sub></td>
<td align="center" valign="top"><a href="docs/backdrops/mint.png"><img src="docs/backdrops/mint.png" width="220" alt="mint"></a><br><sub><code>--bg mint</code></sub></td>
<td align="center" valign="top"><a href="docs/backdrops/rose.png"><img src="docs/backdrops/rose.png" width="220" alt="rose"></a><br><sub><code>--bg rose</code></sub></td>
</tr><tr>
<td align="center" valign="top"><a href="docs/backdrops/ink.png"><img src="docs/backdrops/ink.png" width="220" alt="ink"></a><br><sub><code>--bg ink</code></sub></td>
<td align="center" valign="top"><a href="docs/backdrops/paper.png"><img src="docs/backdrops/paper.png" width="220" alt="paper"></a><br><sub><code>--bg paper</code></sub></td>
<td align="center" valign="top"><a href="docs/backdrops/none.png"><img src="docs/backdrops/none.png" width="220" alt="none"></a><br><sub><code>--bg none</code></sub></td>
</tr></table>

## Pick a font

Every face is bundled. Point <code>--font</code> at any <code>.ttf</code> or <code>.otf</code> on disk to use your own.

<table align="center"><tr>
<td align="center" valign="top"><a href="docs/fonts/cascadia.png"><img src="docs/fonts/cascadia.png" width="220" alt="Cascadia Code"></a><br><sub>Cascadia Code · <code>--font cascadia</code></sub></td>
<td align="center" valign="top"><a href="docs/fonts/jetbrains.png"><img src="docs/fonts/jetbrains.png" width="220" alt="JetBrains Mono"></a><br><sub>JetBrains Mono · <code>--font jetbrains</code></sub></td>
<td align="center" valign="top"><a href="docs/fonts/fira.png"><img src="docs/fonts/fira.png" width="220" alt="Fira Code"></a><br><sub>Fira Code · <code>--font fira</code></sub></td>
<td align="center" valign="top"><a href="docs/fonts/geist.png"><img src="docs/fonts/geist.png" width="220" alt="Geist Mono"></a><br><sub>Geist Mono · <code>--font geist</code></sub></td>
</tr><tr>
<td align="center" valign="top"><a href="docs/fonts/ibm.png"><img src="docs/fonts/ibm.png" width="220" alt="IBM Plex Mono"></a><br><sub>IBM Plex Mono · <code>--font ibm</code></sub></td>
<td align="center" valign="top"><a href="docs/fonts/source.png"><img src="docs/fonts/source.png" width="220" alt="Source Code Pro"></a><br><sub>Source Code Pro · <code>--font source</code></sub></td>
<td align="center" valign="top"><a href="docs/fonts/space.png"><img src="docs/fonts/space.png" width="220" alt="Space Mono"></a><br><sub>Space Mono · <code>--font space</code></sub></td>
</tr></table>

## Presets for the things developers actually screenshot

Each preset sets the render mode, title and language. Add <code>--sample</code> to see it with the built-in snippet, or feed it your own.

<table align="center"><tr>
<td align="center" valign="top"><a href="docs/presets/code.png"><img src="docs/presets/code.png" width="290" alt="Code"></a><br><b>Code</b><br><sub><code>--preset code</code></sub></td>
<td align="center" valign="top"><a href="docs/presets/terminal.png"><img src="docs/presets/terminal.png" width="290" alt="Terminal"></a><br><b>Terminal</b><br><sub><code>--preset terminal</code></sub></td>
<td align="center" valign="top"><a href="docs/presets/api.png"><img src="docs/presets/api.png" width="290" alt="API response"></a><br><b>API response</b><br><sub><code>--preset api</code></sub></td>
</tr><tr>
<td align="center" valign="top"><a href="docs/presets/error-log.png"><img src="docs/presets/error-log.png" width="290" alt="Logs"></a><br><b>Logs</b><br><sub><code>--preset error-log</code></sub></td>
<td align="center" valign="top"><a href="docs/presets/db-schema.png"><img src="docs/presets/db-schema.png" width="290" alt="Schema"></a><br><b>Schema</b><br><sub><code>--preset db-schema</code></sub></td>
<td align="center" valign="top"><a href="docs/presets/sql-query.png"><img src="docs/presets/sql-query.png" width="290" alt="SQL"></a><br><b>SQL</b><br><sub><code>--preset sql-query</code></sub></td>
</tr><tr>
<td align="center" valign="top"><a href="docs/presets/regex.png"><img src="docs/presets/regex.png" width="290" alt="Regex"></a><br><b>Regex</b><br><sub><code>--preset regex</code></sub></td>
<td align="center" valign="top"><a href="docs/presets/git-commit.png"><img src="docs/presets/git-commit.png" width="290" alt="Git log"></a><br><b>Git log</b><br><sub><code>--preset git-commit</code></sub></td>
<td align="center" valign="top"><a href="docs/presets/project-structure.png"><img src="docs/presets/project-structure.png" width="290" alt="Project tree"></a><br><b>Project tree</b><br><sub><code>--preset project-structure</code></sub></td>
</tr><tr>
<td align="center" valign="top"><a href="docs/presets/env-vars.png"><img src="docs/presets/env-vars.png" width="290" alt="Env file"></a><br><b>Env file</b><br><sub><code>--preset env-vars</code></sub></td>
<td align="center" valign="top"><a href="docs/presets/terminal-session.png"><img src="docs/presets/terminal-session.png" width="290" alt="Session"></a><br><b>Session</b><br><sub><code>--preset terminal-session</code></sub></td>
<td align="center" valign="top"><a href="docs/presets/dev-milestone.png"><img src="docs/presets/dev-milestone.png" width="290" alt="Milestone"></a><br><b>Milestone</b><br><sub><code>--preset dev-milestone</code></sub></td>
</tr><tr>
<td align="center" valign="top"><a href="docs/presets/http-request.png"><img src="docs/presets/http-request.png" width="290" alt="HTTP"></a><br><b>HTTP</b><br><sub><code>--preset http-request</code></sub></td>
<td align="center" valign="top"><a href="docs/presets/git-diff.png"><img src="docs/presets/git-diff.png" width="290" alt="Diff"></a><br><b>Diff</b><br><sub><code>--preset git-diff</code></sub></td>
<td align="center" valign="top"><a href="docs/presets/test-results.png"><img src="docs/presets/test-results.png" width="290" alt="Tests"></a><br><b>Tests</b><br><sub><code>--preset test-results</code></sub></td>
</tr><tr>
<td align="center" valign="top"><a href="docs/presets/ascii-tree.png"><img src="docs/presets/ascii-tree.png" width="290" alt="ASCII tree"></a><br><b>ASCII tree</b><br><sub><code>--preset ascii-tree</code></sub></td>
<td align="center" valign="top"><a href="docs/presets/perf-metrics.png"><img src="docs/presets/perf-metrics.png" width="290" alt="Metrics"></a><br><b>Metrics</b><br><sub><code>--preset perf-metrics</code></sub></td>
</tr></table>

## Recipes

```sh
# Share what you just changed (the diff preset is picked from the content)
git diff | codeshot --title "fix: retry on 429" --copy

# A whole shell session, with commands and output highlighted
script -q /dev/null | tee session.txt; codeshot --preset terminal session.txt

# The last ten commits, styled
git log -10 | codeshot --title "git log" -o log.png

# Test output from CI, minus the noise
npm test 2>&1 | tail -20 | codeshot --preset test-results --title "npm test"

# A JSON response, pretty and framed
curl -s https://api.example.com/v1/thing | jq . | codeshot --preset api --status 200 --title thing.json

# Your project layout, for the README
tree -L 2 --noreport | codeshot --preset project-structure --bg ink

# Big, centered, no window: an announcement card
echo "🎉 v2.0 is out" | codeshot --preset dev-milestone --bg dusk

# Long lines wrap so the card stays at most 768px wide. Change the cap, fix the width, or wrap by column
codeshot --max-width 1000 server.go
codeshot --width 900 server.go
codeshot --wrap 80 server.go
codeshot --max-width 0 server.go        # no cap: the card fits the longest line
```

## How it works

```mermaid
flowchart LR
    A[code file<br>or stdin] --> B[tokenize<br><i>chroma</i>]
    B --> C[apply theme<br><i>Prism token colors</i>]
    A -. presets .-> D[line colorizers<br><i>logs, diffs, prompts…</i>]
    C --> E[lay out the card<br><i>metrics from the real font</i>]
    D --> E
    E --> F[SVG]
    F --> G[PNG<br><i>resvg via wazero</i>]
    F --> H[.svg with<br>embedded fonts]
    G --> I[clipboard]
```

1. **Point it at code.** A file, stdin, or a preset's sample. The language comes from the file name, the preset, or <code>--lang</code>; piped content that looks like a diff, a git log, a shell session, JSON, a tree, an .env file, an HTTP request, test output or a log gets that preset automatically.
2. **Frame the shot.** Theme, backdrop, font, size, padding, window, shadow, line numbers, badges.
3. **Export and share.** A retina PNG, a vector SVG, or the clipboard, ready to paste into a chat, a doc, or a slide.

The card is laid out in Go with the font's own metrics, written as SVG, and rasterized in-process by resvg compiled to WebAssembly. No CGO, no system libraries, identical output on every machine.

<details>
<summary><b>Every flag</b></summary>

```
codeshot [flags] [FILE]        FILE omitted or "-" reads stdin; flags may come before or after FILE

-o, --output PATH     output file; a .svg extension writes SVG, anything else PNG
                      (default <title-slug>.png; a temp file with --copy)
--preset KEY          code (default), terminal, api, error-log, db-schema, sql-query, regex,
                      git-commit, project-structure, env-vars, terminal-session, dev-milestone,
                      http-request, git-diff, test-results, ascii-tree, perf-metrics
--sample              render the preset's built-in sample instead of reading input
--lang ID             language for the code presets (default: the preset's, or guessed
                      from the file name; --list languages)
--theme ID            dracula (default), nightOwl, oneDark, palenight, oceanicNext,
                      shadesOfPurple, vsDark, okaidia, gruvboxDark, github, oneLight, nightOwlLight
--bg ID               ember (default), darkroom, tide, dusk, citrus, slate, mint, rose, ink, paper, none
--font ID|PATH        cascadia (default), jetbrains, fira, geist, ibm, source, space, or a .ttf/.otf file
--title TEXT          window title (default: the preset's)
--padding N           space around the card, 0..160 (default 48; 64 for dev-milestone)
--font-size N         11..28 (default 15)
--no-bg               transparent backdrop
--no-chrome           hide the title bar
--no-shadow           no drop shadow
--line-numbers BOOL   default on for the code and log presets
--prompt STR          terminal presets: prompt character ($ ❯ # ~; empty keeps the original)
--method M            api preset badge: GET POST PUT PATCH DELETE
--status N            api preset badge: 200 201 204 400 401 403 404 422 500
--no-badge            api preset: hide the badge
--flags STR           regex preset: badge flags (default gi)
--scale N             PNG pixel ratio, 1..8 (default 2)
--max-width PX        wrap long lines so the card is at most this wide (default 768; 0 = unlimited)
--width PX            fixed card width, wrapping to fit
--wrap COLS           soft-wrap at this many columns
--embed-fonts BOOL    SVG: inline the fonts as data URIs (default true)
--copy                copy the PNG to the clipboard (macOS, Wayland/X11 Linux, Windows)
--list WHAT           themes, backdrops, fonts, languages, presets
```

</details>

<details>
<summary><b>FAQ</b></summary>

**Does my code leave my machine?**
No. There is no network code in the binary at all. Highlighting, layout and rasterizing all happen in-process.

**How does it know what I piped in?**
It looks: <code>diff --git</code> headers, <code>commit</code> lines, <code>$</code> prompts, valid JSON, tree branches, <code>KEY=VALUE</code> lines, an HTTP request line, pass/fail marks, or timestamps and log levels. Anything else is treated as code. <code>--preset</code> or <code>--lang</code> always wins.

**Which languages are supported?**
JavaScript, TypeScript, JSX, TSX, Python, Go, Rust, C, C++, Swift, Kotlin, JSON, YAML, SQL, GraphQL, CSS, HTML and Markdown have theme-tuned highlighting. Any other name chroma knows (<code>bash</code>, <code>toml</code>, <code>dockerfile</code>, <code>diff</code>, dozens more) works too.

**Can I use my own font?**
Yes: <code>--font ./MyMono.ttf</code>. A family with bold and italic faces in the same file also gets real bold and italic; the bundled fonts are regular weight only.

**Why is my emoji black and white?**
Symbols and emoji fall back to DejaVu Sans Mono and monochrome Noto Emoji so they never render as boxes. Color emoji and CJK text need <code>--font</code> with a suitable file.

**How big is the binary, and why?**
About 16 MB: the highlighter's lexers, the resvg WebAssembly rasterizer, and the fonts. That is the price of needing nothing else installed.

**A log line was 400 characters. Why isn't my image 400 characters wide?**
Long lines soft-wrap so the card is at most 768px wide, the width of a comfortable 80-column terminal. Raise the cap with <code>--max-width</code>, fix the width with <code>--width</code>, or pass <code>--max-width 0</code> to let the card grow to the longest line.

**How long does a render take?**
Roughly a second for a typical snippet at 2x. Bigger scales cost more pixels.

**Are the sample snippets safe to share?**
They are fictional. Any credential-looking value in them is a placeholder.

</details>

## Build from source

```sh
git clone <this repo> codeshot && cd codeshot
go build -o codeshot ./cmd/codeshot
go test ./...                     # the full suite renders every preset in every theme
./tools/gen-readme-images.sh      # regenerates every image on this page with the tool itself
```

Go 1.25 or newer. The fonts are already in the tree; <code>tools/fetch-fonts.sh</code> re-downloads them from their upstream releases if you ever need to.

## Credits

Highlighting by [chroma](https://github.com/alecthomas/chroma). Rasterizing by [resvg](https://github.com/RazrFalcon/resvg) through [wazero](https://wazero.io). Fonts: Cascadia Code, JetBrains Mono, Fira Code, Geist Mono, IBM Plex Mono, Source Code Pro, Space Mono, Inter and Noto Emoji under the SIL Open Font License, and DejaVu Sans Mono under the DejaVu license; each license ships next to its font under <code>assets/fonts/</code>.
