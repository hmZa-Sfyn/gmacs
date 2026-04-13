# zed — a fast Emacs-style terminal editor in Go

A lightweight, buffer-based terminal code editor built with [tcell](https://github.com/gdamore/tcell).

## Build & Run

```bash
# In the zed/ directory:
go mod tidy
go build .
./zed file.go             # open a file
./zed                     # open scratch buffer
./zed a.go b.go c.go     # open multiple files as tabs
```

## Features

- **Tabbed buffers** — open as many files as you want
- **Syntax highlighting** — Go, Python, C/C++, Shell, JSON, Markdown
- **Themes** — dark (default), light, monokai, gruvbox
- **File explorer** panel
- **Embedded terminal** split pane (runs your $SHELL) alongside files and explorer
- **Find** / **replace** (whole-buffer)
- **Undo / Redo**
- **Mouse support** — click to place cursor, scroll wheel, click tabs
- **Auto-indent** on Enter
- **Word-jump** (Ctrl+Arrow)
- **Line move** (Alt+Up/Down)

## Keybindings

| Key | Action |
|-----|--------|
| Arrows | Move cursor |
| Shift+Arrows | Extend selection |
| Ctrl+Left/Right | Word jump |
| Shift+Ctrl+Left/Right | Word-select |
| Alt+Left/Right | Move cursor ×10 |
| Alt+Up/Down | Move line up/down |
| Home / End | Line start/end |
| Ctrl+Home / Ctrl+End | File start/end |
| PgUp / PgDn | Page scroll |
| Ctrl+A | Select all |
| Ctrl+C | Copy selection |
| Ctrl+X | Cut selection |
| Ctrl+V | Paste |
| Ctrl+D | Delete selection |
| Ctrl+Z | Undo |
| Ctrl+Y | Redo |
| Ctrl+S | Save |
| Ctrl+F | Find |
| Ctrl+H | Replace (command: `replace <find> <replace>`) |
| Ctrl+T | Toggle terminal split pane |
| Ctrl+E | Toggle file explorer |
| Alt+E | Open explorer in current file's dir |
| Ctrl+O | Open explorer |
| Ctrl+N / Ctrl+P | Next / Prev tab |
| Ctrl+W | Close tab |
| Alt+1–9, 0 | Jump to tab 1–10 |
| Alt+: | Command palette |
| Escape | Cancel / clear selection |

## Commands (Ctrl+: or Ctrl+H)

| Command | Effect |
|---------|--------|
| `theme dark\|light\|monokai\|gruvbox` | Switch theme |
| `replace <find> <replace>` | Replace all in buffer |
| `saveas <path>` | Save as new path |
| `open <path>` | Open file |
| `new [name]` | New buffer |
| `cd <dir>` | Change working directory |
| `w` / `write` | Save |
| `q` / `quit` | Quit |
| `wq` | Save and quit |

## Adding a Syntax Highlighter (Plugin API)

Implement the `Highlighter` interface in any `.go` file in the same package, then register it in `highlight.go`'s `init()`:

```go
type RustHighlighter struct{}

func (r *RustHighlighter) FileMatch(name string) bool {
    return strings.HasSuffix(name, ".rs")
}

func (r *RustHighlighter) Tokenize(lines [][]rune, lineIdx int, state *HLState) []Token {
    // return []Token with Col, Len, Kind for each span
}
```

Then in `highlight.go` `init()`:
```go
Highlighters = append(Highlighters, &RustHighlighter{})
```

## Adding a Theme

Add a new entry to the `Themes` map in `theme.go`:

```go
Themes["solarized"] = Theme{
    Name: "solarized",
    BG:   tcell.NewRGBColor(0, 43, 54),
    FG:   tcell.NewRGBColor(131, 148, 150),
    // ... fill in the rest
}
```

Then switch with `:theme solarized`.

## Architecture

```
main.go        — entry point, screen init
editor.go      — Editor struct, tabs, drawing, commands
buffer.go      — Buffer (lines, cursor, undo/redo, find/replace)
keybindings.go — All key + mouse handling
highlight.go   — Highlighter interface + built-in languages
theme.go       — Theme system
explorer.go    — File explorer panel
terminal.go    — Embedded shell terminal tab
clipboard.go   — System clipboard integration
```
