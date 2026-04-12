package main

import "github.com/gdamore/tcell/v2"

// Theme defines all editor colors
type Theme struct {
	Name string

	// UI chrome
	BG          tcell.Color
	FG          tcell.Color
	StatusBG    tcell.Color
	StatusFG    tcell.Color
	TabActiveBG tcell.Color
	TabActiveFG tcell.Color
	TabBG       tcell.Color
	TabFG       tcell.Color
	LineNumFG   tcell.Color
	CursorLineBG tcell.Color
	SelectBG    tcell.Color
	SelectFG    tcell.Color

	// Syntax
	Keyword   tcell.Color
	String    tcell.Color
	Comment   tcell.Color
	Number    tcell.Color
	Type      tcell.Color
	Function  tcell.Color
	Operator  tcell.Color
	Macro     tcell.Color
	Constant  tcell.Color

	// Explorer
	ExplorerBG    tcell.Color
	ExplorerFG    tcell.Color
	ExplorerSelBG tcell.Color
	ExplorerDirFG tcell.Color
}

var Themes = map[string]Theme{
	"dark": {
		Name:          "dark",
		BG:            tcell.NewRGBColor(24, 24, 27),
		FG:            tcell.NewRGBColor(212, 212, 212),
		StatusBG:      tcell.NewRGBColor(37, 99, 235),
		StatusFG:      tcell.NewRGBColor(255, 255, 255),
		TabActiveBG:   tcell.NewRGBColor(37, 37, 38),
		TabActiveFG:   tcell.NewRGBColor(255, 255, 255),
		TabBG:         tcell.NewRGBColor(30, 30, 30),
		TabFG:         tcell.NewRGBColor(150, 150, 150),
		LineNumFG:     tcell.NewRGBColor(80, 80, 100),
		CursorLineBG:  tcell.NewRGBColor(35, 35, 45),
		SelectBG:      tcell.NewRGBColor(38, 79, 120),
		SelectFG:      tcell.NewRGBColor(255, 255, 255),
		Keyword:       tcell.NewRGBColor(86, 156, 214),
		String:        tcell.NewRGBColor(206, 145, 120),
		Comment:       tcell.NewRGBColor(106, 153, 85),
		Number:        tcell.NewRGBColor(181, 206, 168),
		Type:          tcell.NewRGBColor(78, 201, 176),
		Function:      tcell.NewRGBColor(220, 220, 170),
		Operator:      tcell.NewRGBColor(212, 212, 212),
		Macro:         tcell.NewRGBColor(197, 134, 192),
		Constant:      tcell.NewRGBColor(100, 200, 200),
		ExplorerBG:    tcell.NewRGBColor(30, 30, 35),
		ExplorerFG:    tcell.NewRGBColor(200, 200, 200),
		ExplorerSelBG: tcell.NewRGBColor(50, 90, 150),
		ExplorerDirFG: tcell.NewRGBColor(86, 156, 214),
	},
	"light": {
		Name:          "light",
		BG:            tcell.NewRGBColor(255, 255, 255),
		FG:            tcell.NewRGBColor(30, 30, 30),
		StatusBG:      tcell.NewRGBColor(0, 90, 185),
		StatusFG:      tcell.NewRGBColor(255, 255, 255),
		TabActiveBG:   tcell.NewRGBColor(230, 230, 230),
		TabActiveFG:   tcell.NewRGBColor(0, 0, 0),
		TabBG:         tcell.NewRGBColor(240, 240, 240),
		TabFG:         tcell.NewRGBColor(120, 120, 120),
		LineNumFG:     tcell.NewRGBColor(180, 180, 180),
		CursorLineBG:  tcell.NewRGBColor(245, 245, 255),
		SelectBG:      tcell.NewRGBColor(180, 210, 240),
		SelectFG:      tcell.NewRGBColor(0, 0, 0),
		Keyword:       tcell.NewRGBColor(0, 0, 200),
		String:        tcell.NewRGBColor(180, 60, 0),
		Comment:       tcell.NewRGBColor(80, 140, 80),
		Number:        tcell.NewRGBColor(100, 0, 150),
		Type:          tcell.NewRGBColor(0, 140, 120),
		Function:      tcell.NewRGBColor(130, 100, 0),
		Operator:      tcell.NewRGBColor(30, 30, 30),
		Macro:         tcell.NewRGBColor(150, 0, 180),
		Constant:      tcell.NewRGBColor(0, 130, 130),
		ExplorerBG:    tcell.NewRGBColor(245, 245, 245),
		ExplorerFG:    tcell.NewRGBColor(30, 30, 30),
		ExplorerSelBG: tcell.NewRGBColor(180, 210, 240),
		ExplorerDirFG: tcell.NewRGBColor(0, 80, 180),
	},
	"monokai": {
		Name:          "monokai",
		BG:            tcell.NewRGBColor(39, 40, 34),
		FG:            tcell.NewRGBColor(248, 248, 242),
		StatusBG:      tcell.NewRGBColor(102, 217, 239),
		StatusFG:      tcell.NewRGBColor(39, 40, 34),
		TabActiveBG:   tcell.NewRGBColor(60, 63, 52),
		TabActiveFG:   tcell.NewRGBColor(248, 248, 242),
		TabBG:         tcell.NewRGBColor(45, 46, 40),
		TabFG:         tcell.NewRGBColor(120, 120, 100),
		LineNumFG:     tcell.NewRGBColor(90, 90, 75),
		CursorLineBG:  tcell.NewRGBColor(50, 51, 44),
		SelectBG:      tcell.NewRGBColor(73, 72, 62),
		SelectFG:      tcell.NewRGBColor(248, 248, 242),
		Keyword:       tcell.NewRGBColor(249, 38, 114),
		String:        tcell.NewRGBColor(230, 219, 116),
		Comment:       tcell.NewRGBColor(117, 113, 94),
		Number:        tcell.NewRGBColor(174, 129, 255),
		Type:          tcell.NewRGBColor(102, 217, 239),
		Function:      tcell.NewRGBColor(166, 226, 46),
		Operator:      tcell.NewRGBColor(249, 38, 114),
		Macro:         tcell.NewRGBColor(249, 38, 114),
		Constant:      tcell.NewRGBColor(174, 129, 255),
		ExplorerBG:    tcell.NewRGBColor(35, 36, 31),
		ExplorerFG:    tcell.NewRGBColor(248, 248, 242),
		ExplorerSelBG: tcell.NewRGBColor(73, 72, 62),
		ExplorerDirFG: tcell.NewRGBColor(102, 217, 239),
	},
	"gruvbox": {
		Name:          "gruvbox",
		BG:            tcell.NewRGBColor(40, 40, 40),
		FG:            tcell.NewRGBColor(235, 219, 178),
		StatusBG:      tcell.NewRGBColor(152, 151, 26),
		StatusFG:      tcell.NewRGBColor(40, 40, 40),
		TabActiveBG:   tcell.NewRGBColor(60, 56, 54),
		TabActiveFG:   tcell.NewRGBColor(235, 219, 178),
		TabBG:         tcell.NewRGBColor(50, 48, 47),
		TabFG:         tcell.NewRGBColor(146, 131, 116),
		LineNumFG:     tcell.NewRGBColor(102, 92, 84),
		CursorLineBG:  tcell.NewRGBColor(50, 48, 47),
		SelectBG:      tcell.NewRGBColor(80, 73, 69),
		SelectFG:      tcell.NewRGBColor(235, 219, 178),
		Keyword:       tcell.NewRGBColor(251, 73, 52),
		String:        tcell.NewRGBColor(184, 187, 38),
		Comment:       tcell.NewRGBColor(146, 131, 116),
		Number:        tcell.NewRGBColor(211, 134, 155),
		Type:          tcell.NewRGBColor(250, 189, 47),
		Function:      tcell.NewRGBColor(184, 187, 38),
		Operator:      tcell.NewRGBColor(254, 128, 25),
		Macro:         tcell.NewRGBColor(251, 73, 52),
		Constant:      tcell.NewRGBColor(211, 134, 155),
		ExplorerBG:    tcell.NewRGBColor(36, 36, 36),
		ExplorerFG:    tcell.NewRGBColor(235, 219, 178),
		ExplorerSelBG: tcell.NewRGBColor(80, 73, 69),
		ExplorerDirFG: tcell.NewRGBColor(250, 189, 47),
	},
}

var CurrentTheme = Themes["dark"]

func SetTheme(name string) bool {
	if t, ok := Themes[name]; ok {
		CurrentTheme = t
		return true
	}
	return false
}
