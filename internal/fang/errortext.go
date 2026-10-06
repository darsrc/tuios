package fang

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// RenderErrorText renders s in style, wrapping it at the style's width.
//
// lipgloss wraps a word longer than the width by cutting it, and the words
// that run long in an error are paths, socket names and commands. A value cut
// in two cannot be copied back out of the terminal, and a test or a person
// searching the output for it does not find it. So a word longer than the
// width stays whole when it is the only thing on its line, a lone value the
// reader copies back out, and the terminal soft-wraps it. When its line
// carries other words, the word is split across lines so none of them runs
// past the width.
//
// Each line is rendered on its own. lipgloss pads every line of a block to the
// block's widest, so one long path would pad every other line past the
// terminal's edge too, and each would wrap onto a blank row.
func RenderErrorText(style lipgloss.Style, s string) string {
	// A transform, such as fang's capital on the first word, is for the text
	// as a whole, not for each line it wraps onto.
	if transform := style.GetTransform(); transform != nil {
		s = transform(s)
		style = style.UnsetTransform()
	}
	limit := style.GetWidth() - style.GetHorizontalPadding() - style.GetHorizontalBorderSize()
	if limit > 0 {
		s = wrapWords(s, limit)
	}
	style = style.UnsetWidth()
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = style.Render(line)
	}
	return strings.Join(lines, "\n")
}

// wrapWords wraps every line of s to limit cells, breaking at spaces and
// splitting a word that runs past the limit. A line's leading indent is kept
// on the lines it wraps onto.
func wrapWords(s string, limit int) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = wrapLine(line, limit)
	}
	return strings.Join(lines, "\n")
}

func wrapLine(line string, limit int) string {
	if ansi.StringWidth(line) <= limit {
		return line
	}
	body := strings.TrimLeft(line, " ")
	indent := line[:len(line)-len(body)]
	// A line that is a single token is a value the reader copies back out of
	// the terminal, like a path or a socket name. Cutting it is the bug this
	// renderer exists to avoid, so it stays whole and the terminal soft-wraps it.
	if !strings.Contains(body, " ") {
		return line
	}
	chunkW := limit - ansi.StringWidth(indent)
	var (
		lines []string
		cur   strings.Builder
	)
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		lines = append(lines, indent+cur.String())
		cur.Reset()
	}
	for word := range strings.FieldsSeq(body) {
		w := ansi.StringWidth(word)
		if w > chunkW {
			// The word is wider than the line, and the line carries other words,
			// so it is not a lone value: split it, each piece on its own line.
			flush()
			for _, piece := range splitWord(word, chunkW) {
				lines = append(lines, indent+piece)
			}
			continue
		}
		if cur.Len() > 0 && ansi.StringWidth(cur.String())+1+w > chunkW {
			flush()
		}
		if cur.Len() > 0 {
			cur.WriteString(" ")
		}
		cur.WriteString(word)
	}
	flush()
	return strings.Join(lines, "\n")
}

// splitWord breaks s into pieces of at most width cells, on rune boundaries,
// so a long token can be split across lines without cutting a multibyte rune.
func splitWord(s string, width int) []string {
	if width <= 0 {
		return []string{s}
	}
	var (
		pieces []string
		cur    strings.Builder
		w      int
	)
	for _, r := range s {
		rw := ansi.StringWidth(string(r))
		if w+rw > width {
			pieces = append(pieces, cur.String())
			cur.Reset()
			w = 0
		}
		cur.WriteRune(r)
		w += rw
	}
	if cur.Len() > 0 {
		pieces = append(pieces, cur.String())
	}
	return pieces
}
