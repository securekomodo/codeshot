// Package fonts loads the bundled monospace faces (or a user-supplied TTF/OTF
// file) and measures glyph advances and vertical metrics with x/image/font/sfnt.
package fonts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"

	"codeshot/assets"
)

// Info describes one bundled code font.
type Info struct {
	ID    string
	Label string
	file  string
}

// Registry lists the bundled code fonts in display order.
var Registry = []Info{
	{"cascadia", "Cascadia Code", "fonts/cascadia/CascadiaCode-Regular.ttf"},
	{"jetbrains", "JetBrains Mono", "fonts/jetbrains/JetBrainsMono-Regular.ttf"},
	{"fira", "Fira Code", "fonts/fira/FiraCode-Regular.ttf"},
	{"geist", "Geist Mono", "fonts/geist/GeistMono-Regular.ttf"},
	{"ibm", "IBM Plex Mono", "fonts/ibm/IBMPlexMono-Regular.ttf"},
	{"source", "Source Code Pro", "fonts/source/SourceCodePro-Regular.ttf"},
	{"space", "Space Mono", "fonts/space/SpaceMono-Regular.ttf"},
}

// Default is the font id used when none is requested.
const Default = "cascadia"

const interFile = "fonts/inter/Inter-Medium.ttf"

// fallbackFiles are consulted, in order, for glyphs the code font lacks:
// DejaVu Sans Mono for symbols such as ✓ ✗ ○ and Noto Emoji for emoji.
var fallbackFiles = []string{"fonts/dejavu/DejaVuSansMono.ttf", "fonts/emoji/NotoEmoji-Regular.ttf"}

// IDs returns the bundled font ids in registry order.
func IDs() []string {
	ids := make([]string, len(Registry))
	for i, in := range Registry {
		ids[i] = in.ID
	}
	return ids
}

// Face is a parsed font with a measurement cache. It is safe for concurrent use.
type Face struct {
	ID       string   // bundled id, or "" for an external file
	Family   string   // preferred family name from the name table
	Families []string // Family plus alternate family names, for font-family fallbacks
	Data     []byte   // the raw TTF/OTF bytes, for embedding and for the rasterizer

	font      *sfnt.Font
	fallbacks []*Face
	mu        sync.Mutex
	buf       sfnt.Buffer
	adv       map[advKey]float64
}

type advKey struct {
	r    rune
	size float64
}

// Load returns the bundled face for a font id, or parses the TTF/OTF file at
// the given path.
func Load(idOrPath string) (*Face, error) {
	for _, in := range Registry {
		if in.ID == idOrPath {
			data, err := assets.FS.ReadFile(in.file)
			if err != nil {
				return nil, err
			}
			return parse(in.ID, data)
		}
	}
	switch strings.ToLower(filepath.Ext(idOrPath)) {
	case ".ttf", ".otf":
		data, err := os.ReadFile(idOrPath)
		if err != nil {
			return nil, err
		}
		return parse("", data)
	}
	return nil, fmt.Errorf("unknown font %q: use one of %s, or a path to a .ttf/.otf file",
		idOrPath, strings.Join(IDs(), ", "))
}

// Fallbacks returns the bundled fallback faces (symbols, then emoji).
func Fallbacks() ([]*Face, error) {
	var faces []*Face
	for _, file := range fallbackFiles {
		data, err := assets.FS.ReadFile(file)
		if err != nil {
			return nil, err
		}
		f, err := parse("fallback", data)
		if err != nil {
			return nil, err
		}
		faces = append(faces, f)
	}
	return faces, nil
}

// SetFallbacks makes Advance and Width consult the given faces, in order,
// for runes this face has no glyph for, mirroring the rasterizer's per-glyph
// font fallback.
func (f *Face) SetFallbacks(faces ...*Face) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fallbacks = faces
	f.adv = map[advKey]float64{}
}

// Has reports whether the face has a glyph for r.
func (f *Face) Has(r rune) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	gi, err := f.font.GlyphIndex(&f.buf, r)
	return err == nil && gi != 0
}

// Inter returns the bundled Inter Medium face used for the title bar.
func Inter() (*Face, error) {
	data, err := assets.FS.ReadFile(interFile)
	if err != nil {
		return nil, err
	}
	return parse("inter", data)
}

func parse(id string, data []byte) (*Face, error) {
	f, err := sfnt.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}
	face := &Face{ID: id, Data: data, font: f, adv: map[advKey]float64{}}
	var buf sfnt.Buffer
	for _, nid := range []sfnt.NameID{sfnt.NameIDTypographicFamily, sfnt.NameIDFamily} {
		n, err := f.Name(&buf, nid)
		if err != nil || n == "" || contains(face.Families, n) {
			continue
		}
		face.Families = append(face.Families, n)
	}
	if len(face.Families) == 0 {
		return nil, fmt.Errorf("font has no family name")
	}
	face.Family = face.Families[0]
	return face, nil
}

// Advance returns the horizontal advance of r at the given pixel size.
// Runes the font lacks a glyph for are measured in the fallback faces, and
// failing that are given the advance of '0' so layout stays sane.
func (f *Face) Advance(r rune, size float64) float64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.advance(r, size, true)
}

func (f *Face) advance(r rune, size float64, fallback bool) float64 {
	key := advKey{r, size}
	if a, ok := f.adv[key]; ok {
		return a
	}
	var a float64
	gi, err := f.font.GlyphIndex(&f.buf, r)
	if err == nil && gi != 0 {
		if adv, err := f.font.GlyphAdvance(&f.buf, gi, fixed.Int26_6(size*64), font.HintingNone); err == nil {
			a = float64(adv) / 64
		}
	} else if fallback {
		a = -1
		for _, fb := range f.fallbacks {
			if fb.Has(r) {
				a = fb.Advance(r, size)
				break
			}
		}
		if a < 0 {
			a = 0
			if r != '0' {
				a = f.advance('0', size, false)
			}
		}
	}
	f.adv[key] = a
	return a
}

// Width returns the advance width of s at the given size (no kerning).
func (f *Face) Width(s string, size float64) float64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	var w float64
	for _, r := range s {
		w += f.advance(r, size, true)
	}
	return w
}

// Metrics returns the ascent and descent (both positive, in pixels) at the
// given size, taken from the hhea table as browsers do for line boxes.
func (f *Face) Metrics(size float64) (ascent, descent float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, err := f.font.Metrics(&f.buf, fixed.Int26_6(size*64), font.HintingNone)
	if err != nil {
		return size * 0.8, size * 0.2
	}
	return float64(m.Ascent) / 64, float64(m.Descent) / 64
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
