package command

import (
	"flag"
	"fmt"
	"io"

	"github.com/jarxorg/fssh"
)

type exit struct {
	flagSet *flag.FlagSet
}

func newExit() fssh.Command {
	return &exit{}
}

func (c *exit) Name() string {
	return "exit"
}

func (c *exit) Description() string {
	return "exit " + fssh.ShellName
}

func (c *exit) FlagSet() *flag.FlagSet {
	if c.flagSet == nil {
		s := flag.NewFlagSet("exit", flag.ContinueOnError)
		s.Usage = func() {}
		c.flagSet = s
	}
	return c.flagSet
}

func (c *exit) Reset() {
}

func (c *exit) Exec(sh *fssh.Shell) error {
	return fssh.ErrExit
}

func (c *exit) AutoCompleter() fssh.AutoCompleterFunc {
	return nil
}

func (c *exit) Usage(w io.Writer) {
	fmt.Fprintf(w, "Usage:\n  %s\n", c.Name())
}

func init() {
	fssh.RegisterNewCommandFunc(newExit)
}
