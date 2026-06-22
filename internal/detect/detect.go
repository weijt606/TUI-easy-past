// Package detect classifies a block of text as Markdown or plain text so the
// cleaner can pick a structure-preserving (markdown) or reflow-aggressive
// (plaintext) strategy. Detection runs AFTER chrome/border stripping so that
// terminal padding doesn't fool the heuristics.
package detect

import (
	"regexp"
	"strings"
)

// Format is the classification result.
type Format int

const (
	PlainText Format = iota
	Markdown
)

func (f Format) String() string {
	if f == Markdown {
		return "markdown"
	}
	return "plaintext"
}

// Result carries the decision plus the raw score for --explain output.
type Result struct {
	Format  Format
	Score   float64 // weighted markdown evidence, normalized by line count
	Signals []string
}

var (
	reHeading    = regexp.MustCompile(`^#{1,6}\s+\S`)
	reULItem     = regexp.MustCompile(`^\s*[-*+]\s+\S`)
	reOLItem     = regexp.MustCompile(`^\s*\d+[.)]\s+\S`)
	reBlockq     = regexp.MustCompile(`^\s*>\s`)
	reHRule      = regexp.MustCompile(`^\s*(?:(?:-\s*){3,}|(?:\*\s*){3,}|(?:_\s*){3,})$`)
	reTableSep   = regexp.MustCompile(`^\s*\|?\s*:?-{2,}:?\s*(\|\s*:?-{2,}:?\s*)+\|?\s*$`)
	reLink       = regexp.MustCompile(`\[[^\]]+\]\([^)]+\)`)
	reInlineCode = regexp.MustCompile("`[^`]+`")
	reBold       = regexp.MustCompile(`(\*\*|__)(?:\S.*?\S|\S)(\*\*|__)`)
	reItalic     = regexp.MustCompile(`(?:^|\s)(?:\*[^*\s][^*]*\*|_[^_\s][^_]*_)(?:\s|$)`)
)

// Detect classifies text. It returns Markdown when strong block-level signals
// (fences, headings, tables, lists, blockquotes) are present, or when enough
// inline signals accumulate across the text.
func Detect(text string) Result {
	lines := strings.Split(text, "\n")
	var signals []string
	addSignal := func(s string) {
		for _, e := range signals {
			if e == s {
				return
			}
		}
		signals = append(signals, s)
	}

	var score float64
	inFence := false
	fenceCount := 0
	nonEmpty := 0

	for _, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		if trimmed != "" {
			nonEmpty++
		}

		// Fenced code blocks are the strongest single signal.
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fenceCount++
			inFence = !inFence
			score += 3
			addSignal("code-fence")
			continue
		}
		if inFence {
			continue // don't score content inside fences
		}

		switch {
		case reHeading.MatchString(ln):
			score += 3
			addSignal("heading")
		case reTableSep.MatchString(ln):
			score += 3
			addSignal("table")
		case reBlockq.MatchString(ln):
			score += 1.5
			addSignal("blockquote")
		case reHRule.MatchString(ln):
			score += 1
			addSignal("hrule")
		case reOLItem.MatchString(ln):
			score += 1
			addSignal("ordered-list")
		case reULItem.MatchString(ln):
			score += 1
			addSignal("unordered-list")
		}

		// Inline signals (lower weight, can fire alongside block ones).
		if reLink.MatchString(ln) {
			score += 1.5
			addSignal("link")
		}
		if reInlineCode.MatchString(ln) {
			score += 0.75
			addSignal("inline-code")
		}
		if reBold.MatchString(ln) {
			score += 0.75
			addSignal("bold")
		} else if reItalic.MatchString(ln) {
			score += 0.5
			addSignal("italic")
		}
	}

	// An unterminated single fence is suspicious; only trust paired fences.
	if fenceCount == 1 {
		score -= 1.5
	}

	res := Result{Score: score, Signals: signals}

	// Decision: any strong structural signal, or accumulated score that is
	// meaningful relative to the amount of text.
	strong := contains(signals, "code-fence") || contains(signals, "heading") ||
		contains(signals, "table")
	density := score
	if nonEmpty > 0 {
		density = score / float64(nonEmpty)
	}

	if (strong && score >= 1.5) || score >= 4 || density >= 0.6 {
		res.Format = Markdown
	} else {
		res.Format = PlainText
	}
	return res
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
