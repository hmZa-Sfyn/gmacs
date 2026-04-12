package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type ExplorerEntry struct {
	Name     string
	IsDir    bool
	Size     int64
	ModTime  string
	PermMode string
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

	// Add parent (..) and current (.) entries
	if e.Dir != "/" {
		e.Entries = append(e.Entries, ExplorerEntry{Name: "..", IsDir: true})
	} else {
		e.Entries = append(e.Entries, ExplorerEntry{Name: ".", IsDir: true})
	}

	var dirs, files []ExplorerEntry
	for _, en := range entries {
		nm := en.Name()
		if strings.HasPrefix(nm, ".") && nm != "." && nm != ".." {
			continue // skip hidden files
		}
		info, _ := en.Info()
		var size int64 = 0
		var modTime string = "-"
		var permMode string = "-"
		if info != nil {
			size = info.Size()
			modTime = info.ModTime().Format("2006-01-02")
			permMode = info.Mode().String()
		}
		entry := ExplorerEntry{nm, en.IsDir(), size, modTime, permMode}
		if en.IsDir() {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name < dirs[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	e.Entries = append(e.Entries, dirs...)
	e.Entries = append(e.Entries, files...)
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
	name := e.Entries[e.Sel].Name
	if name == ".." {
		return filepath.Dir(e.Dir)
	}
	if name == "." {
		return e.Dir
	}
	return filepath.Join(e.Dir, name)
}

func (e *Explorer) DeleteSelected() error {
	if len(e.Entries) == 0 || e.Sel >= len(e.Entries) {
		return nil
	}
	name := e.Entries[e.Sel].Name
	if name == ".." || name == "." {
		return nil
	}
	path := filepath.Join(e.Dir, name)
	err := os.RemoveAll(path)
	if err == nil {
		e.Refresh()
	}
	return err
}

func (e *Explorer) RenameSelected(newName string) error {
	if len(e.Entries) == 0 || e.Sel >= len(e.Entries) {
		return nil
	}
	oldName := e.Entries[e.Sel].Name
	if oldName == ".." || oldName == "." {
		return nil
	}
	old := filepath.Join(e.Dir, oldName)
	new := filepath.Join(e.Dir, newName)
	err := os.Rename(old, new)
	if err == nil {
		e.Refresh()
	}
	return err
}

func (e *Explorer) CreateFile(name string) error {
	path := filepath.Join(e.Dir, name)
	file, err := os.Create(path)
	if err == nil {
		file.Close()
		e.Refresh()
	}
	return err
}

func (e *Explorer) CreateDir(name string) error {
	path := filepath.Join(e.Dir, name)
	err := os.Mkdir(path, 0755)
	if err == nil {
		e.Refresh()
	}
	return err
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

	// Total line like ls -la
	total := fmt.Sprintf(" total %d", len(e.Entries))
	drawText(screen, x, y, e.Width, total, bgSt.Foreground(t.LineNumFG))
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
		name := en.Name
		if en.IsDir && name != "." && name != ".." {
			name += "/"
		}
		line := fmt.Sprintf("%s %s %8d %s", en.PermMode, en.ModTime, en.Size, name)
		if len(line) > e.Width {
			line = line[:e.Width-1] + "…"
		}
		for len(line) < e.Width {
			line += " "
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
		for cx, ch := range line {
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
