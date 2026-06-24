package clean

import (
	"strings"
	"testing"

	"github.com/weijt606/TUI-easy-past/internal/detect"
)

func TestPlainReflow(t *testing.T) {
	// Terminal-wrapped prose with 2-space left padding.
	in := "  This is a long paragraph that the terminal\n" +
		"  wrapped across several lines because it did\n" +
		"  not fit the width.\n" +
		"\n" +
		"  Second paragraph here."
	want := "This is a long paragraph that the terminal wrapped across several lines because it did not fit the width.\n" +
		"\n" +
		"Second paragraph here."
	got, rep := Clean(in, Options{})
	if got != want {
		t.Errorf("plain reflow:\n got: %q\nwant: %q", got, want)
	}
	if rep.Format != detect.PlainText {
		t.Errorf("format = %v, want PlainText", rep.Format)
	}
	if rep.Dedented != 2 {
		t.Errorf("dedent = %d, want 2", rep.Dedented)
	}
}

func TestMarkdownPreserved(t *testing.T) {
	in := "  # Title\n" +
		"\n" +
		"  Some intro text that wraps onto\n" +
		"  a second line right here now.\n" +
		"\n" +
		"  - first item\n" +
		"  - second item that is quite long and\n" +
		"    wraps onto another line of text\n" +
		"\n" +
		"  ```go\n" +
		"  func main() {\n" +
		"      fmt.Println(\"hi\")\n" +
		"  }\n" +
		"  ```\n"
	got, rep := Clean(in, Options{})
	if rep.Format != detect.Markdown {
		t.Fatalf("format = %v, want Markdown", rep.Format)
	}
	// Heading preserved.
	if !strings.Contains(got, "# Title") {
		t.Errorf("heading lost:\n%s", got)
	}
	// Both list items preserved as separate logical lines.
	if strings.Count(got, "- ") < 2 {
		t.Errorf("list items lost:\n%s", got)
	}
	// Fence content preserved verbatim, including relative indentation.
	if !strings.Contains(got, "```go\nfunc main() {\n    fmt.Println(\"hi\")\n}\n```") {
		t.Errorf("code fence not preserved verbatim:\n%s", got)
	}
}

func TestStripANSI(t *testing.T) {
	in := "\x1b[1;32mhello\x1b[0m world"
	got, _ := Clean(in, Options{})
	if got != "hello world" {
		t.Errorf("ansi strip: got %q want %q", got, "hello world")
	}
}

func TestBoxBorders(t *testing.T) {
	in := "┌──────────────┐\n" +
		"│ inside text  │\n" +
		"│ more text    │\n" +
		"└──────────────┘"
	got, rep := Clean(in, Options{})
	if strings.ContainsAny(got, "┌┐└┘│─") {
		t.Errorf("box chars not stripped: %q", got)
	}
	if rep.LeftChrome != "│" {
		t.Errorf("left chrome = %q, want │", rep.LeftChrome)
	}
	// Two short interior lines, different content → stay separate.
	if !strings.Contains(got, "inside text") || !strings.Contains(got, "more text") {
		t.Errorf("interior text lost: %q", got)
	}
}

// Empty interior rows of a box must survive as blank lines so that two distinct
// paragraphs inside the box don't get merged into one.
func TestBoxBlankLinesNotMerged(t *testing.T) {
	in := "┌────────────────────────────────┐\n" +
		"│  First paragraph inside a box  │\n" +
		"│  that wraps a little here.     │\n" +
		"│                                │\n" +
		"│  Second separate paragraph.    │\n" +
		"└────────────────────────────────┘"
	got, _ := Clean(in, Options{})
	if !strings.Contains(got, "\n\n") {
		t.Errorf("blank line between paragraphs lost (paragraphs merged):\n%q", got)
	}
	if strings.Contains(got, "here. Second") {
		t.Errorf("two paragraphs were wrongly merged:\n%q", got)
	}
}

// A long URL character-wrapped by the terminal must rejoin without a space
// inserted at the break point.
func TestWrappedURLRejoinsWithoutSpace(t *testing.T) {
	in := "Check out the polling guide at the link below:\n" +
		"http://localhost:3000/\n" +
		"blog/how-to-choose-a-live-polling-tool"
	pt := detect.PlainText
	got, _ := Clean(in, Options{Format: &pt})
	if strings.Contains(got, "3000/ blog") {
		t.Errorf("space wrongly inserted inside URL:\n%q", got)
	}
	if !strings.Contains(got, "http://localhost:3000/blog/how-to-choose-a-live-polling-tool") {
		t.Errorf("URL not rejoined cleanly:\n%q", got)
	}
}

// A complete URL at the end of a word-wrapped line (shorter than the wrap
// width) is followed by a real new word — the space must be kept.
func TestCompleteURLKeepsTrailingSpace(t *testing.T) {
	in := "For the details please see the full writeup at https://example.com/news\n" +
		"and reply with your thoughts whenever you get a free moment today."
	pt := detect.PlainText
	got, _ := Clean(in, Options{Format: &pt})
	if strings.Contains(got, "newsand") {
		t.Errorf("space wrongly removed after a complete URL:\n%q", got)
	}
	if !strings.Contains(got, "https://example.com/news and reply") {
		t.Errorf("expected a space after the complete URL:\n%q", got)
	}
}

// A wrapped paragraph must fully rejoin even when a line ends a few columns
// short of the detected width (its next word landed exactly at the width). The
// old width/forcedWrap check left such seams broken.
func TestWrappedParagraphFullyRejoins(t *testing.T) {
	in := "Memorize the structure, not the sentences.\n" +
		"Break the speech into 5 to 7 beats and learns ok\n" +
		"now and again."
	pt := detect.PlainText
	got, _ := Clean(in, Options{Format: &pt})
	if strings.Contains(got, "\n") {
		t.Errorf("paragraph not fully rejoined into one line:\n%q", got)
	}
	if !strings.Contains(got, "sentences. Break") {
		t.Errorf("seam lost its space:\n%q", got)
	}
}

func TestNoRejoin(t *testing.T) {
	in := "  line one here\n  line two here"
	got, _ := Clean(in, Options{NoRejoin: true})
	if got != "line one here\nline two here" {
		t.Errorf("no-rejoin: got %q", got)
	}
}

func TestForceFormat(t *testing.T) {
	in := "- item one that is long and gets wrapped\n  by the terminal here onto two lines"
	pt := detect.PlainText
	got, rep := Clean(in, Options{Format: &pt})
	if rep.Format != detect.PlainText {
		t.Fatalf("format = %v, want forced PlainText", rep.Format)
	}
	// In plain mode the dash line is just prose; it should still rejoin.
	if strings.Count(got, "\n") != 0 {
		t.Errorf("forced plain should rejoin to one line: %q", got)
	}
}
