// Command tep (TUI Easy Paste) cleans text copied out of a terminal/TUI so it
// can be pasted elsewhere without mangled line breaks, indentation, or borders.
//
// Default: read the clipboard, clean it, write it back.
//
//	tep                 # clean clipboard in place
//	tep --dry-run       # print the cleaned result, don't touch the clipboard
//	tep --stdin         # read stdin, write cleaned text to stdout
//	cat log | tep -     # same as --stdin
//	tep --explain       # report what was done (to stderr)
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/weijt606/TUI-easy-past/internal/clean"
	"github.com/weijt606/TUI-easy-past/internal/clipboard"
	"github.com/weijt606/TUI-easy-past/internal/detect"
)

const version = "0.1.0"

const usage = `tep ` + version + ` — TUI Easy Paste

Clean text copied out of a terminal/TUI (Claude Code, Codex, etc.) so it pastes
cleanly: strips ANSI, box borders, terminal padding, and rejoins wrapped lines.
Auto-detects Markdown vs plain text and cleans each appropriately.

USAGE:
    tep [flags]              read clipboard, clean, write back
    tep --dry-run            print cleaned result, leave clipboard untouched
    tep --stdin | tep -      read stdin, write stdout

FLAGS:
    -n, --dry-run     print result to stdout; do not modify the clipboard
        --stdin, -    read from stdin and write to stdout instead of clipboard
        --explain     print what was done to stderr
        --no-rejoin   do not rejoin wrapped lines (only strip chrome/whitespace)
        --keep-ansi   keep ANSI escape sequences
        --markdown    force Markdown mode (skip auto-detection)
        --plain       force plain-text mode (skip auto-detection)
    -h, --help        show this help
    -v, --version     show version
`

type config struct {
	stdin    bool
	dryRun   bool
	explain  bool
	noRejoin bool
	keepANSI bool
	force    *detect.Format
}

func parseArgs(args []string) (config, error) {
	var c config
	for _, a := range args {
		switch a {
		case "-h", "--help":
			fmt.Print(usage)
			os.Exit(0)
		case "-v", "--version":
			fmt.Println("tep", version)
			os.Exit(0)
		case "-", "--stdin":
			c.stdin = true
		case "-n", "--dry-run":
			c.dryRun = true
		case "--explain":
			c.explain = true
		case "--no-rejoin":
			c.noRejoin = true
		case "--keep-ansi":
			c.keepANSI = true
		case "--markdown":
			f := detect.Markdown
			c.force = &f
		case "--plain":
			f := detect.PlainText
			c.force = &f
		default:
			return c, fmt.Errorf("unknown flag: %s (try --help)", a)
		}
	}
	return c, nil
}

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "tep:", err)
		os.Exit(2)
	}

	var input string
	if cfg.stdin {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "tep: reading stdin:", err)
			os.Exit(1)
		}
		input = string(b)
	} else {
		input, err = clipboard.Read()
		if err != nil {
			fmt.Fprintln(os.Stderr, "tep: reading clipboard:", err)
			os.Exit(1)
		}
	}

	out, rep := clean.Clean(input, clean.Options{
		Format:   cfg.force,
		NoRejoin: cfg.noRejoin,
		KeepANSI: cfg.keepANSI,
	})

	if cfg.explain {
		printReport(os.Stderr, rep)
	}

	switch {
	case cfg.stdin:
		fmt.Print(out)
		if len(out) > 0 && out[len(out)-1] != '\n' {
			fmt.Println()
		}
	case cfg.dryRun:
		fmt.Println(out)
	default:
		if err := clipboard.Write(out); err != nil {
			fmt.Fprintln(os.Stderr, "tep: writing clipboard:", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "tep: cleaned clipboard (%s)\n", rep.Format)
	}
}

func printReport(w io.Writer, rep clean.Report) {
	fmt.Fprintln(w, "── tep --explain ──")
	fmt.Fprintf(w, "format:      %s\n", rep.Format)
	if len(rep.DetectSignals) > 0 {
		fmt.Fprintf(w, "detect:      score=%.1f signals=%v\n", rep.DetectScore, rep.DetectSignals)
	}
	if rep.LeftChrome != "" {
		fmt.Fprintf(w, "left chrome: stripped %q\n", rep.LeftChrome)
	}
	if rep.RightChrome {
		fmt.Fprintln(w, "right chrome: stripped")
	}
	if rep.Dedented > 0 {
		fmt.Fprintf(w, "dedent:      removed %d columns\n", rep.Dedented)
	}
	fmt.Fprintf(w, "rejoin:      %v\n", rep.Rejoined)
	fmt.Fprintln(w, "───────────────────")
}
