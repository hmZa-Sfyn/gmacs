package main

import (
	"path/filepath"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// TokenKind represents a syntax token type
type TokenKind int

const (
	TokNormal TokenKind = iota
	TokKeyword
	TokString
	TokComment
	TokNumber
	TokType
	TokFunction
	TokOperator
	TokMacro
	TokConstant
)

// Token is a colored span on a line
type Token struct {
	Col  int
	Len  int
	Kind TokenKind
}

// Highlighter is a pluggable interface — implement this for any language
type Highlighter interface {
	// Tokenize returns tokens for line `lineIdx` given full lines context
	// (for multi-line comment state, etc.)
	Tokenize(lines [][]rune, lineIdx int, state *HLState) []Token
	// FileMatch returns true if this highlighter handles the given filename
	FileMatch(name string) bool
}

// HLState carries multi-line state (e.g. inside block comment)
type HLState struct {
	InBlockComment bool
	InMLString     bool
}

// Registry of highlighters — add your own here!
var Highlighters []Highlighter

func init() {
	Highlighters = append(Highlighters,
		&GoHighlighter{},
		&PythonHighlighter{},
		&ShellHighlighter{},
		&CHighlighter{},
		&JSONHighlighter{},
		&MarkdownHighlighter{},
	)
}

// GetHighlighter finds the best highlighter for a filename
func GetHighlighter(name string) Highlighter {
	for _, h := range Highlighters {
		if h.FileMatch(name) {
			return h
		}
	}
	return nil
}

// TokenToStyle maps token kinds to theme colors
func TokenToStyle(kind TokenKind) tcell.Style {
	t := CurrentTheme
	base := tcell.StyleDefault.Background(t.BG)
	switch kind {
	case TokKeyword:
		return base.Foreground(t.Keyword).Bold(true)
	case TokString:
		return base.Foreground(t.String)
	case TokComment:
		return base.Foreground(t.Comment).Italic(true)
	case TokNumber:
		return base.Foreground(t.Number)
	case TokType:
		return base.Foreground(t.Type)
	case TokFunction:
		return base.Foreground(t.Function)
	case TokOperator:
		return base.Foreground(t.Operator)
	case TokMacro:
		return base.Foreground(t.Macro)
	case TokConstant:
		return base.Foreground(t.Constant)
	default:
		return base.Foreground(t.FG)
	}
}

// ---- Shared helpers ----

func matchKeyword(line []rune, col int, kw string) bool {
	kr := []rune(kw)
	if col+len(kr) > len(line) {
		return false
	}
	if !runesEqual(line[col:col+len(kr)], kr) {
		return false
	}
	// Check word boundary after
	end := col + len(kr)
	if end < len(line) && isWordChar(line[end]) {
		return false
	}
	// Check word boundary before
	if col > 0 && isWordChar(line[col-1]) {
		return false
	}
	return true
}

func isDigitStart(r rune) bool {
	return unicode.IsDigit(r)
}

// ===================== GO =====================

type GoHighlighter struct{}

func (g *GoHighlighter) FileMatch(name string) bool {
	return strings.HasSuffix(name, ".go")
}

var goKeywords = []string{
	"break", "case", "chan", "const", "continue",
	"default", "defer", "else", "fallthrough", "for",
	"func", "go", "goto", "if", "import",
	"interface", "map", "package", "range", "return",
	"select", "struct", "switch", "type", "var",
}

var goTypes = []string{
	"bool", "byte", "complex64", "complex128", "error",
	"float32", "float64", "int", "int8", "int16", "int32", "int64",
	"rune", "string", "uint", "uint8", "uint16", "uint32", "uint64",
	"uintptr", "any",
}

var goConstants = []string{
	"true", "false", "nil", "iota",
}

func (g *GoHighlighter) Tokenize(lines [][]rune, lineIdx int, state *HLState) []Token {
	var tokens []Token
	line := lines[lineIdx]
	i := 0

	// Block comment continuation
	if state.InBlockComment {
		end := findStr(line, 0, "*/")
		if end >= 0 {
			tokens = append(tokens, Token{0, end + 2, TokComment})
			i = end + 2
			state.InBlockComment = false
		} else {
			return []Token{{0, len(line), TokComment}}
		}
	}

	for i < len(line) {
		// Line comment
		if i+1 < len(line) && line[i] == '/' && line[i+1] == '/' {
			tokens = append(tokens, Token{i, len(line) - i, TokComment})
			break
		}
		// Block comment start
		if i+1 < len(line) && line[i] == '/' && line[i+1] == '*' {
			end := findStr(line, i+2, "*/")
			if end >= 0 {
				tokens = append(tokens, Token{i, end + 2 - i, TokComment})
				i = end + 2
				continue
			} else {
				tokens = append(tokens, Token{i, len(line) - i, TokComment})
				state.InBlockComment = true
				break
			}
		}
		// String
		if line[i] == '"' || line[i] == '`' || line[i] == '\'' {
			end, tok := scanString(line, i)
			tokens = append(tokens, tok)
			i = end
			continue
		}
		// Number
		if isDigitStart(line[i]) || (line[i] == '.' && i+1 < len(line) && isDigitStart(line[i+1])) {
			end := scanNumber(line, i)
			tokens = append(tokens, Token{i, end - i, TokNumber})
			i = end
			continue
		}
		// Keywords / types / constants / identifiers
		if isWordChar(line[i]) {
			end := i
			for end < len(line) && isWordChar(line[end]) {
				end++
			}
			word := string(line[i:end])
			kind := TokNormal
			for _, kw := range goKeywords {
				if word == kw {
					kind = TokKeyword
					break
				}
			}
			if kind == TokNormal {
				for _, tp := range goTypes {
					if word == tp {
						kind = TokType
						break
					}
				}
			}
			if kind == TokNormal {
				for _, c := range goConstants {
					if word == c {
						kind = TokConstant
						break
					}
				}
			}
			// Function call heuristic: identifier followed by '('
			if kind == TokNormal && end < len(line) && line[end] == '(' {
				kind = TokFunction
			}
			if kind != TokNormal {
				tokens = append(tokens, Token{i, end - i, kind})
			}
			i = end
			continue
		}
		i++
	}
	return tokens
}

// ===================== PYTHON =====================

type PythonHighlighter struct{}

func (p *PythonHighlighter) FileMatch(name string) bool {
	return strings.HasSuffix(name, ".py")
}

var pyKeywords = []string{
	"and", "as", "assert", "async", "await",
	"break", "class", "continue", "def", "del",
	"elif", "else", "except", "finally", "for",
	"from", "global", "if", "import", "in",
	"is", "lambda", "nonlocal", "not", "or",
	"pass", "raise", "return", "try", "while",
	"with", "yield",
}
var pyConstants = []string{"True", "False", "None"}
var pyBuiltins = []string{
	"print", "len", "range", "type", "int", "str", "float",
	"list", "dict", "set", "tuple", "open", "input", "super",
	"enumerate", "zip", "map", "filter", "sorted", "reversed",
}

func (p *PythonHighlighter) Tokenize(lines [][]rune, lineIdx int, state *HLState) []Token {
	var tokens []Token
	line := lines[lineIdx]
	i := 0
	for i < len(line) {
		// Comment
		if line[i] == '#' {
			tokens = append(tokens, Token{i, len(line) - i, TokComment})
			break
		}
		// String
		if line[i] == '"' || line[i] == '\'' {
			end, tok := scanString(line, i)
			tokens = append(tokens, tok)
			i = end
			continue
		}
		// Decorator
		if line[i] == '@' {
			end := i + 1
			for end < len(line) && (isWordChar(line[end]) || line[end] == '.') {
				end++
			}
			tokens = append(tokens, Token{i, end - i, TokMacro})
			i = end
			continue
		}
		// Number
		if isDigitStart(line[i]) {
			end := scanNumber(line, i)
			tokens = append(tokens, Token{i, end - i, TokNumber})
			i = end
			continue
		}
		if isWordChar(line[i]) {
			end := i
			for end < len(line) && isWordChar(line[end]) {
				end++
			}
			word := string(line[i:end])
			kind := TokNormal
			for _, kw := range pyKeywords {
				if word == kw {
					kind = TokKeyword
					break
				}
			}
			if kind == TokNormal {
				for _, c := range pyConstants {
					if word == c {
						kind = TokConstant
						break
					}
				}
			}
			if kind == TokNormal {
				for _, b := range pyBuiltins {
					if word == b {
						kind = TokFunction
						break
					}
				}
			}
			if kind == TokNormal && end < len(line) && line[end] == '(' {
				kind = TokFunction
			}
			if kind != TokNormal {
				tokens = append(tokens, Token{i, end - i, kind})
			}
			i = end
			continue
		}
		i++
	}
	return tokens
}

// ===================== SHELL =====================

type ShellHighlighter struct{}

func (s *ShellHighlighter) FileMatch(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".sh" || ext == ".bash" || ext == ".zsh" || strings.HasSuffix(name, "Dockerfile")
}

var shKeywords = []string{
	"if", "then", "else", "elif", "fi", "for", "while",
	"do", "done", "case", "esac", "function", "in",
	"return", "exit", "export", "local", "readonly",
	"echo", "cd", "ls", "mkdir", "rm", "cp", "mv",
}

func (s *ShellHighlighter) Tokenize(lines [][]rune, lineIdx int, state *HLState) []Token {
	var tokens []Token
	line := lines[lineIdx]
	i := 0
	for i < len(line) {
		if line[i] == '#' {
			tokens = append(tokens, Token{i, len(line) - i, TokComment})
			break
		}
		if line[i] == '"' || line[i] == '\'' {
			end, tok := scanString(line, i)
			tokens = append(tokens, tok)
			i = end
			continue
		}
		if line[i] == '$' {
			end := i + 1
			for end < len(line) && (isWordChar(line[end]) || line[end] == '{' || line[end] == '}') {
				end++
			}
			tokens = append(tokens, Token{i, end - i, TokConstant})
			i = end
			continue
		}
		if isWordChar(line[i]) {
			end := i
			for end < len(line) && isWordChar(line[end]) {
				end++
			}
			word := string(line[i:end])
			kind := TokNormal
			for _, kw := range shKeywords {
				if word == kw {
					kind = TokKeyword
					break
				}
			}
			if kind != TokNormal {
				tokens = append(tokens, Token{i, end - i, kind})
			}
			i = end
			continue
		}
		i++
	}
	return tokens
}

// ===================== C/C++ =====================

type CHighlighter struct{}

func (c *CHighlighter) FileMatch(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".c" || ext == ".h" || ext == ".cpp" || ext == ".cc" || ext == ".cxx" || ext == ".hpp"
}

var cKeywords = []string{
	"auto", "break", "case", "char", "const", "continue",
	"default", "do", "double", "else", "enum", "extern",
	"float", "for", "goto", "if", "inline", "int", "long",
	"register", "restrict", "return", "short", "signed",
	"sizeof", "static", "struct", "switch", "typedef", "union",
	"unsigned", "void", "volatile", "while",
	// C++
	"class", "namespace", "new", "delete", "this", "template",
	"typename", "virtual", "override", "nullptr", "true", "false",
	"public", "private", "protected", "try", "catch", "throw",
}

func (c *CHighlighter) Tokenize(lines [][]rune, lineIdx int, state *HLState) []Token {
	var tokens []Token
	line := lines[lineIdx]
	i := 0

	if state.InBlockComment {
		end := findStr(line, 0, "*/")
		if end >= 0 {
			tokens = append(tokens, Token{0, end + 2, TokComment})
			i = end + 2
			state.InBlockComment = false
		} else {
			return []Token{{0, len(line), TokComment}}
		}
	}

	for i < len(line) {
		if i+1 < len(line) && line[i] == '/' && line[i+1] == '/' {
			tokens = append(tokens, Token{i, len(line) - i, TokComment})
			break
		}
		if i+1 < len(line) && line[i] == '/' && line[i+1] == '*' {
			end := findStr(line, i+2, "*/")
			if end >= 0 {
				tokens = append(tokens, Token{i, end + 2 - i, TokComment})
				i = end + 2
			} else {
				tokens = append(tokens, Token{i, len(line) - i, TokComment})
				state.InBlockComment = true
				break
			}
			continue
		}
		// Preprocessor
		if line[i] == '#' {
			tokens = append(tokens, Token{i, len(line) - i, TokMacro})
			break
		}
		if line[i] == '"' || line[i] == '\'' {
			end, tok := scanString(line, i)
			tokens = append(tokens, tok)
			i = end
			continue
		}
		if isDigitStart(line[i]) {
			end := scanNumber(line, i)
			tokens = append(tokens, Token{i, end - i, TokNumber})
			i = end
			continue
		}
		if isWordChar(line[i]) {
			end := i
			for end < len(line) && isWordChar(line[end]) {
				end++
			}
			word := string(line[i:end])
			kind := TokNormal
			for _, kw := range cKeywords {
				if word == kw {
					kind = TokKeyword
					break
				}
			}
			if kind == TokNormal && end < len(line) && line[end] == '(' {
				kind = TokFunction
			}
			if kind != TokNormal {
				tokens = append(tokens, Token{i, end - i, kind})
			}
			i = end
			continue
		}
		i++
	}
	return tokens
}

// ===================== JSON =====================

type JSONHighlighter struct{}

func (j *JSONHighlighter) FileMatch(name string) bool {
	return strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".jsonc")
}

func (j *JSONHighlighter) Tokenize(lines [][]rune, lineIdx int, state *HLState) []Token {
	var tokens []Token
	line := lines[lineIdx]
	i := 0
	for i < len(line) {
		if line[i] == '"' {
			end, tok := scanString(line, i)
			// Peek after string: if ':' follows (with spaces), it's a key (type color)
			k := end
			for k < len(line) && line[k] == ' ' {
				k++
			}
			if k < len(line) && line[k] == ':' {
				tok.Kind = TokType
			}
			tokens = append(tokens, tok)
			i = end
			continue
		}
		if isDigitStart(line[i]) || (line[i] == '-' && i+1 < len(line) && isDigitStart(line[i+1])) {
			end := i + 1
			if line[i] == '-' {
				end = scanNumber(line, i+1)
			} else {
				end = scanNumber(line, i)
			}
			tokens = append(tokens, Token{i, end - i, TokNumber})
			i = end
			continue
		}
		for _, kw := range []string{"true", "false", "null"} {
			if matchKeyword(line, i, kw) {
				tokens = append(tokens, Token{i, len(kw), TokConstant})
				i += len(kw)
				goto nextJSON
			}
		}
		i++
	nextJSON:
	}
	return tokens
}

// ===================== MARKDOWN =====================

type MarkdownHighlighter struct{}

func (m *MarkdownHighlighter) FileMatch(name string) bool {
	return strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".markdown")
}

func (m *MarkdownHighlighter) Tokenize(lines [][]rune, lineIdx int, state *HLState) []Token {
	var tokens []Token
	line := lines[lineIdx]
	if len(line) == 0 {
		return tokens
	}
	// Heading
	if line[0] == '#' {
		return []Token{{0, len(line), TokKeyword}}
	}
	// Code block
	if len(line) >= 3 && line[0] == '`' && line[1] == '`' && line[2] == '`' {
		return []Token{{0, len(line), TokComment}}
	}
	// List item
	if line[0] == '-' || line[0] == '*' || line[0] == '+' {
		tokens = append(tokens, Token{0, 1, TokOperator})
	}
	// Bold/italic - simple inline
	i := 0
	for i < len(line) {
		if line[i] == '`' {
			end := i + 1
			for end < len(line) && line[end] != '`' {
				end++
			}
			tokens = append(tokens, Token{i, end - i + 1, TokString})
			i = end + 1
			continue
		}
		if line[i] == '[' {
			end := i + 1
			for end < len(line) && line[end] != ']' {
				end++
			}
			tokens = append(tokens, Token{i, end - i + 1, TokFunction})
			i = end + 1
			continue
		}
		i++
	}
	return tokens
}

// ===================== Shared scanner helpers =====================

func scanString(line []rune, start int) (int, Token) {
	quote := line[start]
	i := start + 1
	for i < len(line) {
		if line[i] == '\\' {
			i += 2
			continue
		}
		if line[i] == quote {
			i++
			break
		}
		i++
	}
	return i, Token{start, i - start, TokString}
}

func scanNumber(line []rune, start int) int {
	i := start
	for i < len(line) && (unicode.IsDigit(line[i]) || line[i] == '.' || line[i] == 'x' ||
		line[i] == 'X' || line[i] == 'e' || line[i] == 'E' ||
		(line[i] >= 'a' && line[i] <= 'f') || (line[i] >= 'A' && line[i] <= 'F') ||
		line[i] == '_') {
		i++
	}
	return i
}

func findStr(line []rune, start int, needle string) int {
	nr := []rune(needle)
	for i := start; i <= len(line)-len(nr); i++ {
		if runesEqual(line[i:i+len(nr)], nr) {
			return i
		}
	}
	return -1
}
