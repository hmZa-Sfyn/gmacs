package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// TabKind distinguishes buffer tabs from terminal tabs
type TabKind int

const (
	TabBuffer TabKind = iota
	TabTerminal
)

type Tab struct {
	Kind    TabKind
	Buffer  *Buffer   // nil for terminal
	Term    *TermTab  // nil for buffer
}

func (tab *Tab) Name() string {
	switch tab.Kind {
	case TabBuffer:
		name := tab.Buffer.Name
		if tab.Buffer.Dirty {
			name += " •"
		}
		return name
	case TabTerminal:
		return tab.Term.Name
	}
	return "?"
}

// EditorMode for status line
type EditorMode int

const (
	ModeNormal EditorMode = iota
	ModeFind
	ModeCommand
)

type Editor struct {
	Screen   tcell.Screen
	Tabs     []*Tab
	Active   int

	Explorer     *Explorer
	ShowExplorer bool

	Mode       EditorMode
	ModeInput  string   // text typed in find/command bar
	FindStr    string
	MsgLine    string   // status message

	// Selection drag start for shift-click
	hlStates   map[int]*HLState // per-buffer highlight state cache
	lineNumWidth int
}

func NewEditor(screen tcell.Screen) *Editor {
	e := &Editor{
		Screen:   screen,
		hlStates: make(map[int]*HLState),
	}
	cwd, _ := os.Getwd()
	e.Explorer = NewExplorer(cwd)
	return e
}

func (e *Editor) NewBuffer(name string) *Tab {
	b := NewBuffer(name)
	tab := &Tab{Kind: TabBuffer, Buffer: b}
	e.Tabs = append(e.Tabs, tab)
	e.Active = len(e.Tabs) - 1
	return tab
}

func (e *Editor) OpenFile(path string) {
	// Check if already open
	for i, tab := range e.Tabs {
		if tab.Kind == TabBuffer && tab.Buffer.FilePath == path {
			e.Active = i
			return
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	b, err := NewBufferFromFile(abs)
	if err != nil {
		e.MsgLine = "Error: " + err.Error()
		return
	}
	tab := &Tab{Kind: TabBuffer, Buffer: b}
	e.Tabs = append(e.Tabs, tab)
	e.Active = len(e.Tabs) - 1
}

func (e *Editor) CloseTab(idx int) {
	if len(e.Tabs) == 0 {
		return
	}
	e.Tabs = append(e.Tabs[:idx], e.Tabs[idx+1:]...)
	if e.Active >= len(e.Tabs) {
		e.Active = len(e.Tabs) - 1
	}
	if e.Active < 0 {
		e.Active = 0
	}
	if len(e.Tabs) == 0 {
		e.NewBuffer("*scratch*")
	}
}

func (e *Editor) ToggleTerminal() {
	for i, tab := range e.Tabs {
		if tab.Kind == TabTerminal {
			e.Active = i
			return
		}
	}
	term := NewTermTab()
	tab := &Tab{Kind: TabTerminal, Term: term}
	e.Tabs = append(e.Tabs, tab)
	e.Active = len(e.Tabs) - 1
}

func (e *Editor) ActiveTab() *Tab {
	if len(e.Tabs) == 0 {
		return nil
	}
	return e.Tabs[e.Active]
}

func (e *Editor) ActiveBuffer() *Buffer {
	tab := e.ActiveTab()
	if tab == nil || tab.Kind != TabBuffer {
		return nil
	}
	return tab.Buffer
}

// ---- Main loop ----

func (e *Editor) Run() {
	for {
		e.Draw()
		ev := e.Screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			e.Screen.Sync()
		case *tcell.EventKey:
			if !e.HandleKey(ev) {
				return
			}
		case *tcell.EventMouse:
			e.HandleMouse(ev)
		}
	}
}

// ---- Drawing ----

func (e *Editor) Draw() {
	screen := e.Screen
	w, h := screen.Size()
	screen.Clear()
	t := CurrentTheme
	bgSt := tcell.StyleDefault.Background(t.BG).Foreground(t.FG)

	// Fill background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			screen.SetContent(x, y, ' ', nil, bgSt)
		}
	}

	tabBarH := 1
	statusH := 1
	modeH := 0
	if e.Mode != ModeNormal {
		modeH = 1
	}

	contentY := tabBarH
	contentH := h - tabBarH - statusH - modeH
	contentX := 0
	contentW := w

	if e.ShowExplorer {
		ew := e.Explorer.Width
		e.Explorer.Draw(screen, 0, contentY, contentH)
		contentX = ew
		contentW = w - ew
	}

	// Tab bar
	e.drawTabBar(0, 0, w)

	// Content
	tab := e.ActiveTab()
	if tab != nil {
		if tab.Kind == TabBuffer {
			e.drawBuffer(tab.Buffer, contentX, contentY, contentW, contentH)
		} else if tab.Kind == TabTerminal {
			tab.Term.Draw(screen, contentX, contentY, contentW, contentH)
		}
	}

	// Mode bar (find / command)
	if e.Mode != ModeNormal {
		e.drawModeLine(0, h-statusH-1, w)
	}

	// Status bar
	e.drawStatusBar(0, h-statusH, w)

	screen.Show()
}

func (e *Editor) drawTabBar(x, y, w int) {
	t := CurrentTheme
	screen := e.Screen
	cx := x
	for i, tab := range e.Tabs {
		name := " " + tab.Name() + " "
		st := tcell.StyleDefault.Background(t.TabBG).Foreground(t.TabFG)
		if i == e.Active {
			st = tcell.StyleDefault.Background(t.TabActiveBG).Foreground(t.TabActiveFG).Bold(true)
		}
		for _, ch := range name {
			if cx >= w {
				break
			}
			screen.SetContent(cx, y, ch, nil, st)
			cx++
		}
		// separator
		if cx < w {
			screen.SetContent(cx, y, '│', nil, tcell.StyleDefault.Background(t.TabBG).Foreground(t.TabFG))
			cx++
		}
	}
	// Fill rest
	fillSt := tcell.StyleDefault.Background(t.TabBG).Foreground(t.TabFG)
	for ; cx < w; cx++ {
		screen.SetContent(cx, y, ' ', nil, fillSt)
	}
}

func (e *Editor) drawStatusBar(x, y, w int) {
	t := CurrentTheme
	st := tcell.StyleDefault.Background(t.StatusBG).Foreground(t.StatusFG)

	buf := e.ActiveBuffer()
	var left, right string
	if buf != nil {
		mode := "NRM"
		if e.Mode == ModeFind {
			mode = "FND"
		} else if e.Mode == ModeCommand {
			mode = "CMD"
		}
		left = fmt.Sprintf(" %s  %s", mode, buf.FilePath)
		if buf.FilePath == "" {
			left = fmt.Sprintf(" %s  %s", mode, buf.Name)
		}
		right = fmt.Sprintf("%d:%d ", buf.CurRow+1, buf.CurCol+1)
	}
	if e.MsgLine != "" {
		left = " " + e.MsgLine
	}

	line := left
	for len([]rune(line)) < w-len([]rune(right)) {
		line += " "
	}
	line += right
	runes := []rune(line)
	if len(runes) > w {
		runes = runes[:w]
	}
	for i, ch := range runes {
		e.Screen.SetContent(x+i, y, ch, nil, st)
	}
}

func (e *Editor) drawModeLine(x, y, w int) {
	t := CurrentTheme
	st := tcell.StyleDefault.Background(t.TabActiveBG).Foreground(t.TabActiveFG)
	var prompt string
	switch e.Mode {
	case ModeFind:
		prompt = "Find: " + e.ModeInput
	case ModeCommand:
		prompt = "Command: " + e.ModeInput
	}
	drawText(e.Screen, x, y, w, prompt+"█", st)
}

func (e *Editor) drawBuffer(b *Buffer, x, y, w, h int) {
	t := CurrentTheme

	// Line number gutter width
	totalLines := len(b.Lines)
	lnw := len(fmt.Sprintf("%d", totalLines)) + 1
	e.lineNumWidth = lnw

	textX := x + lnw + 1
	textW := w - lnw - 1

	// Ensure scroll keeps cursor visible
	if b.CurRow < b.ScrollRow {
		b.ScrollRow = b.CurRow
	}
	if b.CurRow >= b.ScrollRow+h {
		b.ScrollRow = b.CurRow - h + 1
	}
	if b.CurCol < b.ScrollCol {
		b.ScrollCol = b.CurCol
	}
	if b.CurCol >= b.ScrollCol+textW {
		b.ScrollCol = b.CurCol - textW + 1
	}

	// Get highlighter
	hl := GetHighlighter(b.Name)
	if b.FilePath != "" {
		hl = GetHighlighter(b.FilePath)
	}

	// Build per-line token map
	hlState := &HLState{}

	// Selection bounds
	var selR1, selC1, selR2, selC2 int
	hasSel := b.HasSelection()
	if hasSel {
		selR1, selC1, selR2, selC2 = b.SelectionBounds()
	}

	for row := 0; row < h; row++ {
		lineIdx := b.ScrollRow + row
		sy := y + row

		// Line number
		lnSt := tcell.StyleDefault.Background(t.BG).Foreground(t.LineNumFG)
		if lineIdx == b.CurRow {
			lnSt = tcell.StyleDefault.Background(t.CursorLineBG).Foreground(t.FG)
		}
		lnStr := ""
		if lineIdx < len(b.Lines) {
			lnStr = fmt.Sprintf("%*d", lnw, lineIdx+1)
		}
		for i, ch := range lnStr {
			e.Screen.SetContent(x+i, sy, ch, nil, lnSt)
		}
		e.Screen.SetContent(x+lnw, sy, ' ', nil, lnSt)

		if lineIdx >= len(b.Lines) {
			continue
		}

		line := b.Lines[lineIdx]

		// Cursor line background
		lineBG := t.BG
		if lineIdx == b.CurRow {
			lineBG = t.CursorLineBG
		}

		// Get tokens for this line
		var tokens []Token
		if hl != nil {
			tokens = hl.Tokenize(b.Lines, lineIdx, hlState)
		}

		// Build style per column
		baseStyle := tcell.StyleDefault.Background(lineBG).Foreground(t.FG)

		// Render
		for col := 0; col < textW; col++ {
			charIdx := b.ScrollCol + col
			sx := textX + col

			var ch rune = ' '
			if charIdx < len(line) {
				ch = line[charIdx]
			}

			// Determine style
			st := baseStyle

			// Syntax highlight
			if charIdx < len(line) {
				for _, tok := range tokens {
					if charIdx >= tok.Col && charIdx < tok.Col+tok.Len {
						ts := TokenToStyle(tok.Kind)
						// Preserve background
						fg, _, _ := ts.Decompose()
						st = tcell.StyleDefault.Background(lineBG).Foreground(fg)
						if tok.Kind == TokKeyword {
							st = st.Bold(true)
						}
						if tok.Kind == TokComment {
							st = st.Italic(true)
						}
						break
					}
				}
			}

			// Selection overlay
			if hasSel {
				inSel := false
				if selR1 == selR2 {
					inSel = lineIdx == selR1 && charIdx >= selC1 && charIdx < selC2
				} else if lineIdx == selR1 {
					inSel = charIdx >= selC1
				} else if lineIdx == selR2 {
					inSel = charIdx < selC2
				} else if lineIdx > selR1 && lineIdx < selR2 {
					inSel = true
				}
				if inSel {
					st = tcell.StyleDefault.Background(t.SelectBG).Foreground(t.SelectFG)
				}
			}

			// Cursor
			if lineIdx == b.CurRow && charIdx == b.CurCol {
				e.Screen.ShowCursor(sx, sy)
			}

			e.Screen.SetContent(sx, sy, ch, nil, st)
		}
	}
}

// ---- Helpers ----

func drawText(screen tcell.Screen, x, y, maxW int, text string, st tcell.Style) {
	cx := x
	for _, ch := range text {
		if cx >= x+maxW {
			break
		}
		screen.SetContent(cx, y, ch, nil, st)
		cx++
	}
	for ; cx < x+maxW; cx++ {
		screen.SetContent(cx, y, ' ', nil, st)
	}
}

// SaveCurrentBuffer saves the active buffer
func (e *Editor) SaveCurrentBuffer() {
	b := e.ActiveBuffer()
	if b == nil {
		return
	}
	if b.FilePath == "" {
		e.MsgLine = "No file path (use :saveas <path>)"
		return
	}
	if err := b.Save(); err != nil {
		e.MsgLine = "Save error: " + err.Error()
	} else {
		e.MsgLine = "Saved: " + b.FilePath
	}
}

// ExecuteCommand handles command palette input
func (e *Editor) ExecuteCommand(cmd string) {
	cmd = strings.TrimSpace(cmd)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}
	switch parts[0] {
	case "theme":
		if len(parts) > 1 {
			if SetTheme(parts[1]) {
				e.MsgLine = "Theme: " + parts[1]
			} else {
				e.MsgLine = "Unknown theme. Available: dark, light, monokai, gruvbox"
			}
		} else {
			e.MsgLine = "Available themes: dark, light, monokai, gruvbox"
		}
	case "saveas":
		b := e.ActiveBuffer()
		if b != nil && len(parts) > 1 {
			b.FilePath = parts[1]
			b.Name = shortName(parts[1])
			e.SaveCurrentBuffer()
		}
	case "open":
		if len(parts) > 1 {
			e.OpenFile(parts[1])
		}
	case "new":
		name := "*scratch*"
		if len(parts) > 1 {
			name = parts[1]
		}
		e.NewBuffer(name)
	case "replace":
		b := e.ActiveBuffer()
		if b != nil && len(parts) >= 3 {
			n := b.ReplaceAll(parts[1], parts[2])
			e.MsgLine = fmt.Sprintf("Replaced %d occurrences", n)
		}
	case "q", "quit":
		os.Exit(0)
	case "w", "write":
		e.SaveCurrentBuffer()
	case "wq":
		e.SaveCurrentBuffer()
		os.Exit(0)
	case "ln", "linenum":
		b := e.ActiveBuffer()
		if b != nil {
			e.MsgLine = fmt.Sprintf("Line: %d / %d", b.CurRow+1, len(b.Lines))
		}
	case "cd":
		if len(parts) > 1 {
			os.Chdir(parts[1])
			e.Explorer.SetDir(parts[1])
			e.MsgLine = "Dir: " + parts[1]
		}
	default:
		e.MsgLine = "Unknown command: " + cmd
	}
}
