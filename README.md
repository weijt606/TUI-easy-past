# tep — TUI Easy Paste

[![lang: English](https://img.shields.io/badge/lang-English-blue.svg)](README.md)
[![lang: 简体中文](https://img.shields.io/badge/lang-%E7%AE%80%E4%BD%93%E4%B8%AD%E6%96%87-lightgrey.svg)](README.zh-CN.md)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8.svg?logo=go&logoColor=white)](go.mod)
[![Works with](https://img.shields.io/badge/works%20with-Claude%20Code%20%C2%B7%20Codex%20CLI-7c5cff.svg)](#)

Copy text out of a terminal/TUI (Claude Code, Codex CLI, etc.) and it usually
arrives mangled somewhere else: every line indented by the UI's padding, prose
hard-wrapped mid-sentence, box-drawing borders and ANSI colors leaking in. `tep`
cleans that up so you can paste straight into Reddit, X, a doc, or a chat.

It **auto-detects Markdown vs. plain text** and cleans each appropriately:

- **Plain text** → terminal-wrapped lines are rejoined into clean paragraphs.
- **Markdown** → structure is preserved: headings, list items, blockquotes,
  tables, and fenced code blocks keep their boundaries; only wrapped prose
  *within* a block is rejoined. Code fences are copied verbatim.

## Quickstart

```sh
# 1. Install
go install github.com/weijt606/TUI-easy-past/cmd/tep@latest

# 2. Copy some text out of your TUI (Claude Code, Codex, …) as usual.

# 3. Clean the clipboard in place:
tep

# 4. Paste anywhere — formatting is fixed.
```

No install? Pipe through it in one shot (macOS):

```sh
pbpaste | tep - | pbcopy
```

## Everyday workflow

Three steps:

```text
①  Copy text in your TUI      (mouse-select, then ⌘C / Ctrl-Shift-C)
        │
        ▼
②  Run  tep                   reads the clipboard, cleans it, writes it back
        │
        ▼
③  Paste anywhere             line breaks, indentation & borders fixed
                              (Reddit · X · docs · chat)
```

`tep` works on the system clipboard, so it runs from any shell — a second
terminal tab is always an option. But usually you don't even need one:

### Run it without leaving Claude Code / Codex CLI

Both **Claude Code** and **Codex CLI** let you run a shell command inline by
starting the line with `!`. So right after you copy, just type:

```
!tep
```

That cleans the clipboard in place — no new terminal, no leaving the session.
Then switch to where you want the text and paste.

- **Claude Code** — `!tep` runs in the session shell and drops you back at the
  prompt.
- **Codex CLI** — `!tep` runs subject to your approval/sandbox settings, and the
  command's output is fed back into the conversation.

## Install

```sh
go install github.com/weijt606/TUI-easy-past/cmd/tep@latest   # installs the `tep` binary
```

Or build from source:

```sh
git clone https://github.com/weijt606/TUI-easy-past
cd TUI-easy-past
go build -o tep ./cmd/tep
```

No cgo, no third-party dependencies. Clipboard access shells out to the native
utility: `pbcopy`/`pbpaste` (macOS), `wl-copy`/`xclip`/`xsel` (Linux),
`clip`/`Get-Clipboard` (Windows).

## Usage

The everyday flow is three keystrokes' worth of work: **copy in your TUI → run
`tep` → paste.** `tep` with no arguments reads the clipboard, cleans it, and
writes it straight back.

```sh
tep                       # clean the clipboard in place (the common case)
tep --dry-run             # print the cleaned result; leave the clipboard alone
tep --explain             # also report what it detected & changed (to stderr)
tep --markdown            # force Markdown mode if auto-detect guesses wrong
pbpaste | tep - | pbcopy  # explicit pipe instead of the in-place default
cat session.log | tep -   # clean a captured log, print to stdout
```

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
