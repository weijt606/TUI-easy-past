# tep — TUI Easy Paste

**English** | [简体中文](README.zh-CN.md)

Copy text out of a terminal/TUI (Claude Code, Codex CLI, etc.) and it usually
arrives mangled somewhere else: every line indented by the UI's padding, prose
hard-wrapped mid-sentence, box-drawing borders and ANSI colors leaking in. `tep`
cleans that up so you can paste straight into Reddit, X, a doc, or a chat.

It **auto-detects Markdown vs. plain text** and cleans each appropriately:

- **Plain text** → terminal-wrapped lines are rejoined into clean paragraphs.
- **Markdown** → structure is preserved: headings, list items, blockquotes,
  tables, and fenced code blocks keep their boundaries; only wrapped prose
  *within* a block is rejoined. Code fences are copied verbatim.

## Why another one?

There are existing cleaners (e.g. `ai-clean`). `tep`'s focus is **correctness of
the plain-vs-Markdown distinction** and a **width-aware reflow** that only rejoins
a line when the next word genuinely wouldn't have fit at the detected wrap width —
so authored line breaks survive and only the terminal's forced wraps are undone.
It also never strips `>` (a Markdown blockquote marker) as if it were border
chrome.

## Install

```sh
go install github.com/weijt606/TUI-easy-past@latest   # installs the `tep` binary
```

Or build from source:

```sh
git clone https://github.com/weijt606/TUI-easy-past
cd TUI-easy-past
go build -o tep .
```

No cgo, no third-party dependencies. Clipboard access shells out to the native
utility: `pbcopy`/`pbpaste` (macOS), `wl-copy`/`xclip`/`xsel` (Linux),
`clip`/`Get-Clipboard` (Windows).

## Usage

```sh
tep                 # read clipboard, clean it, write it back (the common case)
tep --dry-run       # print the cleaned result, leave the clipboard untouched
tep --stdin         # read stdin, write cleaned text to stdout
cat session.log | tep -      # same as --stdin
tep --explain       # also report what was done (to stderr)
```

Typical flow: select text in your TUI, copy it, run `tep`, paste.

### Flags

| Flag | Effect |
|---|---|
| `-n`, `--dry-run` | Print result to stdout; don't modify the clipboard. |
| `--stdin`, `-` | Read from stdin, write to stdout. |
| `--explain` | Print detected format, stripped chrome, dedent, etc. to stderr. |
| `--no-rejoin` | Strip chrome/whitespace but keep line breaks as-is. |
| `--keep-ansi` | Leave ANSI escape sequences in place. |
| `--markdown` | Force Markdown mode (skip auto-detection). |
| `--plain` | Force plain-text mode (skip auto-detection). |

## What it does, in order

1. Normalize line endings and strip ANSI escapes.
2. Drop horizontal box borders (`┌──┐`, `└──┘`); keep empty interior rows as
   blank lines so paragraph breaks aren't lost.
3. Strip a consistent left/right vertical border (`│`, `|`, …) used as chrome.
4. Trim trailing whitespace and dedent the common left padding.
5. Detect Markdown vs. plain text.
6. Reflow wrapped lines (width-aware), respecting Markdown structure.
7. Collapse runs of blank lines.

## Limitations

Reflow is a heuristic — the terminal throws away whether a line break was
authored or forced when it wraps, so `tep` infers it from the detected wrap
width. Consequences:

- Content narrower than ~40 columns is left unreflowed, to avoid wrongly merging
  intentionally short lines (poems, narrow lists). Real TUI output (wrapped at
  80+ columns) reflows fine.
- If you want a guaranteed-lossless copy of Claude Code output specifically, the
  built-in `/copy` command copies the raw Markdown directly. `tep` is the general
  fallback for any TUI, working from whatever you selected.

Use `--no-rejoin` if you'd rather keep every line break and only strip the chrome.

## License

MIT
