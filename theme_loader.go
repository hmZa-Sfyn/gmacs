package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// LoadThemeFiles loads all .theme files from the themes directory
func LoadThemeFiles(themesDir string) {
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".theme") {
			path := filepath.Join(themesDir, entry.Name())
			if theme, err := parseThemeFile(path); err == nil {
				Themes[theme.Name] = theme
			}
		}
	}
}

func parseThemeFile(path string) (Theme, error) {
	file, err := os.Open(path)
	if err != nil {
		return Theme{}, err
	}
	defer file.Close()

	theme := Theme{
		// defaults
		BG:      tcell.ColorBlack,
		FG:      tcell.ColorWhite,
		Keyword: tcell.ColorYellow,
		String:  tcell.ColorGreen,
		Comment: tcell.ColorBlue,
		Number:  tcell.Color103,
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "name:") {
			theme.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		} else {
			parseThemeLine(&theme, line)
		}
	}

	if theme.Name == "" {
		return Theme{}, fmt.Errorf("theme missing name")
	}

	return theme, nil
}

func parseThemeLine(theme *Theme, line string) {
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	color := parseColor(value)

	switch key {
	case "bg":
		theme.BG = color
	case "fg":
		theme.FG = color
	case "status-bg":
		theme.StatusBG = color
	case "status-fg":
		theme.StatusFG = color
	case "tab-active-bg":
		theme.TabActiveBG = color
	case "tab-active-fg":
		theme.TabActiveFG = color
	case "tab-bg":
		theme.TabBG = color
	case "tab-fg":
		theme.TabFG = color
	case "linenum-fg":
		theme.LineNumFG = color
	case "cursor-line-bg":
		theme.CursorLineBG = color
	case "select-bg":
		theme.SelectBG = color
	case "select-fg":
		theme.SelectFG = color
	case "keyword":
		theme.Keyword = color
	case "string":
		theme.String = color
	case "comment":
		theme.Comment = color
	case "number":
		theme.Number = color
	case "type":
		theme.Type = color
	case "function":
		theme.Function = color
	case "operator":
		theme.Operator = color
	case "macro":
		theme.Macro = color
	case "constant":
		theme.Constant = color
	case "explorer-bg":
		theme.ExplorerBG = color
	case "explorer-fg":
		theme.ExplorerFG = color
	case "explorer-sel-bg":
		theme.ExplorerSelBG = color
	case "explorer-dir-fg":
		theme.ExplorerDirFG = color
	}
}

func parseColor(s string) tcell.Color {
	s = strings.TrimSpace(s)

	// Try rgb(r, g, b) format
	if strings.HasPrefix(s, "rgb(") && strings.HasSuffix(s, ")") {
		inner := s[4 : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 3 {
			r, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			g, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
			b, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
			return tcell.NewRGBColor(int32(r), int32(g), int32(b))
		}
	}

	// Try #RRGGBB format
	if strings.HasPrefix(s, "#") && len(s) == 7 {
		var r, g, b int
		fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
		return tcell.NewRGBColor(int32(r), int32(g), int32(b))
	}

	// Named colors
	switch strings.ToLower(s) {
	case "black":
		return tcell.ColorBlack
	case "red":
		return tcell.ColorRed
	case "green":
		return tcell.ColorGreen
	case "yellow":
		return tcell.ColorYellow
	case "blue":
		return tcell.ColorBlue
	case "magenta":
		return tcell.ColorFuchsia
	case "cyan":
		return tcell.ColorAqua
	case "white":
		return tcell.ColorWhite
	default:
		return tcell.ColorWhite
	}
}
