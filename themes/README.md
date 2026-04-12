# Custom Theme Files (.theme)

Create custom editor color schemes.

## Format

```
name: ThemeName
key: value
key: value
```

## Color Formats

- `rgb(r, g, b)` - RGB values 0-255
- `#RRGGBB` - Hex color
- Named colors: black, red, green, yellow, blue, magenta, cyan, white

## Available Keys

### UI Colors
- `bg` - Background
- `fg` - Foreground text
- `status-bg` - Status bar background
- `status-fg` - Status bar text
- `tab-active-bg` - Active tab background
- `tab-active-fg` - Active tab text
- `tab-bg` - Inactive tab background
- `tab-fg` - Inactive tab text
- `linenum-fg` - Line number color
- `cursor-line-bg` - Current line background
- `select-bg` - Selection background
- `select-fg` - Selection text

### Syntax Highlighting Colors
- `keyword` - Keywords
- `string` - Strings
- `comment` - Comments
- `number` - Numbers
- `type` - Type names
- `function` - Function names
- `operator` - Operators
- `macro` - Macros
- `constant` - Constants

### Explorer Colors
- `explorer-bg` - Explorer background
- `explorer-fg` - Explorer text
- `explorer-sel-bg` - Selected item background
- `explorer-dir-fg` - Directory color

## Example

```
name: solarized_dark

bg: rgb(7, 54, 66)
fg: rgb(131, 148, 150)
status-bg: rgb(38, 139, 210)
status-fg: rgb(7, 54, 66)

keyword: rgb(181, 137, 0)
string: rgb(42, 161, 152)
comment: rgb(101, 123, 131)
number: rgb(181, 137, 0)
type: rgb(108, 113, 196)
function: rgb(38, 139, 210)
```

## Placement

Place `.theme` files in the `./themes` directory relative to the executable.
