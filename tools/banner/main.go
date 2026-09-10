//go:build ignore

// Renders the README banner with the tool's own rasterizer and fonts.
//
//	go run tools/banner/main.go -o docs/banner.png
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"codeshot/internal/fonts"
	"codeshot/internal/raster"
)

const (
	w, h     = 960, 300
	ink      = "#0b0d10"
	panel    = "#14171c"
	line     = "#232830"
	text     = "#e6e8eb"
	muted    = "#8a929e"
	amber    = "#f4a259"
	teal     = "#5fb3a1"
	eyebrow  = "CODE  ·  18+ LANGUAGES  ·  PNG + SVG  ·  ONE BINARY"
	headLead = "Turn code into a "
	headAcc  = "beautiful image"
	headTail = " in seconds."
	sub      = "Real editor themes, ligature-ready fonts, a macOS-style frame, and a crisp export."
	sub2     = "From your terminal. Nothing leaves your machine."
)

func main() {
	out := flag.String("o", "docs/banner.png", "output PNG")
	scale := flag.Int("scale", 2, "pixel ratio")
	flag.Parse()

	inter, err := fonts.Inter()
	must(err)
	mono, err := fonts.Load("jetbrains")
	must(err)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, w**scale, h**scale, w, h)
	b.WriteString(`<defs>
<linearGradient id="glow" gradientUnits="userSpaceOnUse" x1="960" y1="300" x2="520" y2="0">
<stop offset="0" stop-color="#f4a259" stop-opacity="0.16"/><stop offset="0.6" stop-color="#c75c3c" stop-opacity="0.05"/><stop offset="1" stop-color="#7a2e3a" stop-opacity="0"/></linearGradient>
<clipPath id="r"><rect width="960" height="300" rx="18"/></clipPath>
</defs>`)
	fmt.Fprintf(&b, `<g clip-path="url(#r)"><rect width="960" height="300" fill="%s"/>`, ink)
	b.WriteString(`<rect width="960" height="300" fill="url(#glow)"/>`)
	// A faint card silhouette on the right, like a window peeking in.
	fmt.Fprintf(&b, `<rect x="700" y="170" width="420" height="220" rx="14" fill="%s" stroke="%s"/>`, panel, line)
	for i, c := range []string{"#ff5f57", "#febc2e", "#28c840"} {
		fmt.Fprintf(&b, `<circle cx="%d" cy="192" r="6" fill="%s"/>`, 722+i*20, c)
	}
	rows := [][2]string{{amber, "const"}, {text, " shot"}, {muted, " = "}, {teal, "render"}, {text, "(code)"}}
	fmt.Fprintf(&b, `<text x="722" y="238" font-family="'%s'" font-size="15" xml:space="preserve">`, mono.Family)
	for _, r := range rows {
		fmt.Fprintf(&b, `<tspan fill="%s">%s</tspan>`, r[0], r[1])
	}
	b.WriteString(`</text>`)
	fmt.Fprintf(&b, `<text x="722" y="264" font-family="'%s'" font-size="15" fill="%s" xml:space="preserve">export png, svg, clipboard</text>`, mono.Family, muted)
	fmt.Fprintf(&b, `<text x="722" y="290" font-family="'%s'" font-size="15" fill="%s" xml:space="preserve">// no browser, no server</text>`, mono.Family, "#6272a4")
	b.WriteString(`</g>`)
	fmt.Fprintf(&b, `<rect x="0.5" y="0.5" width="959" height="299" rx="18" fill="none" stroke="%s"/>`, line)

	fmt.Fprintf(&b, `<text x="56" y="86" font-family="'%s'" font-size="12" font-weight="500" fill="%s" letter-spacing="2" xml:space="preserve">%s</text>`, mono.Family, amber, eyebrow)
	fmt.Fprintf(&b, `<text x="56" y="140" font-family="'%s'" font-size="42" font-weight="500" fill="%s" xml:space="preserve">%s<tspan fill="%s">%s</tspan></text>`, inter.Family, text, headLead, amber, headAcc)
	fmt.Fprintf(&b, `<text x="56" y="188" font-family="'%s'" font-size="42" font-weight="500" fill="%s" xml:space="preserve">%s</text>`, inter.Family, text, strings.TrimSpace(headTail))
	fmt.Fprintf(&b, `<text x="56" y="232" font-family="'%s'" font-size="15" font-weight="500" fill="%s" xml:space="preserve">%s</text>`, inter.Family, muted, sub)
	fmt.Fprintf(&b, `<text x="56" y="256" font-family="'%s'" font-size="15" font-weight="500" fill="%s" xml:space="preserve">%s</text>`, inter.Family, muted, sub2)
	b.WriteString(`</svg>`)

	png, err := raster.PNG([]byte(b.String()), [][]byte{inter.Data, mono.Data}, uint32(w**scale), uint32(h**scale))
	must(err)
	must(os.WriteFile(*out, png, 0o644))

	fmt.Println(*out)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
