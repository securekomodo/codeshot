package layout

import "strings"

// TabSize is the studio's CSS tab-size.
const TabSize = 2

// PrepareText splits source into logical lines:
// CRLF normalized, exactly one trailing newline dropped, and tabs expanded
// to TabSize-column stops (resvg collapses tabs, so they cannot survive
// into the SVG).
func PrepareText(src string) []string {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src = strings.TrimSuffix(src, "\n")
	lines := strings.Split(src, "\n")
	for i, l := range lines {
		if strings.ContainsRune(l, '\t') {
			lines[i] = expandTabs(l, TabSize)
		}
	}
	return lines
}

func expandTabs(s string, tab int) string {
	var b strings.Builder
	col := 0
	for _, r := range s {
		if r == '\t' {
			n := tab - col%tab
			b.WriteString(strings.Repeat(" ", n))
			col += n
			continue
		}
		b.WriteRune(r)
		col++
	}
	return b.String()
}
