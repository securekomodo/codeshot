// Package assets holds the files embedded into the binary: the bundled fonts
// and their SIL Open Font License texts.
package assets

import "embed"

// FS contains fonts/<id>/<Family>-Regular.ttf and fonts/<id>/OFL.txt for
// every bundled family, plus fonts/inter/Inter-Medium.ttf for the title bar.
//
//go:embed fonts
var FS embed.FS
