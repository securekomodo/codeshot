package theme

import "testing"

func TestRegistry(t *testing.T) {
	want := []string{"dracula", "nightOwl", "oneDark", "palenight", "oceanicNext", "shadesOfPurple",
		"vsDark", "okaidia", "gruvboxDark", "github", "oneLight", "nightOwlLight", "kali"}
	got := IDs()
	if len(got) != len(want) {
		t.Fatalf("themes = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("theme %d = %s want %s", i, got[i], want[i])
		}
	}
	if th, _ := Get("github"); !th.Light() || th.Window != "#ffffff" {
		t.Errorf("github: light=%v window=%s", th.Light(), th.Window)
	}
	if ids := BackdropIDs(); len(ids) != 11 || ids[0] != "ember" || ids[10] != "none" {
		t.Errorf("backdrops = %v", ids)
	}
}

func TestResolveAndMerge(t *testing.T) {
	vs, _ := Get("vsDark")
	js := vs.Resolve("javascript").StyleFor("tag")
	markup := vs.Resolve("markup").StyleFor("tag")
	if js.Color != "rgb(78, 201, 176)" || markup.Color != "rgb(86, 156, 214)" {
		t.Errorf("vsDark tag: js=%q markup=%q", js.Color, markup.Color)
	}
	if p := vs.Resolve("MARKUP").StyleFor("punctuation"); p.Color != "#808080" {
		t.Errorf("language match should be case-insensitive, got %q", p.Color)
	}

	dr, _ := Get("dracula")
	r := dr.Resolve("javascript")
	kw := r.StyleFor("keyword")
	if kw.Color != "rgb(189, 147, 249)" || !kw.Italic() {
		t.Errorf("dracula keyword = %+v", kw)
	}
	// Later types override earlier ones, field by field.
	m := r.StyleFor("keyword", "string")
	if m.Color != "rgb(255, 121, 198)" || !m.Italic() {
		t.Errorf("merge keyword+string = %+v", m)
	}
	if s := r.StyleFor("nonexistent"); s != (Style{}) {
		t.Errorf("unknown type should be empty, got %+v", s)
	}

	pn, _ := Get("palenight")
	if k := pn.Resolve("go").StyleFor("keyword"); k.Color != "" || !k.Italic() {
		t.Errorf("palenight keyword should be italic with inherited color, got %+v", k)
	}
	ok, _ := Get("okaidia")
	if a := ok.Resolve("markup").StyleFor("attr-name"); a.Color != "#a6e22e" {
		t.Errorf("okaidia attr-name should have !important stripped, got %q", a.Color)
	}
	oc, _ := Get("oceanicNext")
	if ns := oc.Resolve("go").StyleFor("namespace"); !ns.HasOpacity || ns.Opacity != 0.7 {
		t.Errorf("oceanicNext namespace opacity = %+v", ns)
	}
	if b := (Style{FontWeight: "600"}); !b.Bold() || (Style{FontWeight: "400"}).Bold() || !(Style{FontWeight: "bold"}).Bold() {
		t.Error("Bold()")
	}
}

func TestParseColor(t *testing.T) {
	cases := map[string]Color{
		"#F8F8F2":                 {"#f8f8f2", 1},
		"#abc":                    {"#aabbcc", 1},
		"#00000080":               {"#000000", 128.0 / 255},
		"rgb(189, 147, 249)":      {"#bd93f9", 1},
		"rgba(239, 83, 80, 0.56)": {"#ef5350", 0.56},
		"rgba(255,255,255,0.45)":  {"#ffffff", 0.45},
		"hsl(220, 14%, 71%)":      {"#abb2bf", 1},
		"hsl(230, 1%, 98%)":       {"#fafafa", 1},
		"hsl(0, 0%, 100%)":        {"#ffffff", 1},
		"transparent":             {"#000000", 0},
	}
	for in, want := range cases {
		got, err := ParseColor(in)
		if err != nil {
			t.Errorf("%s: %v", in, err)
			continue
		}
		if got.Hex != want.Hex || abs(got.Alpha-want.Alpha) > 1e-9 {
			t.Errorf("%s = %+v want %+v", in, got, want)
		}
	}
	if _, err := ParseColor("blue"); err == nil {
		t.Error("named colors other than transparent should fail")
	}
}

func TestBackdrops(t *testing.T) {
	em, _ := GetBackdrop("ember")
	if !em.Gradient() || em.Angle != 135 || len(em.Stops) != 3 || em.Stops[1].Offset != 0.5 || em.Stops[2].Color.Hex != "#7a2e3a" {
		t.Errorf("ember = %+v", em)
	}
	dk, _ := GetBackdrop("darkroom")
	if dk.Angle != 160 || len(dk.Stops) != 2 || dk.Stops[1].Offset != 1 {
		t.Errorf("darkroom = %+v", dk)
	}
	ink, _ := GetBackdrop("ink")
	if ink.Gradient() || ink.Transparent || ink.Solid.Hex != "#0b0d10" {
		t.Errorf("ink = %+v", ink)
	}
	none, _ := GetBackdrop("none")
	if !none.Transparent {
		t.Errorf("none = %+v", none)
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
