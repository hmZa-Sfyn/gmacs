package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// HandleKey processes a key event. Returns false to quit.
func (e *Editor) HandleKey(ev *tcell.EventKey) bool {
	switch e.Mode {
	case ModeFind:
		return e.handleFindKey(ev)
	case ModeCommand:
		return e.handleCommandKey(ev)
	case ModeExplorerFullscreen:
		return e.handleExplorerFullscreenKey(ev)
	case ModeTerminalFullscreen:
		return e.handleTerminalFullscreenKey(ev)
	case ModeExplorerRename:
		return e.handleExplorerRenameKey(ev)
	}

	if e.ShowExplorer {
		if e.handleExplorerKey(ev) {
			return true
		}
	}

	tab := e.ActiveTab()
	if tab == nil {
		return true
	}
	if tab.Kind == TabTerminal {
		return e.handleTerminalKey(ev, tab.Term)
	}
	return e.handleBufferKey(ev, tab.Buffer)
}

func (e *Editor) handleBufferKey(ev *tcell.EventKey, b *Buffer) bool {
	mod := ev.Modifiers()
	key := ev.Key()
	ch := ev.Rune()

	ctrl := mod&tcell.ModCtrl != 0
	shift := mod&tcell.ModShift != 0
	alt := mod&tcell.ModAlt != 0

	e.MsgLine = ""

	if ctrl {
		switch key {
		case tcell.KeyCtrlS:
			e.SaveCurrentBuffer()
			return true
		case tcell.KeyCtrlZ:
			b.Undo()
			return true
		case tcell.KeyCtrlY:
			b.Redo()
			return true
		case tcell.KeyCtrlC:
			if b.HasSelection() {
				setClipboard(b.SelectedText())
				e.MsgLine = "Copied"
			}
			return true
		case tcell.KeyCtrlX:
			if b.HasSelection() {
				setClipboard(b.SelectedText())
				b.DeleteSelection()
				e.MsgLine = "Cut"
			}
			return true
		case tcell.KeyCtrlV:
			if text := getClipboard(); text != "" {
				b.InsertText(text)
			}
			return true
		case tcell.KeyCtrlD:
			b.DeleteSelection()
			return true
		case tcell.KeyCtrlF:
			e.Mode = ModeFind
			e.ModeInput = ""
			return true
		case tcell.KeyCtrlH:
			e.Mode = ModeCommand
			e.ModeInput = "replace "
			return true
		case tcell.KeyCtrlT:
			e.ToggleTerminal()
			return true
		case tcell.KeyCtrlE:
			if e.Mode == ModeExplorerFullscreen {
				e.Mode = ModeNormal
			} else {
				e.Mode = ModeExplorerFullscreen
			}
			return true
		case tcell.KeyCtrlO:
			e.ShowExplorer = true
			return true
		case tcell.KeyCtrlW:
			e.CloseTab(e.Active)
			return true
		case tcell.KeyCtrlN:
			if e.Active < len(e.Tabs)-1 {
				e.Active++
			}
			return true
		case tcell.KeyCtrlP:
			if e.Active > 0 {
				e.Active--
			}
			return true
		case tcell.KeyLeft:
			if shift {
				if !b.HasSelection() {
					b.StartSelection()
				}
				b.MoveWordLeft()
			} else {
				b.ClearSelection()
				b.MoveWordLeft()
			}
			return true
		case tcell.KeyRight:
			if shift {
				if !b.HasSelection() {
					b.StartSelection()
				}
				b.MoveWordRight()
			} else {
				b.ClearSelection()
				b.MoveWordRight()
			}
			return true
		case tcell.KeyHome:
			b.MoveFileStart()
			return true
		case tcell.KeyEnd:
			b.MoveFileEnd()
			return true
		case tcell.KeyRune:
			switch ch {
			case ':':
				e.Mode = ModeCommand
				e.ModeInput = ""
				return true
			case ';':
				e.Mode = ModeCommand
				e.ModeInput = ""
				return true
			case 'a':
				b.MoveFileStart()
				b.StartSelection()
				b.MoveFileEnd()
				return true
			}
		}
	}

	if alt {
		switch key {
		case tcell.KeyUp:
			if shift {
				if !b.HasSelection() {
					b.StartSelection()
				}
			} else {
				b.ClearSelection()
			}
			b.MoveLineUp()
			return true
		case tcell.KeyDown:
			if shift {
				if !b.HasSelection() {
					b.StartSelection()
				}
			} else {
				b.ClearSelection()
			}
			b.MoveLineDown()
			return true
		case tcell.KeyLeft:
			if shift {
				if !b.HasSelection() {
					b.StartSelection()
				}
			} else {
				b.ClearSelection()
			}
			for i := 0; i < 10; i++ {
				b.MoveLeft()
			}
			return true
		case tcell.KeyRight:
			if shift {
				if !b.HasSelection() {
					b.StartSelection()
				}
			} else {
				b.ClearSelection()
			}
			for i := 0; i < 10; i++ {
				b.MoveRight()
			}
			return true
		case tcell.KeyRune:
			if ch == 'e' || ch == 'E' {
				if b.FilePath != "" {
					e.Explorer.SetDir(filepath.Dir(b.FilePath))
				}
				e.ShowExplorer = true
				return true
			}
			if ch >= '1' && ch <= '9' {
				idx := int(ch - '1')
				if idx < len(e.Tabs) {
					e.Active = idx
				}
				return true
			}
			if ch == '0' && len(e.Tabs) >= 10 {
				e.Active = 9
				return true
			}
		}
	}

	switch key {
	case tcell.KeyUp:
		if shift {
			if !b.HasSelection() {
				b.StartSelection()
			}
			b.MoveUp()
		} else {
			b.ClearSelection()
			b.MoveUp()
		}
	case tcell.KeyDown:
		if shift {
			if !b.HasSelection() {
				b.StartSelection()
			}
			b.MoveDown()
		} else {
			b.ClearSelection()
			b.MoveDown()
		}
	case tcell.KeyLeft:
		if shift {
			if !b.HasSelection() {
				b.StartSelection()
			}
			b.MoveLeft()
		} else {
			b.ClearSelection()
			b.MoveLeft()
		}
	case tcell.KeyRight:
		if shift {
			if !b.HasSelection() {
				b.StartSelection()
			}
			b.MoveRight()
		} else {
			b.ClearSelection()
			b.MoveRight()
		}
	case tcell.KeyHome:
		if shift {
			if !b.HasSelection() {
				b.StartSelection()
			}
		} else {
			b.ClearSelection()
		}
		b.MoveLineStart()
	case tcell.KeyEnd:
		if shift {
			if !b.HasSelection() {
				b.StartSelection()
			}
		} else {
			b.ClearSelection()
		}
		b.MoveLineEnd()
	case tcell.KeyPgUp:
		_, h := e.Screen.Size()
		for i := 0; i < h-2; i++ {
			b.MoveUp()
		}
	case tcell.KeyPgDn:
		_, h := e.Screen.Size()
		for i := 0; i < h-2; i++ {
			b.MoveDown()
		}
	case tcell.KeyEnter:
		b.InsertNewline()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		b.Backspace()
	case tcell.KeyDelete:
		b.DeleteForward()
	case tcell.KeyTab:
		b.InsertText("    ")
	case tcell.KeyEscape:
		b.ClearSelection()
		e.Mode = ModeNormal
		e.ModeInput = ""
	case tcell.KeyRune:
		if !ctrl && alt {
			switch ch {
			case ';', ':':
				e.Mode = ModeCommand
				e.ModeInput = ""
			default:
				b.InsertRune(ch)
			}
		}
		b.InsertRune(ch)
	}
	return true
}

func (e *Editor) handleFindKey(ev *tcell.EventKey) bool {
	b := e.ActiveBuffer()
	switch ev.Key() {
	case tcell.KeyEscape:
		e.Mode = ModeNormal
		e.ModeInput = ""
	case tcell.KeyEnter:
		e.FindStr = e.ModeInput
		if b != nil {
			row, col := b.FindNext(e.FindStr, b.CurRow, b.CurCol+1)
			if row >= 0 {
				b.SelRow, b.SelCol = row, col
				b.CurRow = row
				b.CurCol = col + len([]rune(e.FindStr))
				b.DesiredCol = b.CurCol
				e.MsgLine = "Found"
			} else {
				e.MsgLine = "Not found: " + e.FindStr
			}
		}
		e.Mode = ModeNormal
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.ModeInput) > 0 {
			r := []rune(e.ModeInput)
			e.ModeInput = string(r[:len(r)-1])
		}
	case tcell.KeyRune:
		e.ModeInput += string(ev.Rune())
	}
	return true
}

func (e *Editor) handleCommandKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		e.Mode = ModeNormal
		e.ModeInput = ""
	case tcell.KeyEnter:
		e.ExecuteCommand(e.ModeInput)
		e.Mode = ModeNormal
		e.ModeInput = ""
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.ModeInput) > 0 {
			r := []rune(e.ModeInput)
			e.ModeInput = string(r[:len(r)-1])
		}
	case tcell.KeyRune:
		e.ModeInput += string(ev.Rune())
	}
	return true
}

func (exp *Explorer) handleKey(ev *tcell.EventKey, ed *Editor) bool {
	switch ev.Key() {
	case tcell.KeyUp:
		exp.MoveUp()
		return true
	case tcell.KeyDown:
		exp.MoveDown()
		return true
	case tcell.KeyEnter:
		if exp.SelectedIsDir() {
			exp.SetDir(exp.SelectedPath())
		} else if path := exp.SelectedPath(); path != "" {
			ed.OpenFile(path)
		}
		return true
	case tcell.KeyEscape:
		ed.ShowExplorer = false
		return true
	case tcell.KeyCtrlP:
		exp.ParentDir()
		return true
	}
	return false
}

func (e *Editor) handleExplorerKey(ev *tcell.EventKey) bool {
	return e.Explorer.handleKey(ev, e)
}

func (e *Editor) handleTerminalKey(ev *tcell.EventKey, term *TermTab) bool {
	mod := ev.Modifiers()
	key := ev.Key()
	ch := ev.Rune()
	ctrl := mod&tcell.ModCtrl != 0

	switch key {
	case tcell.KeyCtrlT:
		for i := len(e.Tabs) - 1; i >= 0; i-- {
			if e.Tabs[i].Kind == TabBuffer {
				e.Active = i
				return true
			}
		}
	case tcell.KeyEnter:
		term.Submit()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		term.Backspace()
	case tcell.KeyTab:
		term.SendInput("\t")
	case tcell.KeyUp:
		term.SendInput("\x1b[A")
	case tcell.KeyDown:
		term.SendInput("\x1b[B")
	case tcell.KeyLeft:
		term.SendInput("\x1b[D")
	case tcell.KeyRight:
		term.SendInput("\x1b[C")
	case tcell.KeyDelete:
		term.SendInput("\x1b[3~")
	case tcell.KeyHome:
		term.SendInput("\x1b[H")
	case tcell.KeyEnd:
		term.SendInput("\x1b[F")
	case tcell.KeyCtrlC:
		term.SendInput("\x03")
	case tcell.KeyCtrlD:
		term.SendInput("\x04")
	case tcell.KeyRune:
		if !ctrl && (ch == ':' || ch == ';') {
			e.Mode = ModeCommand
			e.ModeInput = ""
			return true
		}
		term.TypeRune(ch)
	}
	return true
}

func (e *Editor) handleExplorerFullscreenKey(ev *tcell.EventKey) bool {
	key := ev.Key()
	ch := ev.Rune()

	switch key {
	case tcell.KeyUp:
		e.Explorer.MoveUp()
		return true
	case tcell.KeyDown:
		e.Explorer.MoveDown()
		return true
	case tcell.KeyEscape:
		e.Mode = ModeNormal
		return true
	case tcell.KeyEnter:
		path := e.Explorer.SelectedPath()
		if e.Explorer.SelectedIsDir() {
			e.Explorer.SetDir(path)
		} else {
			e.OpenFile(path)
			e.Mode = ModeNormal
		}
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		e.Explorer.ParentDir()
		return true
	case tcell.KeyCtrlE:
		e.Mode = ModeNormal
		return true
	case tcell.KeyRune:
		switch ch {
		case 'd':
			err := e.Explorer.DeleteSelected()
			if err != nil {
				e.MsgLine = "Error: " + err.Error()
			}
			return true
		case 'c':
			e.Mode = ModeCommand
			e.ModeInput = "newfile "
			return true
		case 'C':
			e.Mode = ModeCommand
			e.ModeInput = "mkdir "
			return true
		case 'r':
			if len(e.Explorer.Entries) > e.Explorer.Sel {
				e.RenameInput = e.Explorer.Entries[e.Explorer.Sel].Name
				e.Mode = ModeExplorerRename
			}
			return true
		}
	}
	return false
}

func (e *Editor) handleTerminalFullscreenKey(ev *tcell.EventKey) bool {
	mod := ev.Modifiers()
	key := ev.Key()
	ch := ev.Rune()
	ctrl := mod&tcell.ModCtrl != 0

	if tab := e.ActiveTab(); tab != nil && tab.Kind == TabTerminal {
		switch key {
		case tcell.KeyCtrlT:
			e.Mode = ModeNormal
			return true
		case tcell.KeyEnter:
			tab.Term.Submit()
			return true
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			tab.Term.Backspace()
			return true
		case tcell.KeyTab:
			tab.Term.SendInput("\t")
			return true
		case tcell.KeyUp:
			tab.Term.SendInput("\x1b[A")
			return true
		case tcell.KeyDown:
			tab.Term.SendInput("\x1b[B")
			return true
		case tcell.KeyLeft:
			tab.Term.SendInput("\x1b[D")
			return true
		case tcell.KeyRight:
			tab.Term.SendInput("\x1b[C")
			return true
		case tcell.KeyDelete:
			tab.Term.SendInput("\x1b[3~")
			return true
		case tcell.KeyHome:
			tab.Term.SendInput("\x1b[H")
			return true
		case tcell.KeyEnd:
			tab.Term.SendInput("\x1b[F")
			return true
		case tcell.KeyCtrlC:
			tab.Term.SendInput("\x03")
			return true
		case tcell.KeyCtrlD:
			tab.Term.SendInput("\x04")
			return true
		case tcell.KeyRune:
			if !ctrl && (ch == ':' || ch == ';') {
				e.Mode = ModeCommand
				e.ModeInput = ""
				return true
			}
			tab.Term.TypeRune(ch)
			return true
		}
	}
	return false
}

func (e *Editor) handleExplorerRenameKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		e.Mode = ModeExplorerFullscreen
		e.RenameInput = ""
		return true
	case tcell.KeyEnter:
		if e.RenameInput != "" {
			err := e.Explorer.RenameSelected(e.RenameInput)
			if err != nil {
				e.MsgLine = "Error: " + err.Error()
			}
		}
		e.Mode = ModeExplorerFullscreen
		e.RenameInput = ""
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.RenameInput) > 0 {
			e.RenameInput = e.RenameInput[:len(e.RenameInput)-1]
		}
		return true
	case tcell.KeyRune:
		e.RenameInput += string(ev.Rune())
		return true
	}
	return true
}

// HandleMouse handles mouse events
func (e *Editor) HandleMouse(ev *tcell.EventMouse) {
	mx, my := ev.Position()
	btn := ev.Buttons()

	if btn == tcell.Button1 {
		if my == 0 {
			e.clickTabBar(mx)
			return
		}
		if e.ShowExplorer && mx < e.Explorer.Width {
			e.clickExplorer(my)
			return
		}
		b := e.ActiveBuffer()
		if b == nil {
			return
		}
		explorerOffset := 0
		if e.ShowExplorer {
			explorerOffset = e.Explorer.Width
		}
		lnw := e.lineNumWidth
		textX := explorerOffset + lnw + 1
		if mx < textX {
			return
		}
		col := b.ScrollCol + (mx - textX)
		row := b.ScrollRow + (my - 1)
		if row < 0 {
			row = 0
		}
		if row >= len(b.Lines) {
			row = len(b.Lines) - 1
		}
		if col > len(b.Lines[row]) {
			col = len(b.Lines[row])
		}
		b.CurRow, b.CurCol = row, col
		b.DesiredCol = col
		b.ClearSelection()
	}

	if btn == tcell.WheelUp {
		if b := e.ActiveBuffer(); b != nil {
			for i := 0; i < 3; i++ {
				b.MoveUp()
			}
		}
	}
	if btn == tcell.WheelDown {
		if b := e.ActiveBuffer(); b != nil {
			for i := 0; i < 3; i++ {
				b.MoveDown()
			}
		}
	}
}

func (e *Editor) clickTabBar(mx int) {
	cx := 0
	for i, tab := range e.Tabs {
		name := " " + tab.Name() + " "
		end := cx + len([]rune(name)) + 1
		if mx >= cx && mx < end {
			e.Active = i
			return
		}
		cx = end
	}
}

func (e *Editor) clickExplorer(my int) {
	row := my - 3
	if row < 0 {
		return
	}
	idx := e.Explorer.Scroll + row
	if idx < len(e.Explorer.Entries) {
		e.Explorer.Sel = idx
		if e.Explorer.SelectedIsDir() {
			e.Explorer.SetDir(e.Explorer.SelectedPath())
		} else {
			e.OpenFile(e.Explorer.SelectedPath())
		}
	}
}

// ---- Clipboard ----

var clipboardInternal string

func setClipboard(text string) {
	if _setClipboard != nil {
		_setClipboard(text)
	} else {
		clipboardInternal = text
	}
}

func getClipboard() string {
	if _getClipboard != nil {
		return _getClipboard()
	}
	return clipboardInternal
}

// ---- Helpers ----

func findExec(name string) string {
	for _, p := range strings.Split(os.Getenv("PATH"), ":") {
		full := filepath.Join(p, name)
		if _, err := os.Stat(full); err == nil {
			return full
		}
	}
	return name
}
