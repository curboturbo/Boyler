package boilerfile

// Boilerfile is the root AST node produced by Parse.
type Boilerfile struct {
	Instructions []Instruction
}

// Instruction is implemented by every AST node.
type Instruction interface {
	Keyword()    string
	LineNumber() int
}

type base struct{ line int }

func (b base) LineNumber() int { return b.line }

// FromInstruction: FROM <image>[:<tag>] [AS <alias>]
type FromInstruction struct {
	base
	Image string
	Tag   string // defaults to "latest"
	Alias string // optional AS clause
}

func (f *FromInstruction) Keyword() string { return "FROM" }

// WorkdirInstruction: WORKDIR <path>
type WorkdirInstruction struct {
	base
	Path string
}

func (w *WorkdirInstruction) Keyword() string { return "WORKDIR" }

// CopyInstruction: COPY <src>... <dst>
type CopyInstruction struct {
	base
	Srcs []string
	Dst  string
}

func (c *CopyInstruction) Keyword() string { return "COPY" }

// RunInstruction: RUN <shell-command>  or  RUN ["cmd","arg",...]
type RunInstruction struct {
	base
	Shell   bool     // true → shell form
	Command string   // shell form command string
	Args    []string // exec form: Args[0] is the executable
}

func (r *RunInstruction) Keyword() string { return "RUN" }

// CmdInstruction: CMD <shell-command>  or  CMD ["cmd","arg",...]
type CmdInstruction struct {
	base
	Shell   bool
	Command string
	Args    []string
}

func (c *CmdInstruction) Keyword() string { return "CMD" }

// EntrypointInstruction: ENTRYPOINT <shell-command>  or  ENTRYPOINT ["cmd","arg",...]
type EntrypointInstruction struct {
	base
	Shell   bool
	Command string
	Args    []string
}

func (e *EntrypointInstruction) Keyword() string { return "ENTRYPOINT" }

// EnvInstruction: ENV <key>=<value>  or  ENV <key> <value>
type EnvInstruction struct {
	base
	Key   string
	Value string
}

func (e *EnvInstruction) Keyword() string { return "ENV" }

// ExposeInstruction: EXPOSE <port>[/<protocol>]
type ExposeInstruction struct {
	base
	Port     string
	Protocol string // "tcp" (default) or "udp"
}

func (e *ExposeInstruction) Keyword() string { return "EXPOSE" }

// LabelInstruction: LABEL <key>=<value> [<key>=<value> ...]
type LabelInstruction struct {
	base
	Labels map[string]string
}

func (l *LabelInstruction) Keyword() string { return "LABEL" }

// ArgInstruction: ARG <name>[=<default>]
type ArgInstruction struct {
	base
	Name       string
	Default    string
	HasDefault bool
}

func (a *ArgInstruction) Keyword() string { return "ARG" }
