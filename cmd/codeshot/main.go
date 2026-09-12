// Command codeshot turns code into a styled PNG or SVG image, entirely
// offline.
//
//	codeshot main.go                      # main-go.png, language from the extension
//	git diff | codeshot                   # piped content picks its own preset (diff here)
//	cat q.sql | codeshot --lang sql -o q.png
//	codeshot --preset git-diff --sample -o diff.svg
//	codeshot --theme github --bg paper --no-shadow snippet.js
//	codeshot --copy notes.md             # PNG straight to the clipboard
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"codeshot/internal/clipboard"
	"codeshot/internal/fonts"
	"codeshot/internal/highlight"
	"codeshot/internal/preset"
	"codeshot/internal/render"
	"codeshot/internal/settings"
	"codeshot/internal/sniff"
	"codeshot/internal/theme"
)

const version = "0.1.0"

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, "codeshot:", err)
		var ue usageError
		if errors.As(err, &ue) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

type usageError struct{ error }

type options struct {
	output, presetKey, lang, themeID, bg, font, title, label, prompt, method, status, flags, lineNumbers, list string
	padding, fontSize, scale, wrap, width, maxWidth, radius, cardRadius                                        int
	sample, noBG, noChrome, noShadow, noBadge, embedFonts, copy, showVersion                                   bool
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	var o options
	fs := flag.NewFlagSet("codeshot", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&o.output, "output", "", "output file; a .svg extension writes SVG, anything else PNG (default <title-slug>.png)")
	fs.StringVar(&o.output, "o", "", "alias for --output")
	fs.StringVar(&o.presetKey, "preset", preset.Default, "tool preset: "+strings.Join(preset.Keys(), ", "))
	fs.BoolVar(&o.sample, "sample", false, "render the preset's built-in sample instead of reading input")
	fs.StringVar(&o.lang, "lang", "", "language for the code presets (default: preset's, or guessed from the file name)")
	fs.StringVar(&o.themeID, "theme", theme.Default, "syntax theme: "+strings.Join(theme.IDs(), ", ")+", or random")
	fs.StringVar(&o.bg, "bg", theme.DefaultBackdrop, "backdrop: "+strings.Join(theme.BackdropIDs(), ", ")+", or random")
	fs.StringVar(&o.bg, "backdrop", theme.DefaultBackdrop, "alias for --bg")
	fs.StringVar(&o.font, "font", fonts.Default, "code font: "+strings.Join(fonts.IDs(), ", ")+", or a path to a .ttf/.otf file")
	fs.StringVar(&o.title, "title", "", "window title (default: preset's)")
	fs.StringVar(&o.label, "label", "", "caption drawn in a pill above the window")
	fs.IntVar(&o.padding, "padding", -1, "space around the card in px, 0..160 (default 48, 64 for dev-milestone)")
	fs.IntVar(&o.fontSize, "font-size", 15, "code font size in px, 11..28")
	fs.IntVar(&o.radius, "radius", 0, "corner radius of the backdrop in px (the PNG gets transparent corners)")
	fs.IntVar(&o.cardRadius, "card-radius", settings.DefaultCardRadius, "corner radius of the window in px")
	fs.BoolVar(&o.noBG, "no-bg", false, "no backdrop: a transparent PNG with just the window and its shadow (same as --bg none)")
	fs.BoolVar(&o.noBG, "transparent", false, "alias for --no-bg")
	fs.BoolVar(&o.noChrome, "no-chrome", false, "hide the window title bar")
	fs.BoolVar(&o.noShadow, "no-shadow", false, "no drop shadow")
	fs.StringVar(&o.lineNumbers, "line-numbers", "", "true or false (default: on for code and log presets)")
	fs.StringVar(&o.prompt, "prompt", "", "terminal presets: prompt character shown ($, ❯, #, ~; empty keeps the original)")
	fs.StringVar(&o.method, "method", "GET", "api preset badge: "+strings.Join(settings.Methods, ", "))
	fs.StringVar(&o.status, "status", "200", "api preset badge: "+strings.Join(settings.Statuses, ", "))
	fs.BoolVar(&o.noBadge, "no-badge", false, "api preset: hide the method/status badge")
	fs.StringVar(&o.flags, "flags", "", "regex preset: flags shown in the badge (default gi)")
	fs.IntVar(&o.scale, "scale", 2, "PNG pixel ratio, 1..8")
	fs.IntVar(&o.wrap, "wrap", 0, "soft-wrap lines at this many columns (0 = fit the longest line)")
	fs.IntVar(&o.width, "width", 0, "fixed card width in px, wrapping to fit (0 = fit content)")
	fs.IntVar(&o.maxWidth, "max-width", settings.DefaultMaxWidth, "wrap long lines so the card is at most this wide in px (0 = unlimited)")
	fs.BoolVar(&o.embedFonts, "embed-fonts", true, "SVG output: embed the fonts as data URIs")
	fs.BoolVar(&o.copy, "copy", false, "copy the PNG to the clipboard")
	fs.StringVar(&o.list, "list", "", "print options and exit: themes, backdrops, fonts, languages, presets")
	fs.BoolVar(&o.showVersion, "version", false, "print the version and exit")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: codeshot [flags] [FILE]\n\nReads FILE, or stdin when FILE is omitted or \"-\".\n\n")
		fs.PrintDefaults()
	}
	positional, err := parseInterspersed(fs, args)
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { seen[f.Name] = true })

	switch {
	case o.showVersion:
		fmt.Fprintln(stdout, "codeshot", version)
		return nil
	case o.list != "":
		return list(stdout, o.list)
	case len(positional) > 1:
		return usageError{fmt.Errorf("expected at most one input file, got %d", len(positional))}
	}

	p, ok := preset.Get(o.presetKey)
	if !ok {
		return usageError{fmt.Errorf("unknown preset %q (one of %s)", o.presetKey, preset.String())}
	}

	file := ""
	if len(positional) == 1 && positional[0] != "-" {
		file = positional[0]
	}
	src, err := readInput(p, o.sample, file, len(positional) == 1, stdin)
	if err != nil {
		return err
	}
	// With no preset or language asked for, and no file extension to go by,
	// look at the content: a diff, a git log, a terminal session, JSON, a
	// tree, an .env file, HTTP, test output or a log each get their preset.
	guessed := ""
	if file != "" {
		guessed = highlight.DetectLanguage(file)
	}
	if !seen["preset"] && !seen["lang"] && !o.sample && (guessed == "" || guessed == "plaintext") {
		p = sniff.Lookup(src, p)
	}
	s := settings.Defaults(p)
	s.Theme, s.Backdrop, s.Font = o.themeID, o.bg, o.font
	if o.bg == "random" {
		s.Backdrop = pickRandom(theme.BackdropIDs(), "none")
		fmt.Fprintln(os.Stderr, "backdrop:", s.Backdrop)
	}
	if o.themeID == "random" {
		s.Theme = pickRandom(theme.IDs())
		fmt.Fprintln(os.Stderr, "theme:", s.Theme)
	}
	if seen["title"] {
		s.Title = o.title
	}
	if o.padding >= 0 {
		s.Padding = o.padding
	}
	s.FontSize = o.fontSize
	s.Radius, s.CardRadius = o.radius, o.cardRadius
	s.Label = o.label
	s.ShowBackground = !o.noBG
	s.ShowChrome = s.ShowChrome && !o.noChrome
	s.Shadow = !o.noShadow
	if seen["line-numbers"] {
		v, err := strconv.ParseBool(o.lineNumbers)
		if err != nil {
			return usageError{fmt.Errorf("--line-numbers wants true or false")}
		}
		s.ShowLineNumbers = v
	}
	if seen["prompt"] {
		s.Prompt = o.prompt
	}
	s.Method, s.Status, s.ShowBadge = o.method, o.status, !o.noBadge
	if seen["flags"] {
		s.Flags = o.flags
	}
	s.Scale, s.Wrap, s.Width, s.MaxWidth = o.scale, o.wrap, o.width, o.maxWidth

	switch {
	case seen["lang"]:
		s.Language = o.lang
	case p.Key == preset.Default && guessed != "":
		s.Language = guessed
	}
	if p.Render == preset.Prism && !highlight.KnownLanguage(s.Language) {
		return usageError{fmt.Errorf("unknown language %q (see --list languages)", s.Language)}
	}
	if file != "" && !seen["title"] && p.Key == preset.Default {
		s.Title = filepath.Base(file)
	}
	card, err := render.Build(s, src)
	if err != nil {
		return err
	}

	out := o.output
	wantSVG := strings.EqualFold(filepath.Ext(out), ".svg")
	if out == "" && !o.copy {
		out = settings.Slug(s.Title, p.Key) + ".png"
	}
	if wantSVG {
		if err := os.WriteFile(out, card.SVG(o.embedFonts), 0o644); err != nil {
			return err
		}
		fmt.Fprintln(stdout, out)
		if o.copy {
			return errors.New("--copy needs a PNG; the SVG was written")
		}
		return nil
	}
	data, err := card.PNG(s.Scale)
	if err != nil {
		return err
	}
	if out == "" {
		tmp, err := os.CreateTemp("", "codeshot-*.png")
		if err != nil {
			return err
		}
		out = tmp.Name()
		tmp.Close()
		defer os.Remove(out)
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		return err
	}
	if o.output != "" || !o.copy {
		fmt.Fprintln(stdout, out)
	}
	if o.copy {
		cp, err := clipboard.Detect()
		if err != nil {
			return err
		}
		if err := cp.CopyPNG(out); err != nil {
			return err
		}
		w, h := card.Size(s.Scale)
		fmt.Fprintf(stdout, "copied %dx%d PNG to the clipboard via %s\n", w, h, cp.Name)
	}
	return nil
}

// pickRandom returns a random id from ids, skipping any in exclude.
func pickRandom(ids []string, exclude ...string) string {
	var pool []string
	for _, id := range ids {
		if !slices.Contains(exclude, id) {
			pool = append(pool, id)
		}
	}
	return pool[rand.IntN(len(pool))]
}

// parseInterspersed parses flags that appear before or after positional
// arguments (Go's flag package stops at the first positional), so
// "codeshot main.go -o out.png" works as people expect. "--" ends flag parsing.
func parseInterspersed(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		args = fs.Args()
		if len(args) == 0 {
			return positional, nil
		}
		if len(args) > 0 && strings.HasPrefix(args[0], "-") && args[0] != "-" {
			// fs.Parse consumed a "--"; everything after it is positional.
			return append(positional, args...), nil
		}
		positional = append(positional, args[0])
		args = args[1:]
	}
}

func readInput(p preset.Preset, sample bool, file string, hasArg bool, stdin io.Reader) (string, error) {
	switch {
	case sample:
		return p.Sample()
	case file != "":
		b, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	if !hasArg {
		if f, ok := stdin.(*os.File); ok {
			if st, err := f.Stat(); err == nil && st.Mode()&os.ModeCharDevice != 0 {
				return "", usageError{errors.New("no input: pass a file, pipe to stdin, or use --sample")}
			}
		}
	}
	b, err := io.ReadAll(stdin)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func list(w io.Writer, what string) error {
	switch what {
	case "themes":
		for _, t := range theme.All() {
			fmt.Fprintf(w, "%-15s %s (%s)\n", t.ID, t.Label, t.Mode)
		}
		fmt.Fprintf(w, "%-15s one of the above, chosen for you\n", "random")
	case "backdrops":
		for _, b := range theme.Backdrops() {
			fmt.Fprintf(w, "%-10s %s\n", b.ID, b.Label)
		}
		fmt.Fprintf(w, "%-10s one of the above except none, chosen for you\n", "random")
	case "fonts":
		for _, f := range fonts.Registry {
			fmt.Fprintf(w, "%-10s %s\n", f.ID, f.Label)
		}
	case "languages":
		for _, l := range highlight.Languages {
			fmt.Fprintf(w, "%-11s %s\n", l.ID, l.Label)
		}
		fmt.Fprintln(w, "(any other chroma lexer name works too, e.g. bash, diff, toml)")
	case "presets":
		for _, p := range preset.All {
			fmt.Fprintf(w, "%-18s %-9s %s\n", p.Key, p.Render, p.Title)
		}
	default:
		return usageError{fmt.Errorf("--list wants themes, backdrops, fonts, languages or presets")}
	}
	return nil
}
