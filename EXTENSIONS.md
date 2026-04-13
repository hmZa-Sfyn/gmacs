# gmacs - Syntax & Theme Extension System

## Overview

gmacs supports extendable syntax highlighting and themes through simple configuration files.

## Directory Structure

```
gmacs/
├── syntax/              # Custom syntax highlighters (.syntax files)
├── themes/              # Custom color themes (.theme files)
└── main executable
```

## Creating Custom Syntax Highlighters

Syntax files use the `.syntax` extension and define regex-based coloring rules.

### Quick Start

Create `my-language.syntax`:

```
name: MyLanguage
files: *.mylang,*.ml

rule: keyword | \b(if|else|for|while|function)\b
rule: string | "(?:\\.|[^"\\])*"
rule: comment | //.*$
rule: number | \b\d+\b
```

Save it to `./syntax/my-language.syntax` and restart the editor.

See `syntax/README.md` for full documentation.

## Creating Custom Themes

Theme files use the `.theme` extension and define color schemes.

### Quick Start

Create `my-theme.theme`:

```
name: my_theme

bg: rgb(30, 30, 35)
fg: rgb(200, 200, 200)
keyword: rgb(100, 200, 255)
string: rgb(230, 150, 100)
comment: rgb(120, 160, 100)
```

Save it to `./themes/my-theme.theme` and use `theme my_theme` in the command palette.

See `themes/README.md` for full documentation and all available color keys.

## Loading Extensions

Extensions are automatically loaded from:
- `./syntax/` - Scanned for `.syntax` files
- `./themes/` - Scanned for `.theme` files

The editor searches for these directories relative to the executable location on startup.

## Built-in Examples

### Syntax Files
- `rust.syntax` - Rust language
- `typescript.syntax` - TypeScript/JavaScript
- `python.syntax` - Python

### Theme Files
- `custom_dark.theme` - Dark theme
- `custom_light.theme` - Light theme

## Command Palette Support

### Syntax Commands
- Syntax files are loaded automatically when the editor starts

### Theme Commands
- `theme dark` - Switch to dark theme
- `theme light` - Switch to light theme
- `theme custom_dark` - Switch to custom dark theme
- `theme custom_light` - Switch to custom light theme
- `theme my_theme` - Switch to any custom theme

## Format Details

### Syntax Files (.syntax)

```
name: LanguageName
files: *.ext,*.ext2,*.ext3
rule: TokenType | regex_pattern
```

Token types: `keyword`, `string`, `comment`, `number`, `type`, `function`, `operator`, `macro`, `constant`

Patterns use Go `regexp` syntax. Multiple rules are applied in order and overlapping tokens are merged.

### Theme Files (.theme)

```
name: theme_name
key: color_value
```

Color values:
- `rgb(r, g, b)` - RGB 0-255
- `#RRGGBB` - Hex
- Named: black, red, green, yellow, blue, magenta, cyan, white

## Tips

1. **Theme filename to command**: The theme name in the file, not the filename. Use `theme custom_dark` not `theme custom_dark.theme`
2. **Pattern syntax**: Use raw regex patterns. `\b` for word boundaries, `\.` for literal dots
3. **Multi-line patterns**: Use `[\s\S]*?` for non-greedy multi-line matching
4. **Testing**: Create a test file with your syntax and check the coloring
5. **Restart required**: New syntax/theme files require an editor restart to load

## Examples

See `syntax/README.md` and `themes/README.md` for detailed examples and all available configuration options.
