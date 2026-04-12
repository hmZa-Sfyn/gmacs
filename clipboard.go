package main

import (
	"bytes"
	"os/exec"
)

// realClipboard overrides the stubs in keybindings.go with real exec calls

func setClipboardReal(text string) bool {
	cmds := []struct {
		name string
		args []string
	}{
		{"xclip", []string{"-selection", "clipboard"}},
		{"xsel", []string{"--clipboard", "--input"}},
		{"pbcopy", nil},
		{"wl-copy", nil},
	}
	for _, c := range cmds {
		cmd := exec.Command(c.name, c.args...)
		cmd.Stdin = bytes.NewBufferString(text)
		if err := cmd.Run(); err == nil {
			return true
		}
	}
	return false
}

func getClipboardReal() (string, bool) {
	cmds := []struct {
		name string
		args []string
	}{
		{"xclip", []string{"-selection", "clipboard", "-o"}},
		{"xsel", []string{"--clipboard", "--output"}},
		{"pbpaste", nil},
		{"wl-paste", nil},
	}
	for _, c := range cmds {
		out, err := exec.Command(c.name, c.args...).Output()
		if err == nil {
			return string(out), true
		}
	}
	return "", false
}

func init() {
	// Monkey-patch the clipboard functions via package-level vars
	_setClipboard = func(text string) {
		if !setClipboardReal(text) {
			clipboardInternal = text
		}
	}
	_getClipboard = func() string {
		if t, ok := getClipboardReal(); ok {
			return t
		}
		return clipboardInternal
	}
}

// Function vars so keybindings.go can call them
var _setClipboard func(string)
var _getClipboard func() string
