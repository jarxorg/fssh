package command

import (
	"flag"
	"fmt"
	"io"

	"github.com/jarxorg/fssh"
)

type help struct {
	flagSet *flag.FlagSet
}

func newHelp() fssh.Command {
	return &help{}
}

func (c *help) Name() string {
	return "help"
}

func (c *help) Description() string {
	return "show help messages"
}

func (c *help) FlagSet() *flag.FlagSet {
	if c.flagSet == nil {
		s := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		s.Usage = func() {}
		c.flagSet = s
	}
	return c.flagSet
}

func (c *help) Reset() {
}

func (c *help) Exec(sh *fssh.Shell) error {
	args := c.FlagSet().Args()
	if len(args) == 0 || args[0] == c.Name() {
		c.Usage(sh.Stdout)
		return nil
	}
	if cmd := fssh.AquireCommand(args[0]); cmd != nil {
		defer fssh.ReleaseCommand(cmd)
		cmd.Usage(sh.Stdout)
	}
	return nil
}

func (c *help) Usage(w io.Writer) {
	helpName := c.Name()
	fmt.Fprintf(w, "Usage:\n  %s ([command])\n", helpName)
	fmt.Fprintln(w, "Commands:")
	for _, name := range fssh.SortedCommandNames() {
		if name != helpName {
			cmd := fssh.AquireCommand(name)
			fmt.Fprintf(w, "  %s\t\t%s\n", name, cmd.Description())
		}
	}
}

func (c *help) AutoCompleter() fssh.AutoCompleter {
	return nil
}

func init() {
	fssh.RegisterNewCommandFunc(newHelp)
}
