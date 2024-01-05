package fssh

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"reflect"
	"testing"
)

type testCommand struct {
	name        string
	description string
	flagSet     *flag.FlagSet
	execFunc    func(sh *Shell) error
	ac          AutoCompleterFunc
}

func (c *testCommand) Name() string {
	return c.name
}

func (c *testCommand) Description() string {
	return c.description
}

func (c *testCommand) FlagSet() *flag.FlagSet {
	if c.flagSet == nil {
		c.flagSet = flag.NewFlagSet(c.name, flag.ContinueOnError)
	}
	return c.flagSet
}

func (c *testCommand) Exec(sh *Shell) error {
	if c.execFunc != nil {
		return c.execFunc(sh)
	}
	fmt.Fprintf(sh.Stdout, "%s\n", c.name)
	return nil
}

func (c *testCommand) Usage(w io.Writer) {
	fmt.Fprintf(w, "Usage:\n  %s\n", c.name)
}

func (c *testCommand) AutoCompleter() AutoCompleterFunc {
	return c.ac
}

func (c *testCommand) Reset() {
}

func TestRegisterNewCommandFunc(t *testing.T) {
	commandPoolsOrg := commandPools
	commandNamesOrg := commandNames
	commandNamesSortedOrg := commandNamesSorted
	defer func() {
		commandPools = commandPoolsOrg
		commandNames = commandNamesOrg
		commandNamesSorted = commandNamesSortedOrg
	}()

	testCmd1 := &testCommand{name: "test1"}
	testCmd2 := &testCommand{name: "test2"}
	testCmd3 := &testCommand{name: "test3"}

	RegisterNewCommandFunc(func() Command { return testCmd1 })
	RegisterNewCommandFunc(func() Command { return testCmd2 })
	RegisterNewCommandFunc(func() Command { return testCmd3 })
	DeregisterNewCommandFunc("test2")

	wantCommandNames := []string{"test1", "test3"}
	if !reflect.DeepEqual(commandNames, wantCommandNames) {
		t.Errorf("got %v; want %v", commandNames, wantCommandNames)
	}
}

func TestSortedCommandNames(t *testing.T) {
	commandNamesOrg := commandNames
	commandNamesSortedOrg := commandNamesSorted
	defer func() {
		commandNames = commandNamesOrg
		commandNamesSorted = commandNamesSortedOrg
	}()

	commandNamesSorted = false
	commandNames = []string{"c", "b", "a"}

	want := []string{"a", "b", "c"}
	got := SortedCommandNames()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v; want %v", got, want)
	}
}

func Test_autoComplete(t *testing.T) {
	tests := []struct {
		sh   *Shell
		cmd  Command
		line string
		want []string
	}{
		{
			sh: &Shell{
				Stderr: io.Discard,
			},
			cmd: &testCommand{
				name:    "ls",
				flagSet: &flag.FlagSet{},
				ac: func(sh *Shell, arg string) ([]string, error) {
					return []string{"a0", "a1", "a2"}, nil
				},
			},
			line: "ls a",
			want: []string{"a0", "a1", "a2"},
		}, {
			sh: &Shell{
				Stderr: io.Discard,
			},
			cmd: &testCommand{
				name: "ls",
				flagSet: func() *flag.FlagSet {
					flagSet := &flag.FlagSet{}
					flagSet.Bool("l", false, "")
					return flagSet
				}(),
				ac: func(sh *Shell, arg string) ([]string, error) {
					return []string{"a0", "a1", "a2"}, nil
				},
			},
			line: "ls -l a",
			want: []string{"-l a0", "-l a1", "-l a2"},
		}, {
			sh: &Shell{
				Stderr: io.Discard,
			},
			cmd: &testCommand{
				name:    "ls",
				flagSet: &flag.FlagSet{},
				ac: func(sh *Shell, arg string) ([]string, error) {
					return []string{"a0", "a1", "a2"}, nil
				},
			},
			line: "ls -l a",
		}, {
			sh: &Shell{
				Stderr: io.Discard,
			},
			cmd: &testCommand{
				name:    "ls",
				flagSet: &flag.FlagSet{},
				ac: func(sh *Shell, arg string) ([]string, error) {
					return nil, errors.New("test-error")
				},
			},
			line: "ls a",
		},
	}
	for i, test := range tests {
		got := autoComplete(test.sh, test.cmd, test.line)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("tests[%d]: got %v; want %v", i, got, test.want)
		}
	}
}
