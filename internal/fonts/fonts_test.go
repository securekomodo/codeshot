package fonts

import "testing"

func TestBundledFacesAreMonospace(t *testing.T) {
	for _, in := range Registry {
		f, err := Load(in.ID)
		if err != nil {
			t.Fatalf("%s: %v", in.ID, err)
		}
		if f.Family == "" || len(f.Families) == 0 {
			t.Errorf("%s: no family name", in.ID)
		}
		zero, em := f.Advance('0', 15), f.Advance('M', 15)
		if zero <= 0 || zero != em {
			t.Errorf("%s (%s): advance '0'=%v 'M'=%v, want equal and > 0", in.ID, f.Family, zero, em)
		}
		if f.Width("ab", 15) != 2*zero {
			t.Errorf("%s: Width(ab)=%v want %v", in.ID, f.Width("ab", 15), 2*zero)
		}
		if f.Advance('\U0010ffff', 15) != zero {
			t.Errorf("%s: missing glyph should fall back to the advance of '0'", in.ID)
		}
		a, d := f.Metrics(15)
		if a <= 0 || d <= 0 || a < d {
			t.Errorf("%s: metrics ascent=%v descent=%v", in.ID, a, d)
		}
	}
}

func TestFallbacks(t *testing.T) {
	fb, err := Fallbacks()
	if err != nil {
		t.Fatal(err)
	}
	if len(fb) != 2 || !fb[0].Has('✔') || !fb[0].Has('○') || !fb[0].Has('├') || !fb[1].Has('🎉') {
		t.Fatalf("fallback coverage: %v", fb)
	}
	code, _ := Load("cascadia")
	if code.Has('✔') {
		t.Skip("cascadia gained ✔; test assumes it is missing")
	}
	code.SetFallbacks(fb...)
	if a := code.Advance('✔', 15); a != fb[0].Advance('✔', 15) || a <= 0 {
		t.Errorf("✔ advance %v should come from DejaVu (%v)", a, fb[0].Advance('✔', 15))
	}
	if a := code.Advance('🎉', 15); a != fb[1].Advance('🎉', 15) || a <= 0 {
		t.Errorf("🎉 advance %v should come from Noto Emoji", a)
	}
}

func TestInterAndErrors(t *testing.T) {
	f, err := Inter()
	if err != nil {
		t.Fatal(err)
	}
	if f.Family != "Inter" {
		t.Errorf("Inter family = %q (families %v)", f.Family, f.Families)
	}
	if _, err := Load("nope"); err == nil {
		t.Error("Load(nope) should fail")
	}
	if _, err := Load("/nonexistent/x.ttf"); err == nil {
		t.Error("Load(missing path) should fail")
	}
}
