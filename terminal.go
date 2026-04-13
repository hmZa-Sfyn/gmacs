package main

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

// TermTab is a simple pseudo-terminal buffer
type TermTab struct {
	Name    string
	Lines   []string
	Input   string
	Cmd     *exec.Cmd
	Stdin   io.WriteCloser
	Scroll  int
	Running bool
}

func NewTermTab() *TermTab {
	t := &TermTab{Name: "*terminal*"}
	t.Lines = append(t.Lines, "$ ")
	t.startShell()
	return t
}

func (t *TermTab) startShell() {
	shell, args, env, promptCmd := t.getShellCommand()
	t.Cmd = exec.Command(shell, args...)
	t.Cmd.Env = append(os.Environ(), append(env, "TERM=xterm-256color", "INPUTRC=/dev/null")...)

	stdin, err := t.Cmd.StdinPipe()
	if err != nil {
		t.appendLine("Error starting shell: " + err.Error())
		return
	}
	t.Stdin = stdin

	stdout, err := t.Cmd.StdoutPipe()
	if err != nil {
		t.appendLine("Error: " + err.Error())
		return
	}
	stderr, err := t.Cmd.StderrPipe()
	if err != nil {
		t.appendLine("Error: " + err.Error())
		return
	}

	if err := t.Cmd.Start(); err != nil {
		t.appendLine("Error: " + err.Error())
		return
	}
	t.Running = true

	go t.readOutput(stdout)
	go t.readOutput(stderr)

	if promptCmd != "" {
		go func() {
			time.Sleep(100 * time.Millisecond)
			if t.Stdin != nil {
				io.WriteString(t.Stdin, promptCmd+"\n")
			}
		}()
	}
}

func (t *TermTab) getShellCommand() (string, []string, []string, string) {
	shell := "bash"
	args := []string{"--norc", "--noprofile"}
	env := []string{}
	promptCmd := "export PS1='$ '"
	return shell, args, env, promptCmd
}

func (t *TermTab) readOutput(r io.Reader) {
	buf := make([]byte, 1024)
	var partial strings.Builder

	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := stripAnsi(string(buf[:n]))
			partial.WriteString(chunk)
			for {
				text := partial.String()
				idx := strings.IndexByte(text, '\n')
				if idx < 0 {
					break
				}
				line := strings.TrimRight(text[:idx], "\r")
				t.appendLine(line)
				partial.Reset()
				partial.WriteString(text[idx+1:])
			}
		}
		if err != nil {
			if err != io.EOF {
				t.appendLine("Error: " + err.Error())
			}
			break
		}
	}
	if partial.Len() > 0 {
		t.appendLine(strings.TrimRight(partial.String(), "\r"))
	}
}

func (t *TermTab) SendInput(text string) {
	if t.Stdin == nil {
		return
	}
	io.WriteString(t.Stdin, text)
}

func (t *TermTab) appendLine(line string) {
	t.Lines = append(t.Lines, line)
	if len(t.Lines) > 2000 {
		t.Lines = t.Lines[len(t.Lines)-2000:]
	}
}

func (t *TermTab) TypeRune(r rune) {
	t.Input += string(r)
}

func (t *TermTab) Backspace() {
	if len(t.Input) > 0 {
		runes := []rune(t.Input)
		t.Input = string(runes[:len(runes)-1])
	}
}

func (t *TermTab) Submit() {
	if t.Stdin == nil {
		return
	}
	cmd := t.Input
	t.appendLine("$ " + cmd)
	t.Input = ""
	io.WriteString(t.Stdin, cmd+"\n")
}

func (t *TermTab) Draw(screen tcell.Screen, x, y, w, h int) {
	th := CurrentTheme
	bgSt := tcell.StyleDefault.Background(th.BG).Foreground(th.FG)
	promptSt := tcell.StyleDefault.Background(th.BG).Foreground(th.Keyword)

	// Adjust scroll to show last line
	visible := h - 1
	total := len(t.Lines)
	if t.Scroll < total-visible {
		t.Scroll = total - visible
	}
	if t.Scroll < 0 {
		t.Scroll = 0
	}

	for row := 0; row < visible; row++ {
		idx := t.Scroll + row
		if idx >= len(t.Lines) {
			break
		}
		drawText(screen, x, y+row, w, t.Lines[idx], bgSt)
	}
	// Input line
	prompt := t.Input + "█"
	drawText(screen, x, y+visible, w, prompt, promptSt)
}

// stripAnsi removes ANSI escape sequences
func stripAnsi(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' && s[i] != 'K' && s[i] != 'H' && s[i] != 'J' && s[i] != 'A' && s[i] != 'B' && s[i] != 'C' && s[i] != 'D' {
				i++
			}
			i++
			continue
		}
		if s[i] >= 32 || s[i] == '\t' {
			b.WriteByte(s[i])
		}
		i++
	}
	return b.String()
}
