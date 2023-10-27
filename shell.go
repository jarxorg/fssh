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
	"github.com/jarxorg/gcsfs"
	"github.com/jarxorg/s3fs"
	"github.com/jarxorg/wfs"
	"github.com/jarxorg/wfs/memfs"
	"github.com/jarxorg/wfs/osfs"
)

const ShellName = "fssh"

var ErrExit = errors.New("exit")

func Main(osArgs []string) error {
	flagSet := flag.NewFlagSet(ShellName, flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Printf("Usage:\n  %s ([dir])\n", ShellName)
		fmt.Println("Examples:")
		fmt.Printf("  %s\n", ShellName)
		fmt.Printf("  %s DIR\n", ShellName)
		fmt.Printf("  %s (s3|gs)://BUCKET/\n", ShellName)
	}
	if err := flagSet.Parse(osArgs[1:]); err != nil {
		return err
	}

	args := flagSet.Args()
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}
	sh, err := New(dir)
	if err != nil {
		return err
	}
	defer sh.Close()

	return sh.Run()
}

type Shell struct {
	l *readline.Instance

	Stdout        io.Writer
	Stderr        io.Writer
	FS            wfs.WriteFileFS
	Protocol      string
	Host          string
	Dir           string
	PrefixMatcher PrefixMatcher
}

func NewFS(name string) (wfs.WriteFileFS, string, string, string, error) {
	protocol, host, dir, err := parseDir(name)
	if err != nil {
		return nil, "", "", "", err
	}
	switch protocol {
	case "s3://":
		return s3fs.New(host), protocol, host, dir, nil
	case "gs://":
		return gcsfs.New(host), protocol, host, dir, nil
	case "mem://":
		return memfs.New(), protocol, host, dir, nil
	default:
		return osfs.New(host), protocol, host, dir, nil
	}
}

func New(name string) (*Shell, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	fsys, protocol, host, dir, err := NewFS(name)
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
	l, err := readline.NewEx(&readline.Config{
		Prompt:            "\033[36m>\033[0m ",
		HistoryFile:       filepath.Join(homeDir, fmt.Sprintf(".%s_history", ShellName)),
		AutoComplete:      newReadlineAutoCompleter(sh),
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		return nil, err
	}
	sh.l = l
	sh.Stdout = l.Stdout()
	sh.Stderr = l.Stderr()
	sh.UpdatePrompt()
	return sh, nil
}

func (sh *Shell) DirWithProtocol() string {
	return sh.Protocol + path.Join(sh.Host, sh.Dir)
}

func (sh *Shell) UpdatePrompt() {
	sh.PrefixMatcher.Reset()
	sh.l.SetPrompt("\033[36m" + sh.DirWithProtocol() + ">\033[0m ")
}

func (sh *Shell) Close() error {
	return sh.l.Close()
}

func (sh *Shell) Run() error {
	sh.l.CaptureExitSignal()
	log.SetOutput(sh.Stderr)
	for {
		line, err := sh.l.Readline()
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

func (sh *Shell) Usage(w io.Writer) {
	w.Write([]byte("Commands:\n"))
	for _, name := range SortedCommandNames() {
		w.Write([]byte("  "))
		w.Write([]byte(name))
		w.Write([]byte("\n"))
	}
}

func (sh *Shell) SubFS(name string) (wfs.WriteFileFS, string, error) {
	if IsCurrentPath(name) {
		return sh.FS, path.Join(sh.Dir, name), nil
	}
	fsys, _, _, dir, err := NewFS(name)
	if err != nil {
		return nil, "", err
	}
	return fsys, dir, nil
}
