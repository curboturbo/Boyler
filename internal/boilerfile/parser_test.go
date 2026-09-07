package boilerfile

import (
	"strings"
	"testing"
)

func parse(t *testing.T, src string) *Boilerfile {
	t.Helper()
	bf, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	return bf
}

func parseErr(t *testing.T, src string) error {
	t.Helper()
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	return err
}

// FROM

func TestParseFromBasic(t *testing.T) {
	bf := parse(t, "FROM ubuntu")
	f := bf.Instructions[0].(*FromInstruction)
	if f.Image != "ubuntu" || f.Tag != "latest" || f.Alias != "" {
		t.Errorf("unexpected: %+v", f)
	}
}

func TestParseFromWithTag(t *testing.T) {
	bf := parse(t, "FROM ubuntu:24.04")
	f := bf.Instructions[0].(*FromInstruction)
	if f.Image != "ubuntu" || f.Tag != "24.04" {
		t.Errorf("unexpected: %+v", f)
	}
}

func TestParseFromWithAlias(t *testing.T) {
	bf := parse(t, "FROM ubuntu:24.04 AS builder")
	f := bf.Instructions[0].(*FromInstruction)
	if f.Alias != "builder" {
		t.Errorf("alias not parsed: %+v", f)
	}
}

func TestParseFromMissing(t *testing.T) {
	parseErr(t, "WORKDIR /app")
}

func TestParseFromEmpty(t *testing.T) {
	parseErr(t, "FROM")
}

func TestParseEmptyFile(t *testing.T) {
	parseErr(t, "")
}

func TestParseUnknownInstruction(t *testing.T) {
	parseErr(t, "FROM ubuntu\nINVALID foo")
}

// WORKDIR

func TestParseWorkdir(t *testing.T) {
	bf := parse(t, "FROM ubuntu\nWORKDIR /app")
	w := bf.Instructions[1].(*WorkdirInstruction)
	if w.Path != "/app" {
		t.Errorf("unexpected path: %q", w.Path)
	}
}

func TestParseWorkdirEmpty(t *testing.T) {
	parseErr(t, "FROM ubuntu\nWORKDIR")
}

// COPY

func TestParseCopySingle(t *testing.T) {
	bf := parse(t, "FROM ubuntu\nCOPY main.go /app/")
	c := bf.Instructions[1].(*CopyInstruction)
	if len(c.Srcs) != 1 || c.Srcs[0] != "main.go" || c.Dst != "/app/" {
		t.Errorf("unexpected: %+v", c)
	}
}

func TestParseCopyMultipleSrc(t *testing.T) {
	bf := parse(t, "FROM ubuntu\nCOPY a.go b.go /app/")
	c := bf.Instructions[1].(*CopyInstruction)
	if len(c.Srcs) != 2 || c.Dst != "/app/" {
		t.Errorf("unexpected: %+v", c)
	}
}

func TestParseCopyMissingDst(t *testing.T) {
	parseErr(t, "FROM ubuntu\nCOPY src")
}

// RUN

func TestParseRunShellForm(t *testing.T) {
	bf := parse(t, "FROM ubuntu\nRUN make build")
	r := bf.Instructions[1].(*RunInstruction)
	if !r.Shell || r.Command != "make build" {
		t.Errorf("unexpected: %+v", r)
	}
}

func TestParseRunExecForm(t *testing.T) {
	bf := parse(t, `FROM ubuntu\nRUN ["/bin/sh","-c","make"]`)
	// Note: the \n is a literal; use multi-line string instead.
	bf = parse(t, "FROM ubuntu\nRUN [\"/bin/sh\",\"-c\",\"make\"]")
	r := bf.Instructions[1].(*RunInstruction)
	if r.Shell || r.Command != "/bin/sh" {
		t.Errorf("unexpected: %+v", r)
	}
	if len(r.Args) != 2 || r.Args[0] != "-c" {
		t.Errorf("unexpected args: %v", r.Args)
	}
}

func TestParseRunInvalidExecJSON(t *testing.T) {
	parseErr(t, "FROM ubuntu\nRUN [not valid json]")
}

func TestParseRunEmpty(t *testing.T) {
	parseErr(t, "FROM ubuntu\nRUN")
}

// CMD

func TestParseCmdExecForm(t *testing.T) {
	bf := parse(t, `FROM ubuntu
CMD ["./app","--port","8080"]`)
	c := bf.Instructions[1].(*CmdInstruction)
	if c.Shell || c.Command != "./app" || len(c.Args) != 2 {
		t.Errorf("unexpected: %+v", c)
	}
}

func TestParseCmdShellForm(t *testing.T) {
	bf := parse(t, "FROM ubuntu\nCMD ./app --port 8080")
	c := bf.Instructions[1].(*CmdInstruction)
	if !c.Shell || c.Command != "./app --port 8080" {
		t.Errorf("unexpected: %+v", c)
	}
}

func TestParseCmdEmptyExecForm(t *testing.T) {
	parseErr(t, "FROM ubuntu\nCMD []")
}

// ENTRYPOINT

func TestParseEntrypointExecForm(t *testing.T) {
	bf := parse(t, `FROM ubuntu
ENTRYPOINT ["/bin/sh","-c"]`)
	e := bf.Instructions[1].(*EntrypointInstruction)
	if e.Shell || e.Command != "/bin/sh" {
		t.Errorf("unexpected: %+v", e)
	}
}

// ENV

func TestParseEnvKeyValue(t *testing.T) {
	bf := parse(t, "FROM ubuntu\nENV FOO=bar")
	e := bf.Instructions[1].(*EnvInstruction)
	if e.Key != "FOO" || e.Value != "bar" {
		t.Errorf("unexpected: %+v", e)
	}
}

func TestParseEnvSpaceSeparated(t *testing.T) {
	bf := parse(t, "FROM ubuntu\nENV FOO bar")
	e := bf.Instructions[1].(*EnvInstruction)
	if e.Key != "FOO" || e.Value != "bar" {
		t.Errorf("unexpected: %+v", e)
	}
}

func TestParseEnvQuotedValue(t *testing.T) {
	bf := parse(t, `FROM ubuntu
ENV MESSAGE="hello world"`)
	e := bf.Instructions[1].(*EnvInstruction)
	if e.Value != "hello world" {
		t.Errorf("quotes not stripped: %q", e.Value)
	}
}

func TestParseEnvEmpty(t *testing.T) {
	parseErr(t, "FROM ubuntu\nENV")
}

// EXPOSE

func TestParseExposeDefault(t *testing.T) {
	bf := parse(t, "FROM ubuntu\nEXPOSE 8080")
	e := bf.Instructions[1].(*ExposeInstruction)
	if e.Port != "8080" || e.Protocol != "tcp" {
		t.Errorf("unexpected: %+v", e)
	}
}

func TestParseExposeUDP(t *testing.T) {
	bf := parse(t, "FROM ubuntu\nEXPOSE 5353/udp")
	e := bf.Instructions[1].(*ExposeInstruction)
	if e.Port != "5353" || e.Protocol != "udp" {
		t.Errorf("unexpected: %+v", e)
	}
}

// LABEL

func TestParseLabel(t *testing.T) {
	bf := parse(t, `FROM ubuntu
LABEL maintainer="kunal" version="1.0"`)
	l := bf.Instructions[1].(*LabelInstruction)
	if l.Labels["maintainer"] != "kunal" || l.Labels["version"] != "1.0" {
		t.Errorf("unexpected labels: %v", l.Labels)
	}
}

func TestParseLabelMissingEquals(t *testing.T) {
	parseErr(t, "FROM ubuntu\nLABEL noequals")
}

// ARG

func TestParseArgNoDefault(t *testing.T) {
	bf := parse(t, "FROM ubuntu\nARG BUILD_VERSION")
	a := bf.Instructions[1].(*ArgInstruction)
	if a.Name != "BUILD_VERSION" || a.HasDefault {
		t.Errorf("unexpected: %+v", a)
	}
}

func TestParseArgWithDefault(t *testing.T) {
	bf := parse(t, "FROM ubuntu\nARG BUILD_VERSION=dev")
	a := bf.Instructions[1].(*ArgInstruction)
	if a.Name != "BUILD_VERSION" || !a.HasDefault || a.Default != "dev" {
		t.Errorf("unexpected: %+v", a)
	}
}

func TestParseLineNumbers(t *testing.T) {
	src := "# comment\nFROM ubuntu\nWORKDIR /app"
	bf := parse(t, src)
	if bf.Instructions[0].LineNumber() != 2 {
		t.Errorf("FROM should be line 2, got %d", bf.Instructions[0].LineNumber())
	}
	if bf.Instructions[1].LineNumber() != 3 {
		t.Errorf("WORKDIR should be line 3, got %d", bf.Instructions[1].LineNumber())
	}
}

func TestParseKeywords(t *testing.T) {
	for _, kw := range []string{"FROM ubuntu", "WORKDIR /x", "COPY a b", "RUN x", "CMD x",
		"ENTRYPOINT x", "ENV K=V", "EXPOSE 80", "LABEL k=v", "ARG N"} {
		src := "FROM ubuntu\n" + kw
		if kw == "FROM ubuntu" {
			src = kw
		}
		if _, err := Parse(strings.NewReader(src)); err != nil {
			t.Errorf("failed to parse %q: %v", kw, err)
		}
	}
}
