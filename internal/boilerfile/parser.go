package boilerfile

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// supported is the complete set of recognised Boilerfile keywords.
var supported = map[string]bool{
	"FROM":       true,
	"WORKDIR":    true,
	"COPY":       true,
	"RUN":        true,
	"CMD":        true,
	"ENTRYPOINT": true,
	"ENV":        true,
	"EXPOSE":     true,
	"LABEL":      true,
	"ARG":        true,
}

// ParseFile opens the file at path and calls Parse.
func ParseFile(path string) (*Boilerfile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// Parse tokenizes r and builds the AST. It returns the first error encountered.
func Parse(r io.Reader) (*Boilerfile, error) {
	tokens, err := Lex(r)
	if err != nil {
		return nil, err
	}
	return build(tokens)
}

func build(tokens []Token) (*Boilerfile, error) {
	bf := &Boilerfile{}
	fromSeen := false

	for i := 0; i < len(tokens); {
		tok := tokens[i]

		switch tok.Type {
		case EOF:
			// Reached end of token stream.
			if !fromSeen {
				return nil, &ParseError{Line: 1, Message: "FROM is required"}
			}
			return bf, nil

		case COMMENT:
			i++
			continue

		case INSTRUCTION:
			// Every INSTRUCTION is immediately followed by an ARGS token.
			if i+1 >= len(tokens) || tokens[i+1].Type != ARGS {
				return nil, &ParseError{Line: tok.Line, Message: "internal: missing ARGS token after " + tok.Value}
			}
			argsTok := tokens[i+1]
			i += 2

			if !supported[tok.Value] {
				return nil, &ParseError{Line: tok.Line, Message: fmt.Sprintf("unknown instruction %q", tok.Value)}
			}

			if !fromSeen && tok.Value != "FROM" {
				return nil, &ParseError{Line: tok.Line, Message: "FROM must be the first instruction"}
			}

			node, err := parseInstruction(tok, argsTok)
			if err != nil {
				return nil, err
			}
			if tok.Value == "FROM" {
				fromSeen = true
			}
			bf.Instructions = append(bf.Instructions, node)

		default:
			i++
		}
	}

	if !fromSeen {
		return nil, &ParseError{Line: 1, Message: "FROM is required"}
	}
	return bf, nil
}

func parseInstruction(kw, args Token) (Instruction, error) {
	raw := strings.TrimSpace(args.Value)
	line := kw.Line

	switch kw.Value {
	case "FROM":
		return parseFrom(line, raw)
	case "WORKDIR":
		return parseWorkdir(line, raw)
	case "COPY":
		return parseCopy(line, raw)
	case "RUN":
		return parseExecOrShell(line, raw, func(shell bool, cmd string, a []string) Instruction {
			return &RunInstruction{base: base{line}, Shell: shell, Command: cmd, Args: a}
		})
	case "CMD":
		return parseExecOrShell(line, raw, func(shell bool, cmd string, a []string) Instruction {
			return &CmdInstruction{base: base{line}, Shell: shell, Command: cmd, Args: a}
		})
	case "ENTRYPOINT":
		return parseExecOrShell(line, raw, func(shell bool, cmd string, a []string) Instruction {
			return &EntrypointInstruction{base: base{line}, Shell: shell, Command: cmd, Args: a}
		})
	case "ENV":
		return parseEnv(line, raw)
	case "EXPOSE":
		return parseExpose(line, raw)
	case "LABEL":
		return parseLabel(line, raw)
	case "ARG":
		return parseArg(line, raw)
	default:
		return nil, &ParseError{Line: line, Message: "unknown instruction " + kw.Value}
	}
}

// parseFrom handles: FROM <image>[:<tag>] [AS <alias>]
func parseFrom(line int, raw string) (*FromInstruction, error) {
	if raw == "" {
		return nil, &ParseError{Line: line, Message: "FROM requires an image argument"}
	}
	parts := strings.Fields(raw)
	instr := &FromInstruction{base: base{line}, Tag: "latest"}

	imageTag := parts[0]
	if idx := strings.LastIndex(imageTag, ":"); idx != -1 {
		instr.Image = imageTag[:idx]
		instr.Tag = imageTag[idx+1:]
		if instr.Image == "" || instr.Tag == "" {
			return nil, &ParseError{Line: line, Message: "invalid image reference: " + imageTag}
		}
	} else {
		instr.Image = imageTag
	}

	if len(parts) >= 3 {
		if !strings.EqualFold(parts[1], "AS") {
			return nil, &ParseError{Line: line, Message: "expected AS after image reference, got " + parts[1]}
		}
		instr.Alias = parts[2]
	}
	return instr, nil
}

// parseWorkdir handles: WORKDIR <path>
func parseWorkdir(line int, raw string) (*WorkdirInstruction, error) {
	if raw == "" {
		return nil, &ParseError{Line: line, Message: "WORKDIR requires a path"}
	}
	return &WorkdirInstruction{base: base{line}, Path: raw}, nil
}

// parseCopy handles: COPY <src>... <dst>
func parseCopy(line int, raw string) (*CopyInstruction, error) {
	fields := strings.Fields(raw)
	if len(fields) < 2 {
		return nil, &ParseError{Line: line, Message: "COPY requires at least one source and a destination"}
	}
	return &CopyInstruction{
		base: base{line},
		Srcs: fields[:len(fields)-1],
		Dst:  fields[len(fields)-1],
	}, nil
}

type nodeBuilder func(shell bool, cmd string, args []string) Instruction

// parseExecOrShell handles both exec form ["cmd","arg"] and shell form.
func parseExecOrShell(line int, raw string, build nodeBuilder) (Instruction, error) {
	if raw == "" {
		return nil, &ParseError{Line: line, Message: "instruction requires arguments"}
	}
	if strings.HasPrefix(raw, "[") {
		var argv []string
		if err := json.Unmarshal([]byte(raw), &argv); err != nil {
			return nil, &ParseError{Line: line, Message: "invalid exec-form JSON: " + err.Error()}
		}
		if len(argv) == 0 {
			return nil, &ParseError{Line: line, Message: "exec-form must have at least one element"}
		}
		return build(false, argv[0], argv[1:]), nil
	}
	return build(true, raw, nil), nil
}

// parseEnv handles: ENV <key>=<value>  and  ENV <key> <value>
func parseEnv(line int, raw string) (*EnvInstruction, error) {
	if raw == "" {
		return nil, &ParseError{Line: line, Message: "ENV requires a key and value"}
	}
	if idx := strings.IndexByte(raw, '='); idx != -1 {
		key := strings.TrimSpace(raw[:idx])
		if key == "" {
			return nil, &ParseError{Line: line, Message: "ENV key cannot be empty"}
		}
		return &EnvInstruction{base: base{line}, Key: key, Value: unquote(strings.TrimSpace(raw[idx+1:]))}, nil
	}
	key, val, ok := strings.Cut(raw, " ")
	if !ok {
		return nil, &ParseError{Line: line, Message: "ENV requires a value"}
	}
	return &EnvInstruction{base: base{line}, Key: key, Value: strings.TrimSpace(val)}, nil
}

// parseExpose handles: EXPOSE <port>[/<proto>]
func parseExpose(line int, raw string) (*ExposeInstruction, error) {
	if raw == "" {
		return nil, &ParseError{Line: line, Message: "EXPOSE requires a port number"}
	}
	port, proto, ok := strings.Cut(raw, "/")
	if !ok {
		proto = "tcp"
	}
	return &ExposeInstruction{base: base{line}, Port: port, Protocol: strings.ToLower(proto)}, nil
}

// parseLabel handles: LABEL <key>=<value> [<key>=<value> ...]
func parseLabel(line int, raw string) (*LabelInstruction, error) {
	if raw == "" {
		return nil, &ParseError{Line: line, Message: "LABEL requires at least one key=value pair"}
	}
	labels, err := parseKeyValuePairs(line, raw)
	if err != nil {
		return nil, err
	}
	return &LabelInstruction{base: base{line}, Labels: labels}, nil
}

// parseArg handles: ARG <name>[=<default>]
func parseArg(line int, raw string) (*ArgInstruction, error) {
	if raw == "" {
		return nil, &ParseError{Line: line, Message: "ARG requires a name"}
	}
	name, def, hasDefault := strings.Cut(raw, "=")
	return &ArgInstruction{
		base:       base{line},
		Name:       strings.TrimSpace(name),
		Default:    unquote(strings.TrimSpace(def)),
		HasDefault: hasDefault,
	}, nil
}

// parseKeyValuePairs parses space-separated key=value pairs, respecting quotes.
func parseKeyValuePairs(line int, raw string) (map[string]string, error) {
	result := make(map[string]string)
	for _, field := range splitQuoted(raw) {
		key, val, ok := strings.Cut(field, "=")
		if !ok {
			return nil, &ParseError{Line: line, Message: fmt.Sprintf("expected key=value, got %q", field)}
		}
		result[key] = unquote(val)
	}
	return result, nil
}

// splitQuoted splits s on unquoted spaces.
func splitQuoted(s string) []string {
	var out []string
	var cur strings.Builder
	inQ := false
	qch := byte(0)

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case inQ && ch == qch:
			inQ = false
			cur.WriteByte(ch)
		case !inQ && (ch == '"' || ch == '\''):
			inQ = true
			qch = ch
			cur.WriteByte(ch)
		case !inQ && ch == ' ':
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(ch)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

// unquote strips surrounding double or single quotes.
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
