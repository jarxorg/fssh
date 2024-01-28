package fssh

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"path"
	"path/filepath"

	"github.com/chzyer/readline"
	"github.com/jarxorg/wfs"
)

// ShellName is "fssh".
const ShellName = "fssh"

// ErrExit represents an exit error. If this error is detected then the shell will terminate.
var ErrExit = errors.New("exit")

// Shell reads stdin, interprets lines, and executes commands.
type Shell struct {
	rl *readline.Instance

	Stdout        io.Writer
	Stderr        io.Writer
	FS            wfs.WriteFileFS
	Protocol      string
	Host          string
	Dir           string
	PrefixMatcher PrefixMatcher
}

// NewShell creates a new Shell.
func NewShell(dirUrl string) (*Shell, error) {
	homeDir, err := osUserHomeDir()
	if err != nil {
		return nil, err
	}

	fsys, protocol, host, dir, err := NewFS(dirUrl)
	if err != nil {
		return nil, err
	}
	sh := &Shell{
		FS:            fsys,
		Protocol:      protocol,
		Host:          host,
		Dir:           dir,
		PrefixMatcher: &GlobPrefixMatcher{},
	}
	rl, err := readline.NewEx(&readline.Config{
		HistoryFile:       filepath.Join(homeDir, fmt.Sprintf(".%s_history", ShellName)),
		AutoComplete:      newReadlineAutoCompleter(sh),
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		return nil, err
	}
	sh.rl = rl
	sh.Stdout = rl.Stdout()
	sh.Stderr = rl.Stderr()
	sh.UpdatePrompt()
	return sh, nil
}

// DirWithProtocol returns the current directory held by the shell.
func (sh *Shell) DirWithProtocol() string {
	return sh.Protocol + path.Join(sh.Host, sh.Dir)
}

// UpdatePrompt updates the command line prompt.
func (sh *Shell) UpdatePrompt() {
	sh.PrefixMatcher.Reset()
	sh.rl.SetPrompt("\033[36m" + sh.DirWithProtocol() + ">\033[0m ")
}

// Close closes the shell.
func (sh *Shell) Close() error {
	return sh.rl.Close()
}

// Run runs the shell.
func (sh *Shell) Run() error {
	sh.rl.CaptureExitSignal()
	log.SetOutput(sh.Stderr)
	for {
		line, err := sh.rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		if err := sh.ExecCommand(ParseArgs(line)); err != nil {
			if errors.Is(err, ErrExit) {
				return nil
			}
			fmt.Fprintf(sh.Stderr, "%s: %v\n", ShellName, err)
		}
	}
	return nil
}

// ExecCommand executes a command.
func (sh *Shell) ExecCommand(args []string) error {
	if len(args) == 0 {
		return nil
	}
	cmd := AquireCommand(args[0])
	if cmd == nil {
		return fmt.Errorf("command not found: %s", args[0])
	}
	defer ReleaseCommand(cmd)

	if err := cmd.FlagSet().Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			cmd.Usage(sh.Stdout)
			return nil
		}
		return err
	}
	return cmd.Exec(sh)
}

// Usage prints the usage to the specified writer..
func (sh *Shell) Usage(w io.Writer) {
	w.Write([]byte("Commands:\n"))
	for _, name := range SortedCommandNames() {
		w.Write([]byte("  "))
		w.Write([]byte(name))
		w.Write([]byte("\n"))
	}
}

// SubFS returns the FS and related path. If the dirUrl has protocol then this creates a new FS.
func (sh *Shell) SubFS(filenameUrl string) (FS, string, error) {
	if IsCurrentPath(filenameUrl) {
		return sh.FS, path.Join(sh.Dir, filenameUrl), nil
	}
	fsys, _, _, filename, err := NewFS(filenameUrl)
	if err != nil {
		return nil, "", err
	}
	return fsys, filename, nil
}

// SubFS returns the FS and related path. If the dirUrl has protocol then this creates a new FS.
func (sh *Shell) SubDirFS(dirUrl string) (FS, string, error) {
	if IsCurrentPath(dirUrl) {
		return sh.FS, path.Join(sh.Dir, dirUrl), nil
	}
	fsys, _, _, dir, err := NewDirFS(dirUrl)
	if err != nil {
		return nil, "", err
	}
	return fsys, dir, nil
}
