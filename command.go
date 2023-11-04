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

type Command interface {
	Name() string
	Description() string
	FlagSet() *flag.FlagSet
	Exec(sh *Shell) error
	Usage(w io.Writer)
	AutoCompleter() AutoCompleter
	Reset()
}

type NewCommandFunc func() Command

type AutoCompleter func(sh *Shell, arg string) ([]string, error)

var (
	commandPools       = map[string]*sync.Pool{}
	commandNames       []string
	commandNamesSorted bool
)

func RegisterNewCommandFunc(fn NewCommandFunc) {
	cmd := fn()
	commandPools[cmd.Name()] = &sync.Pool{
		New: func() any { return fn() },
	}
	commandNames = append(commandNames, cmd.Name())
	commandNamesSorted = false
}

func AquireCommand(name string) Command {
	if pool, ok := commandPools[name]; ok {
		return pool.Get().(Command)
	}
	return nil
}

func ReleaseCommand(cmd Command) {
	if pool, ok := commandPools[cmd.Name()]; ok {
		cmd.Reset()
		pool.Put(cmd)
	}
}

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
		return WithPrefixes(matches, prefix, " ")
	}))
	return pc
}
