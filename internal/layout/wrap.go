package layout

import (
	"unicode/utf8"

	"github.com/securekomodo/codeshot/internal/highlight"
)

// wrapped is one visual row: its spans and the logical line number it
// starts (0 for a continuation row, which gets no gutter number).
type wrapped struct {
	line   highlight.Line
	number int
}

// wrapAll soft-wraps every logical line at cols columns (0 = no wrapping).
func wrapAll(lines []highlight.Line, cols int) []wrapped {
	var out []wrapped
	for i, l := range lines {
		if cols <= 0 || utf8.RuneCountInString(l.Text()) <= cols {
			out = append(out, wrapped{l, i + 1})
			continue
		}
		for j, part := range wrapLine(l, cols) {
			n := 0
			if j == 0 {
				n = i + 1
			}
			out = append(out, wrapped{part, n})
		}
	}
	return out
}

// wrapLine breaks a line into rows of at most cols runes, preferring to
// break after the last space that fits (CSS pre-wrap with break-word), and
// breaking mid-word only when a word is longer than the row.
func wrapLine(l highlight.Line, cols int) []highlight.Line {
	runes := []rune(l.Text())
	var breaks []int
	start := 0
	for len(runes)-start > cols {
		cut := -1
		for k := start + cols; k > start; k-- {
			if runes[k-1] == ' ' {
				cut = k
				break
			}
		}
		if cut <= start {
			cut = start + cols
		}
		breaks = append(breaks, cut)
		start = cut
	}
	return splitAt(l, breaks)
}

// splitAt cuts the line's spans at the given rune offsets.
func splitAt(l highlight.Line, breaks []int) []highlight.Line {
	var rows []highlight.Line
	var cur highlight.Line
	pos := 0
	bi := 0
	for _, sp := range l {
		text := []rune(sp.Text)
		for len(text) > 0 {
			if bi < len(breaks) && breaks[bi] <= pos {
				rows = append(rows, cur)
				cur = nil
				bi++
				continue
			}
			take := len(text)
			if bi < len(breaks) && pos+take > breaks[bi] {
				take = breaks[bi] - pos
			}
			piece := sp
			piece.Text = string(text[:take])
			cur = append(cur, piece)
			text = text[take:]
			pos += take
		}
	}
	for bi < len(breaks) {
		rows = append(rows, cur)
		cur = nil
		bi++
	}
	return append(rows, cur)
}
