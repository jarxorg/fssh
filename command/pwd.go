package command

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/jarxorg/fssh"
)

type pwd struct {
	flagSet *flag.FlagSet
}

func newPwd() fssh.Command {
	return &pwd{}
}

func (c *pwd) Name() string {
	return "pwd"
}

func (c *pwd) FlagSet() *flag.FlagSet {
	if c.flagSet == nil {
		s := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		s.Usage = func() {}
		c.flagSet = s
	}
	return c.flagSet
}

func (c *pwd) Reset() {
}

func (c *pwd) Exec(sh *fssh.Shell) error {
	if sh.Protocol == "" {
		abs, err := filepath.Abs(filepath.Join(sh.Host, sh.Dir))
		if err != nil {
			return err
		}
		fmt.Fprintf(sh.Stdout, "%s\n", abs)
		return nil
	}
	fmt.Fprintf(sh.Stdout, "%s\n", sh.DirWithProtocol())
	return nil
}

func (c *pwd) AutoCompleter() fssh.AutoCompleter {
	return nil
}

func (c *pwd) Usage(w io.Writer) {
	fmt.Fprintf(w, "Usage:\n  %s\n", c.Name())
}

func init() {
	fssh.RegisterNewCommandFunc(newPwd)
}
