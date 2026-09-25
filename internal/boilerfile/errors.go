package boilerfile

import "fmt"

// ParseError is a syntax error found while lexing or parsing a Boilerfile.
type ParseError struct {
	Line    int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Boilerfile:%d: %s", e.Line, e.Message)
}
