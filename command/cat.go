package command

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"path"

	"github.com/jarxorg/fssh"
)

type cat struct {
	flagSet *flag.FlagSet
}

func newCat() fssh.Command {
	return &cat{}
}

func (c *cat) Name() string {
	return "cat"
}

func (c *cat) Description() string {
	return "concatenate and print files"
}

func (c *cat) FlagSet() *flag.FlagSet {
	if c.flagSet == nil {
		s := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		s.Usage = func() {}
		c.flagSet = s
	}
	return c.flagSet
}

func (c *cat) Reset() {
}

func (c *cat) Exec(sh *fssh.Shell) error {
	args := c.FlagSet().Args()
	if len(args) == 0 {
		c.Usage(sh.Stderr)
		return nil
	}
	filename := path.Join(sh.Dir, args[0])
	bin, err := fs.ReadFile(sh.FS, filename)
	if err != nil {
		return err
	}
	fmt.Fprintf(sh.Stdout, "%s", bin)
	if len(bin) > 0 && bin[len(bin)-1] != '\n' {
		fmt.Fprint(sh.Stdout, "\n")
	}
	return nil
}

func (c *cat) AutoCompleter() fssh.AutoCompleter {
	return c.autoComplete
}

func (c *cat) autoComplete(sh *fssh.Shell, arg string) ([]string, error) {
	return sh.PrefixMatcher.MatchFiles(sh, arg)
}

func (c *cat) Usage(w io.Writer) {
	name := c.Name()
	fmt.Fprintf(w, "Usage:\n  %s [file]\n", name)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintf(w, "  %s FILE\n", name)
	fmt.Fprintf(w, "  %s (s3|gs)://BUCKET/DIR/FILE\n", name)
}

func init() {
	fssh.RegisterNewCommandFunc(newCat)
}
