package render

import (
	"bytes"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/securekomodo/codeshot/internal/preset"
	"github.com/securekomodo/codeshot/internal/settings"
	"github.com/securekomodo/codeshot/internal/theme"
)

// Every preset with its sample, in every theme, must lay out, serialize and
// rasterize; the PNG must decode to the declared size.
func TestEveryPresetEveryTheme(t *testing.T) {
	if testing.Short() {
		t.Skip("rasterizing 200+ cards")
	}
	dir := t.TempDir()
	themes := theme.IDs()
	for i, p := range preset.All {
		sample, err := p.Sample()
		if err != nil {
			t.Fatal(err)
		}
		for j, th := range themes {
			// Rasterize each theme once and each preset once; SVG for the rest.
			s := settings.Defaults(p)
			s.Theme = th
			s.Backdrop = theme.BackdropIDs()[(i+j)%len(theme.BackdropIDs())]
			c, err := Build(s, sample)
			if err != nil {
				t.Fatalf("%s/%s: %v", p.Key, th, err)
			}
			doc := c.SVG(false)
			if !bytes.HasPrefix(doc, []byte("<svg")) || !bytes.HasSuffix(bytes.TrimSpace(doc), []byte("</svg>")) {
				t.Errorf("%s/%s: malformed svg", p.Key, th)
			}
			if i != j && j != 0 {
				continue
			}
			data, err := c.PNG(2)
			if err != nil {
				t.Fatalf("%s/%s: %v", p.Key, th, err)
			}
			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("%s/%s: %v", p.Key, th, err)
			}
			w, h := c.Size(2)
			if b := img.Bounds(); b.Dx() != w || b.Dy() != h {
				t.Errorf("%s/%s: png %v want %dx%d", p.Key, th, b, w, h)
			}
			if err := os.WriteFile(filepath.Join(dir, p.Key+"-"+th+".png"), data, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestTransparentAndSolidCorners(t *testing.T) {
	p, _ := preset.Get("code")
	s := settings.Defaults(p)
	s.Backdrop = "none"
	c, err := Build(s, "x")
	if err != nil {
		t.Fatal(err)
	}
	data, err := c.PNG(1)
	if err != nil {
		t.Fatal(err)
	}
	img, _ := png.Decode(bytes.NewReader(data))
	if _, _, _, a := img.At(0, 0).RGBA(); a != 0 {
		t.Errorf("none backdrop: corner alpha %d want 0", a)
	}
	s.Backdrop = "ink"
	c, _ = Build(s, "x")
	data, _ = c.PNG(1)
	img, _ = png.Decode(bytes.NewReader(data))
	r, g, b, _ := img.At(0, 0).RGBA()
	if r>>8 != 0x0b || g>>8 != 0x0d || b>>8 != 0x10 {
		t.Errorf("ink backdrop: corner = %x %x %x", r>>8, g>>8, b>>8)
	}
	if !strings.Contains(string(c.SVG(true)), "@font-face") {
		t.Error("SVG(true) should embed fonts")
	}
}

func TestBuildErrors(t *testing.T) {
	p, _ := preset.Get("code")
	s := settings.Defaults(p)
	s.Font = "nope"
	if _, err := Build(s, "x"); err == nil {
		t.Error("unknown font should fail")
	}
	s = settings.Defaults(p)
	s.Theme = "nope"
	if _, err := Build(s, "x"); err == nil {
		t.Error("unknown theme should fail")
	}
}
