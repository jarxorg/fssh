package command

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/jarxorg/fssh"
)

type shEsc struct {
	flagSet *flag.FlagSet
}

func NewShEsc() fssh.Command {
	return &shEsc{}
}

func (c *shEsc) Name() string {
	return "!"
}

func (c *shEsc) Description() string {
	return "shell escape"
}

func (c *shEsc) FlagSet() *flag.FlagSet {
	if c.flagSet == nil {
		s := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		s.Usage = func() {}
		c.flagSet = s
	}
	return c.flagSet
}

func (c *shEsc) Reset() {
}

func (c *shEsc) Exec(sh *fssh.Shell) error {
	args := c.FlagSet().Args()
	if len(args) == 0 {
		c.Usage(sh.Stderr)
		return nil
	}

	name, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}

	cmd := exec.Command(name, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (c *shEsc) AutoCompleter() fssh.AutoCompleter {
	return nil
}

func (c *shEsc) Usage(w io.Writer) {
	fmt.Fprintf(w, "Usage:\n  %s [local shell commands]\n", c.Name())
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  ! ls -al")
	fmt.Fprintln(w, "  ! vi example.txt")
}

func init() {
	fssh.RegisterNewCommandFunc(NewShEsc)
}
