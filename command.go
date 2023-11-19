package fssh

import (
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/chzyer/readline"
)

// Command is an interface that defines a shell command.
type Command interface {
	// Name returns the name of a command.
	Name() string
	// Description returns the description of a command.
	Description() string
	// FlagSet returns the flagSet of a command.
	FlagSet() *flag.FlagSet
	// Exec executes a command.
	Exec(sh *Shell) error
	// Usage writes help usage.
	Usage(w io.Writer)
	// AutoCompleter returns a AutoCompleter if the command supports auto completion.
	AutoCompleter() AutoCompleterFunc
	// Reset resets the command status. This is called after Exec.
	Reset()
}

// NewCommandFunc represents a function to create a new command.
type NewCommandFunc func() Command

// AutoCompleterFunc represent a function for auto completion.
type AutoCompleterFunc func(sh *Shell, arg string) ([]string, error)

var (
	commandPools       = map[string]*sync.Pool{}
	commandNames       []string
	commandNamesSorted bool
)

// RegisterNewCommandFunc registers a specified NewCommandFunc.
func RegisterNewCommandFunc(fn NewCommandFunc) {
	cmd := fn()
	commandPools[cmd.Name()] = &sync.Pool{
		New: func() any { return fn() },
	}
	commandNames = append(commandNames, cmd.Name())
	commandNamesSorted = false
}

// AquireCommand returns an Command instance from command pool.
func AquireCommand(name string) Command {
	if pool, ok := commandPools[name]; ok {
		return pool.Get().(Command)
	}
	return nil
}

// ReleaseCommand releases a acquired Command via AquireCommand to command pool.
func ReleaseCommand(cmd Command) {
	if pool, ok := commandPools[cmd.Name()]; ok {
		cmd.Reset()
		pool.Put(cmd)
	}
}

// SortedCommandNames returns sorted command names from registered commands.
func SortedCommandNames() []string {
	if !commandNamesSorted {
		sort.Strings(commandNames)
		commandNamesSorted = true
	}
	return commandNames
}

func newReadlineAutoCompleter(sh *Shell) readline.AutoCompleter {
	var pcs []readline.PrefixCompleterInterface
	for _, name := range SortedCommandNames() {
		pcs = append(pcs, newReadlinePrefixCompleter(sh, name))
	}
	return readline.NewPrefixCompleter(pcs...)
}

func newReadlinePrefixCompleter(sh *Shell, name string) readline.PrefixCompleterInterface {
	cmd := AquireCommand(name)
	defer ReleaseCommand(cmd)

	pc := readline.PcItem(name)
	ac := cmd.AutoCompleter()
	if ac == nil {
		return pc
	}
	pc.Children = append(pc.Children, readline.PcItemDynamic(func(line string) []string {
		cmd := AquireCommand(name)
		defer ReleaseCommand(cmd)

		args := ParseArgs(line)[1:]
		if err := cmd.FlagSet().Parse(args); err != nil {
			return nil
		}
		fargs := cmd.FlagSet().Args()
		lastFarg := ""
		if len(fargs) > 0 {
			lastFarg = fargs[len(fargs)-1]
		}
		matches, err := ac(sh, lastFarg)
		if err != nil {
			fmt.Fprintf(sh.Stderr, "%s: failed to auto complete: %v\n", ShellName, err)
			return nil
		}
		prefix := strings.TrimSpace(line[len(cmd.Name()):])
		if len(fargs) > 0 {
			prefix = strings.TrimSpace(strings.TrimSuffix(prefix, lastFarg))
		}
		return WithPrefixes(matches, prefix+" ")
	}))
	return pc
}
