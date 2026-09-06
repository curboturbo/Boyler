package boilerfile

// TokenType identifies the kind of a scanned token.
type TokenType int

const (
	EOF TokenType = iota
	COMMENT     // # ...
	INSTRUCTION // FROM, RUN, COPY, ...
	ARGS        // everything after the instruction keyword on the same logical line
	ILLEGAL     // unrecognised content (never returned by the lexer; reserved for future use)
)

func (t TokenType) String() string {
	switch t {
	case EOF:
		return "EOF"
	case COMMENT:
		return "COMMENT"
	case INSTRUCTION:
		return "INSTRUCTION"
	case ARGS:
		return "ARGS"
	default:
		return "ILLEGAL"
	}
}

// Token is a single lexical unit produced by the Lexer.
type Token struct {
	Type  TokenType
	Value string
	Line  int // 1-based source line where this token starts
}
