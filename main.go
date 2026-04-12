package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
)

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := screen.Init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer screen.Fini()

	screen.SetStyle(tcell.StyleDefault)
	screen.EnableMouse()
	screen.Clear()

	// Load custom syntax and theme files
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	syntaxDir := filepath.Join(exeDir, "syntax")
	themesDir := filepath.Join(exeDir, "themes")

	LoadSyntaxFiles(syntaxDir)
	LoadThemeFiles(themesDir)

	ed := NewEditor(screen)

	// Open files from args
	if len(os.Args) > 1 {
		for _, f := range os.Args[1:] {
			ed.OpenFile(f)
		}
	} else {
		ed.NewBuffer("*empty*")
	}

	ed.Run()
}
