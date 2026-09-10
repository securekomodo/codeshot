#!/usr/bin/env bash
# Fetches the Regular TTF of each bundled code font (plus Inter Medium for the
# title bar) and the matching SIL OFL license into assets/fonts/<id>/.
# Static Regular files are preferred; a Google Fonts variable file (default
# instance = Regular) is the fallback. Run once; the results are committed.
set -euo pipefail
cd "$(dirname "$0")/.."
G=https://raw.githubusercontent.com/google/fonts/main/ofl
TMP=$(mktemp -d); trap 'rm -rf "$TMP"' EXIT

get() { # get <dest> <url>...  (first URL that yields a real TTF wins)
  local dest=$1; shift
  for u in "$@"; do
    if curl -fsSL --retry 3 -o "$TMP/f" "$u" && head -c4 "$TMP/f" | grep -q $'^\x00\x01\x00\x00\|^true\|^OTTO'; then
      mv "$TMP/f" "$dest"; echo "ok   $dest  <- $u"; return 0; fi
  done; echo "FAIL $dest"; return 1
}
lic() { mkdir -p "assets/fonts/$1"; curl -fsSL --retry 3 -o "assets/fonts/$1/OFL.txt" "$G/$2/OFL.txt"; }
unz() { # unz <dest> <zip-url> <member>
  curl -fsSL --retry 3 -o "$TMP/z.zip" "$2" && unzip -p "$TMP/z.zip" "$3" > "$1" && echo "ok   $1  <- $2!$3"
}

lic cascadia cascadiacode
unz assets/fonts/cascadia/CascadiaCode-Regular.ttf \
  https://github.com/microsoft/cascadia-code/releases/download/v2404.23/CascadiaCode-2404.23.zip \
  ttf/static/CascadiaCode-Regular.ttf \
 || get assets/fonts/cascadia/CascadiaCode-Regular.ttf "$G/cascadiacode/CascadiaCode%5Bwght%5D.ttf"

lic jetbrains jetbrainsmono
get assets/fonts/jetbrains/JetBrainsMono-Regular.ttf \
  https://github.com/JetBrains/JetBrainsMono/raw/master/fonts/ttf/JetBrainsMono-Regular.ttf \
  "$G/jetbrainsmono/JetBrainsMono%5Bwght%5D.ttf"

lic fira firacode
unz assets/fonts/fira/FiraCode-Regular.ttf \
  https://github.com/tonsky/FiraCode/releases/download/6.2/Fira_Code_v6.2.zip ttf/FiraCode-Regular.ttf \
 || get assets/fonts/fira/FiraCode-Regular.ttf "$G/firacode/FiraCode%5Bwght%5D.ttf"

lic geist geistmono
get assets/fonts/geist/GeistMono-Regular.ttf \
  https://github.com/vercel/geist-font/raw/main/packages/next/dist/fonts/geist-mono/GeistMono-Regular.ttf \
  "$G/geistmono/GeistMono%5Bwght%5D.ttf"

lic ibm ibmplexmono
get assets/fonts/ibm/IBMPlexMono-Regular.ttf \
  https://github.com/IBM/plex/raw/master/packages/plex-mono/fonts/complete/ttf/IBMPlexMono-Regular.ttf \
  "$G/ibmplexmono/IBMPlexMono-Regular.ttf"

lic source sourcecodepro
get assets/fonts/source/SourceCodePro-Regular.ttf \
  https://github.com/adobe-fonts/source-code-pro/raw/release/TTF/SourceCodePro-Regular.ttf \
  "$G/sourcecodepro/SourceCodePro%5Bwght%5D.ttf"

lic space spacemono
get assets/fonts/space/SpaceMono-Regular.ttf "$G/spacemono/SpaceMono-Regular.ttf"

lic inter inter
unz assets/fonts/inter/Inter-Medium.ttf \
  https://github.com/rsms/inter/releases/download/v4.1/Inter-4.1.zip extras/ttf/Inter-Medium.ttf \
 || get assets/fonts/inter/Inter-Medium.ttf "$G/inter/Inter%5Bopsz,wght%5D.ttf"

# Fallback faces for glyphs the code fonts lack: DejaVu Sans Mono for symbols
# such as ✓ ✗ ○ (Bitstream Vera / DejaVu license), Noto Emoji for monochrome
# emoji (OFL). resvg falls back to them per glyph.
mkdir -p assets/fonts/dejavu
unz assets/fonts/dejavu/DejaVuSansMono.ttf \
  https://github.com/dejavu-fonts/dejavu-fonts/releases/download/version_2_37/dejavu-fonts-ttf-2.37.zip \
  dejavu-fonts-ttf-2.37/ttf/DejaVuSansMono.ttf
curl -fsSL --retry 3 -o "$TMP/z.zip" https://github.com/dejavu-fonts/dejavu-fonts/releases/download/version_2_37/dejavu-fonts-ttf-2.37.zip \
  && unzip -p "$TMP/z.zip" dejavu-fonts-ttf-2.37/LICENSE > assets/fonts/dejavu/LICENSE
lic emoji notoemoji
get assets/fonts/emoji/NotoEmoji-Regular.ttf "$G/notoemoji/NotoEmoji%5Bwght%5D.ttf"

ls -la assets/fonts/*/
