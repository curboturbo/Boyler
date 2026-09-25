package boilerfile

import (
	"bufio"
	"io"
	"strings"
)

// Lex reads a Boilerfile from r and returns a flat token stream ending with EOF.
// It handles blank lines (skipped), comments (#), and backslash line continuations.
// Instruction keywords are emitted verbatim in upper-case; validation of whether
// they are recognised is left to the parser.
func Lex(r io.Reader) ([]Token, error) {
	scanner := bufio.NewScanner(r)
	var tokens []Token
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		startLine := lineNum

		// Merge continuation lines: a logical line ends when the physical line
		// does NOT end with a backslash (after trimming trailing whitespace).
		for {
			physical := strings.TrimRight(line, " \t")
			if !strings.HasSuffix(physical, "\\") {
				break
			}
			// Remove the trailing backslash.
			line = physical[:len(physical)-1]
			if !scanner.Scan() {
				break
			}
			lineNum++
			line += strings.TrimLeft(scanner.Text(), " \t")
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			tokens = append(tokens,
				Token{Type: COMMENT, Value: strings.TrimSpace(trimmed[1:]), Line: startLine},
			)
			continue
		}

		// Split into keyword and the rest. SplitN with n=2 so the rest is untouched.
		keyword, rest, _ := strings.Cut(trimmed, " ")
		tokens = append(tokens,
			Token{Type: INSTRUCTION, Value: strings.ToUpper(keyword), Line: startLine},
			Token{Type: ARGS, Value: strings.TrimSpace(rest), Line: startLine},
		)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	tokens = append(tokens, Token{Type: EOF, Line: lineNum + 1})
	return tokens, nil
}
