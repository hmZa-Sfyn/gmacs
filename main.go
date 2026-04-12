package main

import (
	"fmt"
	"os"

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

	ed := NewEditor(screen)

	// Open files from args
	if len(os.Args) > 1 {
		for _, f := range os.Args[1:] {
			ed.OpenFile(f)
		}
	} else {
		ed.NewBuffer("*scratch*")
	}

	ed.Run()
}
