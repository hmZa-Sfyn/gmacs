package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type ExplorerEntry struct {
	Name  string
	IsDir bool
}

type Explorer struct {
	Dir     string
	Entries []ExplorerEntry
	Sel     int
	Scroll  int
	Width   int
}

func NewExplorer(dir string) *Explorer {
	e := &Explorer{Width: 28}
	e.SetDir(dir)
	return e
}

func (e *Explorer) SetDir(dir string) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	e.Dir = abs
	e.Refresh()
	e.Sel = 0
	e.Scroll = 0
}

func (e *Explorer) Refresh() {
	entries, err := os.ReadDir(e.Dir)
	if err != nil {
		e.Entries = nil
		return
	}
	e.Entries = nil
	// Dirs first, then files
	var dirs, files []ExplorerEntry
	for _, en := range entries {
		nm := en.Name()
		if strings.HasPrefix(nm, ".") {
			continue // skip hidden
		}
		if en.IsDir() {
			dirs = append(dirs, ExplorerEntry{nm, true})
		} else {
			files = append(files, ExplorerEntry{nm, false})
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name < dirs[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	e.Entries = append(dirs, files...)
}

func (e *Explorer) MoveUp() {
	if e.Sel > 0 {
		e.Sel--
	}
}

func (e *Explorer) MoveDown() {
	if e.Sel < len(e.Entries)-1 {
		e.Sel++
	}
}

func (e *Explorer) ParentDir() {
	parent := filepath.Dir(e.Dir)
	if parent != e.Dir {
		e.SetDir(parent)
	}
}

// SelectedPath returns the full path of selected entry
func (e *Explorer) SelectedPath() string {
	if len(e.Entries) == 0 {
		return ""
	}
	if e.Sel >= len(e.Entries) {
		e.Sel = len(e.Entries) - 1
	}
	return filepath.Join(e.Dir, e.Entries[e.Sel].Name)
}

func (e *Explorer) SelectedIsDir() bool {
	if len(e.Entries) == 0 || e.Sel >= len(e.Entries) {
		return false
	}
	return e.Entries[e.Sel].IsDir
}

func (e *Explorer) Draw(screen tcell.Screen, x, y, h int) {
	t := CurrentTheme
	bgSt := tcell.StyleDefault.Background(t.ExplorerBG).Foreground(t.ExplorerFG)
	selSt := tcell.StyleDefault.Background(t.ExplorerSelBG).Foreground(t.ExplorerFG)
	dirSt := tcell.StyleDefault.Background(t.ExplorerBG).Foreground(t.ExplorerDirFG)
	selDirSt := tcell.StyleDefault.Background(t.ExplorerSelBG).Foreground(t.ExplorerDirFG)

	// Title bar
	titleSt := tcell.StyleDefault.Background(t.StatusBG).Foreground(t.StatusFG).Bold(true)
	drawText(screen, x, y, e.Width, " 📁 Explorer", titleSt)
	y++
	h--

	// Dir path
	dirLabel := shortPath(e.Dir, e.Width-2)
	drawText(screen, x, y, e.Width, " "+dirLabel, bgSt.Foreground(t.LineNumFG))
	y++
	h--

	// Adjust scroll
	if e.Sel < e.Scroll {
		e.Scroll = e.Sel
	}
	if e.Sel >= e.Scroll+h {
		e.Scroll = e.Sel - h + 1
	}

	for row := 0; row < h; row++ {
		idx := e.Scroll + row
		if idx >= len(e.Entries) {
			// Clear rest
			drawText(screen, x, y+row, e.Width, "", bgSt)
			continue
		}
		en := e.Entries[idx]
		prefix := "  "
		if en.IsDir {
			prefix = "▸ "
		}
		label := prefix + en.Name
		if len(label) > e.Width {
			label = label[:e.Width-1] + "…"
		}
		// Pad
		for len(label) < e.Width {
			label += " "
		}
		st := bgSt
		if en.IsDir {
			st = dirSt
		}
		if idx == e.Sel {
			if en.IsDir {
				st = selDirSt
			} else {
				st = selSt
			}
		}
		for cx, ch := range label {
			screen.SetContent(x+cx, y+row, ch, nil, st)
		}
	}
}

func shortPath(path string, max int) string {
	if len(path) <= max {
		return path
	}
	return "…" + path[len(path)-max+1:]
}
