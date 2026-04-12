package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SyntaxRule defines a regex-based syntax coloring rule
type SyntaxRule struct {
	Pattern *regexp.Regexp
	Kind    TokenKind
}

// CustomHighlighter is a regex-based highlighter loaded from .syntax files
type CustomHighlighter struct {
	Name      string
	FileGlobs []string
	Rules     []SyntaxRule
}

// FileMatch implements Highlighter
func (ch *CustomHighlighter) FileMatch(name string) bool {
	for _, glob := range ch.FileGlobs {
		if matched, _ := filepath.Match(glob, name); matched {
			return true
		}
	}
	return false
}

// Tokenize implements Highlighter
func (ch *CustomHighlighter) Tokenize(lines [][]rune, lineIdx int, state *HLState) []Token {
	if lineIdx >= len(lines) {
		return nil
	}
	line := lines[lineIdx]
	lineStr := string(line)
	var tokens []Token

	for _, rule := range ch.Rules {
		matches := rule.Pattern.FindAllStringIndex(lineStr, -1)
		for _, match := range matches {
			tokens = append(tokens, Token{
				Col:  match[0],
				Len:  match[1] - match[0],
				Kind: rule.Kind,
			})
		}
	}

	// Remove overlapping tokens, keeping first
	if len(tokens) > 1 {
		tokens = mergeTokens(tokens)
	}
	return tokens
}

func mergeTokens(tokens []Token) []Token {
	if len(tokens) == 0 {
		return tokens
	}
	var result []Token
	for _, t := range tokens {
		overlaps := false
		for _, r := range result {
			if t.Col >= r.Col && t.Col < r.Col+r.Len {
				overlaps = true
				break
			}
			if r.Col >= t.Col && r.Col < t.Col+t.Len {
				overlaps = true
				break
			}
		}
		if !overlaps {
			result = append(result, t)
		}
	}
	return result
}

// LoadSyntaxFiles loads all .syntax files from the syntax directory
func LoadSyntaxFiles(syntaxDir string) {
	entries, err := os.ReadDir(syntaxDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".syntax") {
			path := filepath.Join(syntaxDir, entry.Name())
			if hl, err := parseSyntaxFile(path); err == nil && hl != nil {
				Highlighters = append(Highlighters, hl)
			}
		}
	}
}

func parseSyntaxFile(path string) (*CustomHighlighter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hl := &CustomHighlighter{
		FileGlobs: []string{},
		Rules:     []SyntaxRule{},
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "name:") {
			hl.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		} else if strings.HasPrefix(line, "files:") {
			filesStr := strings.TrimSpace(strings.TrimPrefix(line, "files:"))
			hl.FileGlobs = strings.Split(filesStr, ",")
			for i := range hl.FileGlobs {
				hl.FileGlobs[i] = strings.TrimSpace(hl.FileGlobs[i])
			}
		} else if strings.HasPrefix(line, "rule:") {
			parts := strings.Split(strings.TrimPrefix(line, "rule:"), "|")
			if len(parts) == 2 {
				kindStr := strings.TrimSpace(parts[0])
				patternStr := strings.TrimSpace(parts[1])
				kind := parseTokenKind(kindStr)
				if pattern, err := regexp.Compile(patternStr); err == nil {
					hl.Rules = append(hl.Rules, SyntaxRule{Pattern: pattern, Kind: kind})
				}
			}
		}
	}

	if len(hl.Rules) == 0 {
		return nil, nil
	}

	return hl, nil
}

func parseTokenKind(s string) TokenKind {
	switch strings.ToLower(s) {
	case "keyword":
		return TokKeyword
	case "string":
		return TokString
	case "comment":
		return TokComment
	case "number":
		return TokNumber
	case "type":
		return TokType
	case "function":
		return TokFunction
	case "operator":
		return TokOperator
	case "macro":
		return TokMacro
	case "constant":
		return TokConstant
	default:
		return TokNormal
	}
}
