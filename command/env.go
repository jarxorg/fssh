package command

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jarxorg/fssh"
)

type env struct {
	flagSet *flag.FlagSet
}

func newEnv() fssh.Command {
	return &env{}
}

func (c *env) Name() string {
	return "env"
}

func (c *env) Description() string {
	return "prints or sets environment"
}

func (c *env) FlagSet() *flag.FlagSet {
	if c.flagSet == nil {
		s := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		s.Usage = func() {}
		c.flagSet = s
	}
	return c.flagSet
}

func (c *env) Reset() {
}

func (c *env) Exec(sh *fssh.Shell) error {
	args := c.FlagSet().Args()
	if len(args) == 0 {
		for _, e := range os.Environ() {
			fmt.Fprintf(sh.Stdout, "%s\n", e)
		}
		return nil
	}
	set := 0
	for _, arg := range args {
		kv := strings.SplitN(arg, "=", 2)
		if len(kv) == 2 {
			set++
			os.Setenv(kv[0], kv[1])
		} else {
			fmt.Fprintf(os.Stdout, "%s=%s\n", arg, os.Getenv(arg))
		}
	}
	if set > 0 {
		// NOTE: Re-create FS for apply environments.
		fsys, _, _, _, err := fssh.NewFS(sh.Host)
		if err != nil {
			return nil
		}
		sh.FS = fsys
	}
	return nil
}

func (c *env) AutoCompleter() fssh.AutoCompleter {
	return nil
}

func (c *env) Usage(w io.Writer) {
	name := c.Name()
	fmt.Fprintf(w, "Usage:\n  %s ([KEY](=[VALUE]))\n", name)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintf(w, "  %s           # Show all environs\n", name)
	fmt.Fprintf(w, "  %s KEY       # Show value of KEY\n", name)
	fmt.Fprintf(w, "  %s KEY=VALUE # Set environ\n", name)
}

func init() {
	fssh.RegisterNewCommandFunc(newEnv)
}
