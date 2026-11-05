package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const usage = `tfv layers *.tfvars from parent directories onto terraform commands.

  eval "$(tfv init bash)"   wrap terraform in this shell (bash or zsh)
  tfv [status]              show the files terraform would get
  tfv -args                 print them as arguments
  tfv exec ARGS...          run terraform with them
  tfv -deactivate           pause the wrapper in this shell
  tfv -activate             resume it

Flags for status and -args:
`

const hook = `export TFV_HOOK=1
terraform() { command tfv exec "$@"; }
tfv() {
  case "$1" in
    -deactivate|--deactivate) export TFV_DISABLED=1 ;;
    -activate|--activate) unset TFV_DISABLED ;;
    *) command tfv "$@" ;;
  esac
}
`

var varCommands = map[string]bool{
	"plan": true, "apply": true, "destroy": true,
	"refresh": true, "import": true, "console": true,
}

var applyBoolFlags = map[string]bool{
	"-auto-approve": true, "-compact-warnings": true, "-destroy": true, "-input": true,
	"-json": true, "-lock": true, "-no-color": true, "-refresh": true, "-refresh-only": true,
}

type search struct {
	ext      string
	maxDepth int
	stopAt   string
}

func defaultSearch() search {
	s := search{ext: ".tfvars", maxDepth: -1, stopAt: ".git"}
	if v, err := strconv.Atoi(os.Getenv("TFV_MAX_DEPTH")); err == nil {
		s.maxDepth = v
	}
	if v, ok := os.LookupEnv("TFV_STOP_AT"); ok {
		s.stopAt = v
	}
	return s
}

func find(start string, s search) (files []string, top string, unbounded bool, err error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil, "", false, err
	}
	dirs := []string{dir}
	found := s.stopAt == ""
	for depth := 0; s.maxDepth < 0 || depth < s.maxDepth; depth++ {
		if s.stopAt != "" {
			if _, err := os.Stat(filepath.Join(dir, s.stopAt)); err == nil {
				found = true
				break
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
		dirs = append(dirs, dir)
	}
	if !found && s.maxDepth < 0 {
		dirs = dirs[:1]
		unbounded = true
	}
	for i := len(dirs) - 1; i >= 0; i-- {
		entries, err := os.ReadDir(dirs[i])
		if err != nil {
			return nil, "", false, err
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), s.ext) {
				files = append(files, filepath.Join(dirs[i], e.Name()))
			}
		}
	}
	return files, dirs[len(dirs)-1], unbounded, nil
}

func warnUnbounded(w io.Writer, s search) {
	fmt.Fprintf(w, "tfv: no %s found above; using the start directory only (set -stop-at to change)\n", s.stopAt)
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func cliArgs(files []string) string {
	args := make([]string, len(files))
	for i, f := range files {
		args[i] = "-var-file=" + shellQuote(f)
	}
	return strings.Join(args, " ")
}

func parseTerraform(args []string) (chdir, cmd string, rest []string) {
	for i, a := range args {
		if v, ok := strings.CutPrefix(a, "-chdir="); ok {
			chdir = v
			continue
		}
		if strings.HasPrefix(a, "-") {
			continue
		}
		return chdir, a, args[i+1:]
	}
	return chdir, "", nil
}

func appliesSavedPlan(args []string) bool {
	if len(args) == 0 || strings.HasPrefix(args[len(args)-1], "-") {
		return false
	}
	if len(args) == 1 {
		return true
	}
	prev := args[len(args)-2]
	return !strings.HasPrefix(prev, "-") || strings.Contains(prev, "=") || applyBoolFlags[prev]
}

func runExec(args []string) int {
	chdir, cmd, rest := parseTerraform(args)
	env := os.Environ()

	if varCommands[cmd] && os.Getenv("TFV_DISABLED") == "" && !(cmd == "apply" && appliesSavedPlan(rest)) {
		start := "."
		if chdir != "" {
			start = chdir
		}
		s := defaultSearch()
		files, _, unbounded, err := find(start, s)
		if err != nil {
			fmt.Fprintln(os.Stderr, "tfv:", err)
			return 1
		}
		if unbounded {
			warnUnbounded(os.Stderr, s)
		}
		if len(files) > 0 {
			fmt.Fprintf(os.Stderr, "tfv: %d var file(s)\n", len(files))
			key := "TF_CLI_ARGS_" + cmd
			value := cliArgs(files)
			if existing := os.Getenv(key); existing != "" {
				value += " " + existing
			}
			env = append(env, key+"="+value)
		}
	}

	c := exec.Command("terraform", args...)
	c.Env, c.Stdin, c.Stdout, c.Stderr = env, os.Stdin, os.Stdout, os.Stderr
	if err := c.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "tfv:", err)
		return 127
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range sigs {
			if sig == syscall.SIGTERM {
				c.Process.Signal(sig)
			}
		}
	}()

	if err := c.Wait(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() > 0 {
			return exit.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "tfv:", err)
		return 1
	}
	return 0
}

func runInit(args []string) int {
	if len(args) != 1 || (args[0] != "bash" && args[0] != "zsh") {
		fmt.Fprintln(os.Stderr, "usage: tfv init bash|zsh")
		return 2
	}
	fmt.Print(hook)
	return 0
}

func status(w io.Writer, start string, files []string, top string) {
	state := "active"
	switch {
	case os.Getenv("TFV_DISABLED") != "":
		state = "deactivated"
	case os.Getenv("TFV_HOOK") == "":
		state = "no shell hook; use tfv exec, or eval \"$(tfv init bash)\""
	}
	abs, _ := filepath.Abs(start)
	env, err := filepath.Rel(top, abs)
	if err != nil {
		env = abs
	}
	fmt.Fprintf(w, "state: %s\nroot:  %s\nenv:   %s\n", state, top, env)
	if len(files) == 0 {
		fmt.Fprintln(w, "files: none")
		return
	}
	fmt.Fprintln(w, "files, last wins:")
	for _, f := range files {
		fmt.Fprintln(w, " ", f)
	}
}

func run(args []string) int {
	if len(args) > 0 {
		switch args[0] {
		case "exec":
			return runExec(args[1:])
		case "init":
			return runInit(args[1:])
		case "status":
			args = args[1:]
		}
	}

	s := defaultSearch()
	fs := flag.NewFlagSet("tfv", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), usage)
		fs.PrintDefaults()
	}
	start := fs.String("d", ".", "start directory")
	fs.StringVar(&s.ext, "f", s.ext, "file extension to match")
	fs.IntVar(&s.maxDepth, "max-depth", s.maxDepth, "parent directories to climb, -1 for unlimited (env TFV_MAX_DEPTH)")
	fs.StringVar(&s.stopAt, "stop-at", s.stopAt, "stop at the directory containing this entry, empty to climb to / (env TFV_STOP_AT)")
	printArgs := fs.Bool("args", false, "print files as arguments")
	format := fs.String("format", "-var-file=%s", "printf format for each file with -args")
	deactivate := fs.Bool("deactivate", false, "pause the wrapper (needs the shell hook)")
	activate := fs.Bool("activate", false, "resume the wrapper (needs the shell hook)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if *deactivate || *activate {
		fmt.Fprintln(os.Stderr, `tfv: needs the shell hook: eval "$(tfv init bash)"`)
		return 1
	}
	if strings.Count(*format, "%s") != 1 {
		fmt.Fprintf(os.Stderr, "tfv: -format needs exactly one %%s\n")
		return 2
	}

	files, top, unbounded, err := find(*start, s)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tfv:", err)
		return 1
	}
	if unbounded {
		warnUnbounded(os.Stderr, s)
	}
	if !*printArgs {
		status(os.Stdout, *start, files, top)
		return 0
	}
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = fmt.Sprintf(*format, f)
	}
	fmt.Println(strings.Join(out, " "))
	return 0
}

func main() {
	os.Exit(run(os.Args[1:]))
}
