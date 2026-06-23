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

```mermaid
flowchart LR
    A(["Install once"]) --> B["Copy text<br/>in your TUI"]
    B --> C{"Run tep"}
    C -->|"new terminal tab"| T["tep"]
    C -->|"Claude Code: !tep"| T
    C -->|"global hotkey"| T
    T --> D(["Paste at destination<br/>Reddit · X · docs"])
```

1. **Copy** text out of your TUI the usual way (mouse-select, then ⌘C /
   Ctrl-Shift-C). The selection lands on the system clipboard.
2. **Run `tep`.** It reads the clipboard, cleans it, and writes it back.
3. **Paste** at the destination — the formatting is fixed.

### Running `tep` while a TUI is open

`tep` works on the *system clipboard*, so it doesn't have to run inside the TUI
you copied from. Use whichever is handy:

- **A second terminal tab / window / split** — the universal option, works
  everywhere.
- **Claude Code** — type `!tep` at the prompt. The `!` prefix runs a shell
  command in the session, so it cleans the clipboard without leaving Claude Code.
- **Codex CLI and other TUIs** — if the tool has a shell-escape, run `tep` there;
  otherwise use a second terminal or a global hotkey.
- **A global hotkey** (recommended) — copy, press the key, paste. No shell prompt
  needed at all. See [Bind a hotkey](#optional-bind-a-hotkey).

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

## Optional: bind a hotkey

`tep` is a plain CLI on purpose — it does **not** run a background daemon or grab
global keys itself (that keeps it simple, scriptable, and free of a always-on
process). To get a one-press "clean my clipboard" key, point your OS hotkey tool
at `tep`:

**macOS**

- **Hammerspoon** (free, scriptable) — add to `~/.hammerspoon/init.lua`:
  ```lua
  hs.hotkey.bind({ "cmd", "alt" }, "V", function()
    hs.execute("/opt/homebrew/bin/tep")   -- clean the clipboard in place
  end)
  ```
  Now ⌘⌥V cleans the clipboard; paste normally with ⌘V.
- **Raycast / Alfred** — add a Script Command / Workflow that runs `tep` and
  assign it a hotkey.
- **Shortcuts.app** — new shortcut → *Run Shell Script* `tep` → assign a key in
  System Settings → Keyboard → Keyboard Shortcuts.

**Linux** — bind a custom shortcut to `tep` in GNOME/KDE settings, or via `sxhkd`.

**Windows** — AutoHotkey: `^!v::Run, tep, , Hide`.

> Tip: if you usually paste to the same kind of destination, bind a second key to
> a forced mode — e.g. `tep --plain` for X/Twitter, `tep --markdown` for Reddit.

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
