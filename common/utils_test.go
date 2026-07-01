package common

import (
	"fmt"
	"strings"
	"testing"
)

func TestEscapeGraphQLString(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"plain text", "Fix login bug", "Fix login bug"},
		{
			"literal newline (the reported bug)",
			"line one\nline two",
			`line one\nline two`,
		},
		{"double quote", `Say "hi"`, `Say \"hi\"`},
		{"backslash", `C:\path\to\file`, `C:\\path\\to\\file`},
		{"carriage return", "a\rb", `a\rb`},
		{"tab", "a\tb", `a\tb`},
		{
			"backslash before quote must not merge",
			`a\"b`,
			`a\\\"b`,
		},
		{
			"combination",
			"line one\nline two with a \"quoted\" word and a back\\slash",
			`line one\nline two with a \"quoted\" word and a back\\slash`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EscapeGraphQLString(tc.input)
			if got != tc.want {
				t.Errorf("EscapeGraphQLString(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// TestEscapeGraphQLString_ProducesValidStringLiteral guards against
// regressions to the escaping order/logic by asserting the invariant that
// actually matters: once escaped and dropped into a "..." GraphQL string
// literal, the result contains no raw control character that would
// terminate the literal early, and unescaping it byte-for-byte recovers
// the original input.
func TestEscapeGraphQLString_ProducesValidStringLiteral(t *testing.T) {
	inputs := []string{
		"",
		"plain",
		"line one\nline two",
		`quote " and backslash \`,
		"tabs\tand\rcarriage returns",
		`\n literally typed backslash-n, not a real newline`,
	}

	for _, input := range inputs {
		escaped := EscapeGraphQLString(input)

		for _, raw := range []byte{'\n', '\r', '\t'} {
			if strings.IndexByte(escaped, raw) != -1 {
				t.Errorf("EscapeGraphQLString(%q) = %q still contains raw %q", input, escaped, raw)
			}
		}

		literal := fmt.Sprintf(`"%s"`, escaped)
		unescaped, err := unescapeGraphQLStringLiteral(literal)
		if err != nil {
			t.Fatalf("EscapeGraphQLString(%q) produced an invalid string literal %q: %v", input, literal, err)
		}
		if unescaped != input {
			t.Errorf("round-trip mismatch: input %q, literal %q, unescaped back to %q", input, literal, unescaped)
		}
	}
}

// unescapeGraphQLStringLiteral is a minimal decoder for a double-quoted
// GraphQL string literal, used only to verify EscapeGraphQLString's output
// round-trips correctly. It is not a general-purpose GraphQL parser.
func unescapeGraphQLStringLiteral(literal string) (string, error) {
	if len(literal) < 2 || literal[0] != '"' || literal[len(literal)-1] != '"' {
		return "", fmt.Errorf("not a quoted string literal: %q", literal)
	}
	body := literal[1 : len(literal)-1]

	var out strings.Builder
	for i := 0; i < len(body); i++ {
		c := body[i]
		if c == '\n' || c == '\r' || c == '\t' {
			return "", fmt.Errorf("raw control character in literal body at index %d", i)
		}
		if c != '\\' {
			out.WriteByte(c)
			continue
		}
		i++
		if i >= len(body) {
			return "", fmt.Errorf("dangling escape at end of literal")
		}
		switch body[i] {
		case '\\':
			out.WriteByte('\\')
		case '"':
			out.WriteByte('"')
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case 't':
			out.WriteByte('\t')
		default:
			return "", fmt.Errorf("unknown escape sequence \\%c at index %d", body[i], i)
		}
	}
	return out.String(), nil
}
