package main

import (
	"os"
	"strings"
	"unicode"
)

// Edit kinds for undo
const (
	EditInsert = iota
	EditDelete
)

type Edit struct {
	Kind int
	Row  int
	Col  int
	Text string
}

type Buffer struct {
	Name     string
	FilePath string
	Lines    [][]rune
	Dirty    bool

	// Cursor
	CurRow, CurCol int
	// Desired column for vertical movement
	DesiredCol int

	// Selection: anchor point (-1,-1 = none)
	SelRow, SelCol int

	// Undo/Redo stacks
	UndoStack []Edit
	RedoStack []Edit

	// Scroll offsets
	ScrollRow, ScrollCol int
}

func NewBuffer(name string) *Buffer {
	return &Buffer{
		Name:   name,
		Lines:  [][]rune{{}},
		SelRow: -1, SelCol: -1,
	}
}

func NewBufferFromFile(path string) (*Buffer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	b := NewBuffer(shortName(path))
	b.FilePath = path
	raw := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	b.Lines = make([][]rune, len(raw))
	for i, l := range raw {
		b.Lines[i] = []rune(l)
	}
	if len(b.Lines) == 0 {
		b.Lines = [][]rune{{}}
	}
	return b, nil
}

func (b *Buffer) Save() error {
	if b.FilePath == "" {
		return nil
	}
	var sb strings.Builder
	for i, l := range b.Lines {
		sb.WriteString(string(l))
		if i < len(b.Lines)-1 {
			sb.WriteByte('\n')
		}
	}
	err := os.WriteFile(b.FilePath, []byte(sb.String()), 0644)
	if err == nil {
		b.Dirty = false
	}
	return err
}

// ---- Cursor movement ----

func (b *Buffer) clampCursor() {
	if b.CurRow < 0 {
		b.CurRow = 0
	}
	if b.CurRow >= len(b.Lines) {
		b.CurRow = len(b.Lines) - 1
	}
	if b.CurCol < 0 {
		b.CurCol = 0
	}
	ll := len(b.Lines[b.CurRow])
	if b.CurCol > ll {
		b.CurCol = ll
	}
}

func (b *Buffer) MoveCursor(dr, dc int) {
	b.CurRow += dr
	b.CurCol += dc
	b.clampCursor()
	b.DesiredCol = b.CurCol
}

func (b *Buffer) MoveUp() {
	if b.CurRow > 0 {
		b.CurRow--
		b.CurCol = b.DesiredCol
		b.clampCursor()
	}
}

func (b *Buffer) MoveDown() {
	if b.CurRow < len(b.Lines)-1 {
		b.CurRow++
		b.CurCol = b.DesiredCol
		b.clampCursor()
	}
}

func (b *Buffer) MoveLeft() {
	if b.CurCol > 0 {
		b.CurCol--
	} else if b.CurRow > 0 {
		b.CurRow--
		b.CurCol = len(b.Lines[b.CurRow])
	}
	b.DesiredCol = b.CurCol
}

func (b *Buffer) MoveRight() {
	if b.CurCol < len(b.Lines[b.CurRow]) {
		b.CurCol++
	} else if b.CurRow < len(b.Lines)-1 {
		b.CurRow++
		b.CurCol = 0
	}
	b.DesiredCol = b.CurCol
}

func (b *Buffer) MoveWordLeft() {
	if b.CurCol == 0 {
		if b.CurRow > 0 {
			b.CurRow--
			b.CurCol = len(b.Lines[b.CurRow])
		}
		b.DesiredCol = b.CurCol
		return
	}
	line := b.Lines[b.CurRow]
	c := b.CurCol - 1
	for c > 0 && !isWordChar(line[c]) {
		c--
	}
	for c > 0 && isWordChar(line[c-1]) {
		c--
	}
	b.CurCol = c
	b.DesiredCol = b.CurCol
}

func (b *Buffer) MoveWordRight() {
	line := b.Lines[b.CurRow]
	c := b.CurCol
	if c >= len(line) {
		if b.CurRow < len(b.Lines)-1 {
			b.CurRow++
			b.CurCol = 0
		}
		b.DesiredCol = b.CurCol
		return
	}
	for c < len(line) && !isWordChar(line[c]) {
		c++
	}
	for c < len(line) && isWordChar(line[c]) {
		c++
	}
	b.CurCol = c
	b.DesiredCol = b.CurCol
}

func (b *Buffer) MoveLineStart() {
	b.CurCol = 0
	b.DesiredCol = 0
}

func (b *Buffer) MoveLineEnd() {
	b.CurCol = len(b.Lines[b.CurRow])
	b.DesiredCol = b.CurCol
}

func (b *Buffer) MoveFileStart() {
	b.CurRow, b.CurCol = 0, 0
	b.DesiredCol = 0
}

func (b *Buffer) MoveFileEnd() {
	b.CurRow = len(b.Lines) - 1
	b.CurCol = len(b.Lines[b.CurRow])
	b.DesiredCol = b.CurCol
}

// ---- Selection ----

func (b *Buffer) StartSelection() {
	b.SelRow = b.CurRow
	b.SelCol = b.CurCol
}

func (b *Buffer) ClearSelection() {
	b.SelRow = -1
	b.SelCol = -1
}

func (b *Buffer) HasSelection() bool {
	return b.SelRow >= 0
}

func (b *Buffer) clampPosition(row, col int) (int, int) {
	if row < 0 {
		row = 0
	}
	if row >= len(b.Lines) {
		row = len(b.Lines) - 1
	}
	if row < 0 {
		return 0, 0
	}
	if col < 0 {
		col = 0
	}
	lineLen := len(b.Lines[row])
	if col > lineLen {
		col = lineLen
	}
	return row, col
}

func (b *Buffer) SelectionBounds() (r1, c1, r2, c2 int) {
	r1, c1 = b.clampPosition(b.SelRow, b.SelCol)
	r2, c2 = b.clampPosition(b.CurRow, b.CurCol)
	if r1 > r2 || (r1 == r2 && c1 > c2) {
		r1, c1, r2, c2 = r2, c2, r1, c1
	}
	return
}

func (b *Buffer) SelectedText() string {
	if !b.HasSelection() {
		return ""
	}
	r1, c1, r2, c2 := b.SelectionBounds()
	if r1 == r2 {
		return string(b.Lines[r1][c1:c2])
	}
	var sb strings.Builder
	sb.WriteString(string(b.Lines[r1][c1:]))
	for row := r1 + 1; row < r2; row++ {
		sb.WriteByte('\n')
		sb.WriteString(string(b.Lines[row]))
	}
	sb.WriteByte('\n')
	sb.WriteString(string(b.Lines[r2][:c2]))
	return sb.String()
}

// ---- Edit operations ----

func (b *Buffer) DeleteSelection() {
	if !b.HasSelection() {
		return
	}
	r1, c1, r2, c2 := b.SelectionBounds()
	text := b.SelectedText()
	b.pushUndo(Edit{EditDelete, r1, c1, text})
	b.deleteRange(r1, c1, r2, c2)
	b.CurRow, b.CurCol = r1, c1
	b.DesiredCol = b.CurCol
	b.ClearSelection()
	b.Dirty = true
	b.RedoStack = nil
}

func (b *Buffer) deleteRange(r1, c1, r2, c2 int) {
	r1, c1 = b.clampPosition(r1, c1)
	r2, c2 = b.clampPosition(r2, c2)
	if r1 > r2 || (r1 == r2 && c1 >= c2) {
		return
	}
	if r1 == r2 {
		line := b.Lines[r1]
		b.Lines[r1] = append(line[:c1], line[c2:]...)
		return
	}
	head := b.Lines[r1][:c1]
	tail := b.Lines[r2][c2:]
	merged := make([]rune, len(head)+len(tail))
	copy(merged, head)
	copy(merged[len(head):], tail)
	b.Lines = append(b.Lines[:r1], b.Lines[r2+1:]...)
	b.Lines[r1] = merged
}

func (b *Buffer) InsertRune(r rune) {
	if b.HasSelection() {
		b.DeleteSelection()
	}
	line := b.Lines[b.CurRow]
	newLine := make([]rune, len(line)+1)
	copy(newLine, line[:b.CurCol])
	newLine[b.CurCol] = r
	copy(newLine[b.CurCol+1:], line[b.CurCol:])
	b.Lines[b.CurRow] = newLine
	b.pushUndo(Edit{EditInsert, b.CurRow, b.CurCol, string(r)})
	b.CurCol++
	b.DesiredCol = b.CurCol
	b.Dirty = true
	b.RedoStack = nil
}

func (b *Buffer) InsertNewline() {
	if b.HasSelection() {
		b.DeleteSelection()
	}
	line := b.Lines[b.CurRow]
	head := make([]rune, b.CurCol)
	copy(head, line[:b.CurCol])
	tail := make([]rune, len(line)-b.CurCol)
	copy(tail, line[b.CurCol:])

	// Auto-indent
	indent := leadingSpaces(b.Lines[b.CurRow])

	b.pushUndo(Edit{EditInsert, b.CurRow, b.CurCol, "\n"})
	b.Lines[b.CurRow] = head
	newLine := append(indent, tail...)
	after := make([][]rune, len(b.Lines)+1)
	copy(after, b.Lines[:b.CurRow+1])
	after[b.CurRow+1] = newLine
	copy(after[b.CurRow+2:], b.Lines[b.CurRow+1:])
	b.Lines = after
	b.CurRow++
	b.CurCol = len(indent)
	b.DesiredCol = b.CurCol
	b.Dirty = true
	b.RedoStack = nil
}

func (b *Buffer) Backspace() {
	if b.HasSelection() {
		b.DeleteSelection()
		return
	}
	if b.CurCol == 0 {
		if b.CurRow == 0 {
			return
		}
		// Merge with previous line
		prev := b.Lines[b.CurRow-1]
		cur := b.Lines[b.CurRow]
		b.pushUndo(Edit{EditDelete, b.CurRow - 1, len(prev), "\n"})
		merged := append(prev, cur...)
		b.Lines = append(b.Lines[:b.CurRow-1], b.Lines[b.CurRow:]...)
		b.Lines[b.CurRow-1] = merged
		b.CurRow--
		b.CurCol = len(prev)
	} else {
		line := b.Lines[b.CurRow]
		ch := string(line[b.CurCol-1])
		b.pushUndo(Edit{EditDelete, b.CurRow, b.CurCol - 1, ch})
		b.Lines[b.CurRow] = append(line[:b.CurCol-1], line[b.CurCol:]...)
		b.CurCol--
	}
	b.DesiredCol = b.CurCol
	b.Dirty = true
	b.RedoStack = nil
}

func (b *Buffer) DeleteForward() {
	if b.HasSelection() {
		b.DeleteSelection()
		return
	}
	line := b.Lines[b.CurRow]
	if b.CurCol == len(line) {
		if b.CurRow == len(b.Lines)-1 {
			return
		}
		next := b.Lines[b.CurRow+1]
		b.pushUndo(Edit{EditDelete, b.CurRow, b.CurCol, "\n"})
		merged := append(line, next...)
		b.Lines = append(b.Lines[:b.CurRow], b.Lines[b.CurRow+1:]...)
		b.Lines[b.CurRow] = merged
	} else {
		ch := string(line[b.CurCol])
		b.pushUndo(Edit{EditDelete, b.CurRow, b.CurCol, ch})
		b.Lines[b.CurRow] = append(line[:b.CurCol], line[b.CurCol+1:]...)
	}
	b.Dirty = true
	b.RedoStack = nil
}

func (b *Buffer) InsertText(text string) {
	for _, r := range text {
		if r == '\n' {
			b.InsertNewline()
		} else {
			b.InsertRune(r)
		}
	}
}

// MoveLineUp swaps current line with previous
func (b *Buffer) MoveLineUp() {
	if b.CurRow == 0 {
		return
	}
	b.Lines[b.CurRow], b.Lines[b.CurRow-1] = b.Lines[b.CurRow-1], b.Lines[b.CurRow]
	b.CurRow--
	b.Dirty = true
}

// MoveLineDown swaps current line with next
func (b *Buffer) MoveLineDown() {
	if b.CurRow >= len(b.Lines)-1 {
		return
	}
	b.Lines[b.CurRow], b.Lines[b.CurRow+1] = b.Lines[b.CurRow+1], b.Lines[b.CurRow]
	b.CurRow++
	b.Dirty = true
}

// ---- Undo/Redo ----

func (b *Buffer) pushUndo(e Edit) {
	b.UndoStack = append(b.UndoStack, e)
}

func (b *Buffer) Undo() {
	if len(b.UndoStack) == 0 {
		return
	}
	e := b.UndoStack[len(b.UndoStack)-1]
	b.UndoStack = b.UndoStack[:len(b.UndoStack)-1]
	b.RedoStack = append(b.RedoStack, e)
	b.applyInverse(e)
	b.Dirty = true
}

func (b *Buffer) Redo() {
	if len(b.RedoStack) == 0 {
		return
	}
	e := b.RedoStack[len(b.RedoStack)-1]
	b.RedoStack = b.RedoStack[:len(b.RedoStack)-1]
	b.UndoStack = append(b.UndoStack, e)
	b.applyEdit(e)
	b.Dirty = true
}

func (b *Buffer) applyEdit(e Edit) {
	b.CurRow, b.CurCol = e.Row, e.Col
	if e.Kind == EditInsert {
		b.insertRaw(e.Text)
	} else {
		b.deleteRaw(e)
	}
}

func (b *Buffer) applyInverse(e Edit) {
	b.CurRow, b.CurCol = e.Row, e.Col
	if e.Kind == EditInsert {
		// Undo insert = delete
		lines := strings.Split(e.Text, "\n")
		endRow := e.Row + len(lines) - 1
		endCol := e.Col + len([]rune(lines[len(lines)-1]))
		if len(lines) > 1 {
			endCol = len([]rune(lines[len(lines)-1]))
		}
		b.deleteRange(e.Row, e.Col, endRow, endCol)
	} else {
		// Undo delete = insert
		b.insertRaw(e.Text)
	}
}

func (b *Buffer) insertRaw(text string) {
	for _, r := range text {
		if r == '\n' {
			line := b.Lines[b.CurRow]
			head := make([]rune, b.CurCol)
			copy(head, line[:b.CurCol])
			tail := make([]rune, len(line)-b.CurCol)
			copy(tail, line[b.CurCol:])
			b.Lines[b.CurRow] = head
			after := make([][]rune, len(b.Lines)+1)
			copy(after, b.Lines[:b.CurRow+1])
			after[b.CurRow+1] = tail
			copy(after[b.CurRow+2:], b.Lines[b.CurRow+1:])
			b.Lines = after
			b.CurRow++
			b.CurCol = 0
		} else {
			line := b.Lines[b.CurRow]
			newLine := make([]rune, len(line)+1)
			copy(newLine, line[:b.CurCol])
			newLine[b.CurCol] = r
			copy(newLine[b.CurCol+1:], line[b.CurCol:])
			b.Lines[b.CurRow] = newLine
			b.CurCol++
		}
	}
}

func (b *Buffer) deleteRaw(e Edit) {
	lines := strings.Split(e.Text, "\n")
	endRow := e.Row + len(lines) - 1
	endCol := e.Col + len([]rune(lines[len(lines)-1]))
	if len(lines) > 1 {
		endCol = len([]rune(lines[len(lines)-1]))
	}
	b.deleteRange(e.Row, e.Col, endRow, endCol)
}

// ---- Find/Replace ----

// FindNext finds next occurrence of needle from current cursor, returns (row, col) or (-1,-1)
func (b *Buffer) FindNext(needle string, fromRow, fromCol int) (int, int) {
	if needle == "" {
		return -1, -1
	}
	nr := []rune(needle)
	total := len(b.Lines)
	for i := 0; i < total; i++ {
		row := (fromRow + i) % total
		line := b.Lines[row]
		startCol := 0
		if i == 0 {
			startCol = fromCol
		}
		for c := startCol; c <= len(line)-len(nr); c++ {
			if runesEqual(line[c:c+len(nr)], nr) {
				return row, c
			}
		}
	}
	return -1, -1
}

func (b *Buffer) ReplaceAll(find, replace string) int {
	count := 0
	fr := []rune(find)
	rr := []rune(replace)
	for row := range b.Lines {
		line := b.Lines[row]
		var newLine []rune
		i := 0
		for i <= len(line)-len(fr) {
			if runesEqual(line[i:i+len(fr)], fr) {
				newLine = append(newLine, rr...)
				i += len(fr)
				count++
			} else {
				newLine = append(newLine, line[i])
				i++
			}
		}
		newLine = append(newLine, line[i:]...)
		b.Lines[row] = newLine
	}
	if count > 0 {
		b.Dirty = true
	}
	return count
}

// ---- Helpers ----

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func runesEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func leadingSpaces(line []rune) []rune {
	for i, r := range line {
		if r != ' ' && r != '\t' {
			return line[:i]
		}
	}
	return line
}

func shortName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}
