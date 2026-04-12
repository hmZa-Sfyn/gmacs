# Custom Syntax Files (.syntax)

Create custom syntax highlighters using regex-based rules.

## Format

```
name: LanguageName
files: *.ext,*.ext2
rule: tokentype | regex_pattern
rule: tokentype | regex_pattern
```

## Token Types

- `keyword` - Keywords and control flow
- `string` - String literals
- `comment` - Comments
- `number` - Numeric literals
- `type` - Type names
- `function` - Function names
- `operator` - Operators
- `macro` - Macros/decorators
- `constant` - Constants

## Example

```
name: Ruby
files: *.rb

rule: keyword | \b(def|end|if|elsif|else|unless|case|when|for|in|while|until|begin|rescue|ensure|module|class|return|yield|break|next|super|self|true|false|nil)\b
rule: string | "(?:\\.|[^"\\])*"
rule: string | '(?:\\.|[^'\\])*'
rule: comment | #.*$
rule: number | \b\d+(\.\d+)?\b
rule: type | \b[A-Z][a-zA-Z0-9]*\b
rule: function | \b([a-z_]\w*)\s*[\(\[]
```

## Notes

- Patterns are regex expressions (Go `regexp` syntax)
- Rules are applied in order and overlapping matches are merged
- File patterns use glob syntax (*, ?)
- Multiple file patterns separated by commas
