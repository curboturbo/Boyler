package boilerfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fullBoilerfile = `# Build stage
FROM ubuntu:24.04 AS builder

ARG BUILD_VERSION=dev

LABEL maintainer="team@boyler.io" version="1.0"

WORKDIR /app

COPY . .

RUN apt-get update && \
    apt-get install -y make

RUN make build

EXPOSE 8080

ENV APP_ENV=production
ENV LOG_LEVEL debug

CMD ["./app", "--port", "8080"]
`

func TestIntegrationFullBoilerfile(t *testing.T) {
	bf, err := Parse(strings.NewReader(fullBoilerfile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantOrder := []string{"FROM", "ARG", "LABEL", "WORKDIR", "COPY", "RUN", "RUN", "EXPOSE", "ENV", "ENV", "CMD"}
	if len(bf.Instructions) != len(wantOrder) {
		t.Fatalf("expected %d instructions, got %d", len(wantOrder), len(bf.Instructions))
	}
	for i, kw := range wantOrder {
		if bf.Instructions[i].Keyword() != kw {
			t.Errorf("instruction[%d]: want %s, got %s", i, kw, bf.Instructions[i].Keyword())
		}
	}

	from := bf.Instructions[0].(*FromInstruction)
	if from.Image != "ubuntu" || from.Tag != "24.04" || from.Alias != "builder" {
		t.Errorf("FROM: %+v", from)
	}

	arg := bf.Instructions[1].(*ArgInstruction)
	if arg.Name != "BUILD_VERSION" || arg.Default != "dev" {
		t.Errorf("ARG: %+v", arg)
	}

	label := bf.Instructions[2].(*LabelInstruction)
	if label.Labels["maintainer"] != "team@boyler.io" {
		t.Errorf("LABEL: %v", label.Labels)
	}

	run1 := bf.Instructions[5].(*RunInstruction)
	if !run1.Shell {
		t.Error("RUN apt-get should be shell form")
	}

	expose := bf.Instructions[7].(*ExposeInstruction)
	if expose.Port != "8080" || expose.Protocol != "tcp" {
		t.Errorf("EXPOSE: %+v", expose)
	}

	cmd := bf.Instructions[10].(*CmdInstruction)
	if cmd.Shell || cmd.Command != "./app" || len(cmd.Args) != 2 {
		t.Errorf("CMD: %+v", cmd)
	}
}

func TestIntegrationCommentOnly(t *testing.T) {
	src := "# just a comment\n# another comment"
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error: no FROM instruction")
	}
}

func TestIntegrationLineContinuation(t *testing.T) {
	src := `FROM ubuntu
RUN apt-get install \
    curl \
    vim`
	bf, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	r := bf.Instructions[1].(*RunInstruction)
	if !strings.Contains(r.Command, "curl") || !strings.Contains(r.Command, "vim") {
		t.Errorf("continuation not merged: %q", r.Command)
	}
}

func TestIntegrationParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Boilerfile")
	if err := os.WriteFile(path, []byte(fullBoilerfile), 0644); err != nil {
		t.Fatal(err)
	}
	bf, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(bf.Instructions) == 0 {
		t.Error("expected instructions")
	}
}

func TestIntegrationParseFileMissing(t *testing.T) {
	_, err := ParseFile("/no/such/Boilerfile")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestIntegrationMinimal(t *testing.T) {
	bf, err := Parse(strings.NewReader("FROM alpine"))
	if err != nil {
		t.Fatal(err)
	}
	if len(bf.Instructions) != 1 {
		t.Errorf("expected 1 instruction, got %d", len(bf.Instructions))
	}
}

func TestIntegrationEntrypointAndCmd(t *testing.T) {
	src := `FROM ubuntu
ENTRYPOINT ["/bin/sh","-c"]
CMD ["echo hello"]`
	bf, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	ep := bf.Instructions[1].(*EntrypointInstruction)
	if ep.Shell || ep.Command != "/bin/sh" {
		t.Errorf("ENTRYPOINT: %+v", ep)
	}
}

func TestIntegrationErrorLineNumber(t *testing.T) {
	src := "FROM ubuntu\n\n\nBADINSTR foo"
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	if pe.Line != 4 {
		t.Errorf("expected error on line 4, got line %d", pe.Line)
	}
}
