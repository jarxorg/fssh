package fssh

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
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
	homeDir, err := os.UserHomeDir()
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
		args := ParseArgs(line)
		if len(args) == 0 {
			continue
		}
		cmd := AquireCommand(args[0])
		if cmd == nil {
			fmt.Fprintf(sh.Stderr, "%s: command not found: %s\n", ShellName, args[0])
			continue
		}
		if err := func() error {
			defer ReleaseCommand(cmd)

			if err := cmd.FlagSet().Parse(args[1:]); err != nil {
				if err == flag.ErrHelp {
					cmd.Usage(sh.Stdout)
					return nil
				}
				return err
			}
			return cmd.Exec(sh)
		}(); err != nil {
			if errors.Is(err, ErrExit) {
				return nil
			}
			fmt.Fprintf(sh.Stderr, "%s: error: %v\n", ShellName, err)
		}
	}
	return nil
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
func (sh *Shell) SubFS(dirUrl string) (FS, string, error) {
	if IsCurrentPath(dirUrl) {
		return sh.FS, path.Join(sh.Dir, dirUrl), nil
	}
	fsys, _, _, dir, err := NewFS(dirUrl)
	if err != nil {
		return nil, "", err
	}
	return fsys, dir, nil
}
