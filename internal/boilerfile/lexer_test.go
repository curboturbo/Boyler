package boilerfile

import (
	"strings"
	"testing"
)

func lex(t *testing.T, src string) []Token {
	t.Helper()
	tokens, err := Lex(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Lex error: %v", err)
	}
	return tokens
}

func TestLexEmptyInput(t *testing.T) {
	tokens := lex(t, "")
	if len(tokens) != 1 || tokens[0].Type != EOF {
		t.Errorf("expected only EOF, got %v", tokens)
	}
}

func TestLexBlankLinesSkipped(t *testing.T) {
	tokens := lex(t, "\n\n   \n")
	if len(tokens) != 1 || tokens[0].Type != EOF {
		t.Errorf("expected only EOF, got %v", tokens)
	}
}

func TestLexComment(t *testing.T) {
	tokens := lex(t, "# this is a comment")
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != COMMENT || tokens[0].Value != "this is a comment" {
		t.Errorf("unexpected comment token: %+v", tokens[0])
	}
}

func TestLexInstruction(t *testing.T) {
	tokens := lex(t, "FROM ubuntu:24.04")
	// INSTRUCTION + ARGS + EOF
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Type != INSTRUCTION || tokens[0].Value != "FROM" {
		t.Errorf("bad instruction token: %+v", tokens[0])
	}
	if tokens[1].Type != ARGS || tokens[1].Value != "ubuntu:24.04" {
		t.Errorf("bad args token: %+v", tokens[1])
	}
}

func TestLexKeywordUpperCased(t *testing.T) {
	tokens := lex(t, "from alpine")
	if tokens[0].Value != "FROM" {
		t.Errorf("keyword should be upper-cased, got %q", tokens[0].Value)
	}
}

func TestLexLineContinuation(t *testing.T) {
	src := "RUN apt-get install \\\n    curl vim"
	tokens := lex(t, src)
	// INSTRUCTION + ARGS + EOF
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[1].Value != "apt-get install curl vim" {
		t.Errorf("unexpected args after continuation: %q", tokens[1].Value)
	}
}

func TestLexLineNumbers(t *testing.T) {
	src := "# comment\nFROM ubuntu"
	tokens := lex(t, src)
	// COMMENT(1) + INSTRUCTION(2) + ARGS(2) + EOF
	if tokens[0].Line != 1 {
		t.Errorf("comment should be on line 1, got %d", tokens[0].Line)
	}
	if tokens[1].Line != 2 {
		t.Errorf("FROM should be on line 2, got %d", tokens[1].Line)
	}
}

func TestLexMultipleInstructions(t *testing.T) {
	src := "FROM ubuntu\nWORKDIR /app\nRUN make"
	tokens := lex(t, src)
	// 3 × (INSTRUCTION + ARGS) + EOF = 7
	if len(tokens) != 7 {
		t.Fatalf("expected 7 tokens, got %d", len(tokens))
	}
}

func TestLexNoArgsToken(t *testing.T) {
	// An instruction with no arguments still emits an empty ARGS token.
	tokens := lex(t, "CMD")
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[1].Type != ARGS || tokens[1].Value != "" {
		t.Errorf("expected empty ARGS, got %+v", tokens[1])
	}
}
