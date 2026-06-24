// Package clean turns text copied out of a TUI/terminal back into something you
// can paste cleanly. The pipeline strips ANSI, removes box/border chrome,
// dedents terminal padding, then reflows lines in a way that respects whether
// the content is Markdown (preserve structure) or plain text (rejoin prose).
package clean

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/weijt606/TUI-easy-past/internal/detect"
)

// Options controls the cleaning pipeline.
type Options struct {
	// Format forces a content mode. If nil, the format is auto-detected.
	Format *detect.Format
	// NoRejoin disables wrapped-line rejoining (only strips chrome/whitespace).
	NoRejoin bool
	// KeepANSI leaves ANSI escape sequences untouched (off by default).
	KeepANSI bool
}

// Report describes what the cleaner did, for --explain.
type Report struct {
	Format        detect.Format
	DetectScore   float64
	DetectSignals []string
	LeftChrome    string // the left border char that was stripped, if any
	RightChrome   bool
	Dedented      int  // columns removed
	Rejoined      bool // whether reflow ran
}

var (
	// Matches CSI (\x1b[ ... ) and a few other common escape forms.
	reANSI = regexp.MustCompile(`\x1b\[[0-9;:?]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)|\x1b[@-Z\\-_]`)
)

// Clean runs the full pipeline and returns the cleaned text plus a Report.
func Clean(text string, opts Options) (string, Report) {
	var rep Report

	text = normalizeNewlines(text)
	if !opts.KeepANSI {
		text = reANSI.ReplaceAllString(text, "")
	}

	lines := strings.Split(text, "\n")
	lines = removeFullBoxBorders(lines)
	lines, rep.LeftChrome = stripLeftChrome(lines)
	lines, rep.RightChrome = stripRightChrome(lines)

	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}

	lines, rep.Dedented = dedent(lines)

	// Decide format on the cleaned lines (chrome removed) so padding can't fool it.
	joined := strings.Join(lines, "\n")
	if opts.Format != nil {
		rep.Format = *opts.Format
	} else {
		d := detect.Detect(joined)
		rep.Format = d.Format
		rep.DetectScore = d.Score
		rep.DetectSignals = d.Signals
	}

	if !opts.NoRejoin {
		lines = reflow(lines, rep.Format)
		rep.Rejoined = true
	}

	lines = collapseBlankLines(lines)
	out := strings.Join(lines, "\n")
	out = strings.Trim(out, "\n")
	return out, rep
}

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// --- chrome / border removal ---

func isBoxRune(r rune) bool {
	// Box Drawing (U+2500–U+257F) and Block Elements (U+2580–U+259F).
	return (r >= 0x2500 && r <= 0x259F)
}

// removeFullBoxBorders drops horizontal border lines (top/bottom rules and
// corners) made entirely of box-drawing chars. Lines that contain ONLY vertical
// bars and spaces are kept on purpose: they are empty interior rows of a box and
// must survive as blank lines so paragraph breaks aren't lost. stripLeftChrome /
// stripRightChrome later collapse them to "".
func removeFullBoxBorders(lines []string) []string {
	out := lines[:0:0]
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		if t == "" {
			out = append(out, ln)
			continue
		}
		allBox := true
		hasHorizontal := false // a non-bar box rune (─ ═ corners, etc.)
		for _, r := range t {
			if r == ' ' {
				continue
			}
			if !isBoxRune(r) {
				allBox = false
				break
			}
			if !leftBarRunes[r] {
				hasHorizontal = true
			}
		}
		if allBox && hasHorizontal {
			continue // drop the horizontal border line entirely
		}
		out = append(out, ln)
	}
	return out
}

// leftBarRunes are vertical chrome we strip from the left margin. Note '>' is
// deliberately excluded: it is a Markdown blockquote marker.
var leftBarRunes = map[rune]bool{
	'│': true, '┃': true, '┆': true, '┇': true, '┊': true, '┋': true,
	'║': true, '|': true,
}

// stripLeftChrome detects a vertical border char that appears as the first
// non-space rune on most non-empty lines and removes it (plus one trailing
// space). Returns the char stripped, or "".
func stripLeftChrome(lines []string) ([]string, string) {
	counts := map[rune]int{}
	nonEmpty := 0
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		nonEmpty++
		trimmed := strings.TrimLeft(ln, " \t")
		r := []rune(trimmed)
		if len(r) > 0 && leftBarRunes[r[0]] {
			counts[r[0]]++
		}
	}
	if nonEmpty == 0 {
		return lines, ""
	}
	var best rune
	var bestN int
	for r, n := range counts {
		if n > bestN {
			best, bestN = r, n
		}
	}
	if bestN == 0 || float64(bestN)/float64(nonEmpty) < 0.6 {
		return lines, ""
	}
	for i, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		idx := strings.IndexRune(ln, best)
		// Only strip if it's within the leading whitespace region.
		if idx >= 0 && strings.TrimSpace(ln[:idx]) == "" {
			rest := ln[idx+len(string(best)):]
			rest = strings.TrimPrefix(rest, " ") // drop one padding space
			lines[i] = rest
		}
	}
	return lines, string(best)
}

// stripRightChrome removes a trailing vertical border char present on most lines.
func stripRightChrome(lines []string) ([]string, bool) {
	counts := 0
	nonEmpty := 0
	for _, ln := range lines {
		t := strings.TrimRight(ln, " \t")
		if t == "" {
			continue
		}
		nonEmpty++
		r := []rune(t)
		if leftBarRunes[r[len(r)-1]] {
			counts++
		}
	}
	if nonEmpty == 0 || float64(counts)/float64(nonEmpty) < 0.6 {
		return lines, false
	}
	for i, ln := range lines {
		t := strings.TrimRight(ln, " \t")
		if t == "" {
			continue
		}
		r := []rune(t)
		if leftBarRunes[r[len(r)-1]] {
			lines[i] = strings.TrimRight(string(r[:len(r)-1]), " \t")
		}
	}
	return lines, true
}

// dedent removes the common leading-space prefix from all non-empty lines.
func dedent(lines []string) ([]string, int) {
	min := -1
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		n := len(ln) - len(strings.TrimLeft(ln, " "))
		if min == -1 || n < min {
			min = n
		}
	}
	if min <= 0 {
		return lines, 0
	}
	for i, ln := range lines {
		if len(ln) >= min {
			lines[i] = ln[min:]
		}
	}
	return lines, min
}

// collapseBlankLines collapses runs of 2+ blank lines into a single blank line.
func collapseBlankLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	blank := false
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			if blank {
				continue
			}
			blank = true
			out = append(out, "")
		} else {
			blank = false
			out = append(out, ln)
		}
	}
	return out
}

// --- reflow ---

var (
	reHeading  = regexp.MustCompile(`^#{1,6}\s+\S`)
	reULItem   = regexp.MustCompile(`^\s*[-*+]\s+\S`)
	reOLItem   = regexp.MustCompile(`^\s*\d+[.)]\s+\S`)
	reBlockq   = regexp.MustCompile(`^\s*>\s?`)
	reHRule    = regexp.MustCompile(`^\s*(?:(?:-\s*){3,}|(?:\*\s*){3,}|(?:_\s*){3,})$`)
	reTableRow = regexp.MustCompile(`^\s*\|.*\|\s*$`)
	reTableSep = regexp.MustCompile(`^\s*\|?\s*:?-{2,}:?\s*(\|\s*:?-{2,}:?\s*)+\|?\s*$`)
)

func isFence(line string) bool {
	t := strings.TrimSpace(line)
	return strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~")
}

// emitsOwnLine: structural lines that must never absorb a continuation.
func emitsOwnLine(line string) bool {
	return reHeading.MatchString(line) || reHRule.MatchString(line) ||
		reTableSep.MatchString(line) || reTableRow.MatchString(line)
}

// startsBlock: structural lines that begin a logical line but may absorb
// wrapped continuation prose (list items, blockquotes).
func startsBlock(line string) bool {
	return reULItem.MatchString(line) || reOLItem.MatchString(line) ||
		reBlockq.MatchString(line)
}

// reflow rejoins terminal-wrapped lines. Code fences are copied verbatim. In
// Markdown mode, structural lines keep their boundaries; in plain mode every
// blank-delimited block is treated as a wrap-rejoinable paragraph.
func reflow(lines []string, format detect.Format) []string {
	width := detectWidth(lines)
	if width < 40 {
		// No evidence of a consistent wrap width; leave structure as-is.
		return lines
	}

	var out []string
	var para []string
	flush := func() {
		if len(para) > 0 {
			out = append(out, joinParagraph(para, width)...)
			para = nil
		}
	}

	inFence := false
	for _, ln := range lines {
		if isFence(ln) {
			flush()
			inFence = !inFence
			out = append(out, ln)
			continue
		}
		if inFence {
			out = append(out, ln)
			continue
		}
		if strings.TrimSpace(ln) == "" {
			flush()
			out = append(out, "")
			continue
		}
		if format == detect.Markdown {
			if emitsOwnLine(ln) {
				flush()
				out = append(out, ln)
				continue
			}
			if startsBlock(ln) {
				flush()
				para = []string{ln}
				continue
			}
		}
		para = append(para, ln)
	}
	flush()
	return out
}

// detectWidth estimates the terminal wrap width as the longest non-empty,
// non-fence line. Lines inside code fences are excluded.
func detectWidth(lines []string) int {
	max := 0
	inFence := false
	for _, ln := range lines {
		if isFence(ln) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		n := len([]rune(ln))
		if n > max {
			max = n
		}
	}
	return max
}

// joinParagraph merges every line of a wrapped paragraph into one logical line.
// reflow only runs when the detected width is >= 40 columns (see reflow), i.e.
// the text is genuinely terminal-wrapped prose, never a short poem or list — so
// every line break inside a blank-delimited block is a forced wrap and can be
// rejoined. joinSep decides whether each seam takes a space (word wrap) or none
// (a URL/path character-wrapped mid-token).
func joinParagraph(para []string, width int) []string {
	cur := strings.TrimRight(para[0], " ")
	for i := 1; i < len(para); i++ {
		next := strings.TrimLeft(para[i], " \t")
		cur += joinSep(cur, next, width) + next
	}
	return []string{cur}
}

// tokenConnectors are characters found inside URLs, file paths, and code
// identifiers — places a long token can be character-wrapped by the terminal.
const tokenConnectors = `/:._-=?&#%@~\+`

// joinSep decides whether two wrapped fragments are glued with a space or not.
// A normal word wrap broke at a space, so we re-insert one. But a URL, path, or
// identifier longer than the line gets character-wrapped mid-token — there was
// never a space at the break, so joining must not add one (e.g. a wrapped
// "http://host/" + "path" must rejoin as "http://host/path", not ".../ path").
func joinSep(prev, next string, width int) string {
	combined := lastToken(prev) + firstWord(next)
	// A single token wider than the line must have been character-wrapped; if it
	// also carries a URL/path/identifier connector (and isn't a sentence break),
	// the fragments belong together with no space.
	if len([]rune(combined)) > width &&
		strings.ContainsAny(combined, tokenConnectors) &&
		!sentenceBoundary(prev, next) {
		return ""
	}
	return " "
}

// lastToken returns the trailing run of non-space characters in s.
func lastToken(s string) string {
	f := strings.Fields(s)
	if len(f) == 0 {
		return ""
	}
	return f[len(f)-1]
}

// sentenceBoundary reports whether prev ends a sentence (.!?) and next starts
// with a capital — a real prose break that should keep its space even if the
// fragments happen to look token-like.
func sentenceBoundary(prev, next string) bool {
	p := []rune(strings.TrimRight(prev, " \t"))
	if len(p) == 0 {
		return false
	}
	switch p[len(p)-1] {
	case '.', '!', '?':
	default:
		return false
	}
	n := []rune(strings.TrimLeft(next, " \t"))
	return len(n) > 0 && unicode.IsUpper(n[0])
}

func firstWord(s string) string {
	s = strings.TrimLeft(s, " \t")
	i := strings.IndexFunc(s, func(r rune) bool { return unicode.IsSpace(r) })
	if i < 0 {
		return s
	}
	return s[:i]
}
