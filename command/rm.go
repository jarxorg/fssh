package command

import (
	"flag"
	"fmt"
	"io"
	"io/fs"

	"github.com/jarxorg/fssh"
	"github.com/jarxorg/wfs"
)

type rm struct {
	flagSet     *flag.FlagSet
	isRecursive bool
	isForce     bool
	isDryRun    bool
}

func newRm() fssh.Command {
	return &rm{}
}

func (c *rm) Name() string {
	return "rm"
}

func (c *rm) Description() string {
	return "remove files"
}

func (c *rm) FlagSet() *flag.FlagSet {
	if c.flagSet == nil {
		s := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		s.BoolVar(&c.isRecursive, "r", false, "remove directories recursively")
		s.BoolVar(&c.isForce, "f", false, "forse")
		s.BoolVar(&c.isDryRun, "d", false, "dry run")
		c.flagSet = s
	}
	return c.flagSet
}

func (c *rm) Reset() {
	c.isRecursive = false
	c.isForce = false
	c.isDryRun = false
}

func (c *rm) Exec(sh *fssh.Shell) error {
	args := c.FlagSet().Args()
	if len(args) < 1 {
		c.Usage(sh.Stderr)
		return nil
	}
	for _, arg := range args {
		fsys, name, err := sh.SubFS(arg)
		if err != nil {
			return err
		}
		info, err := fs.Stat(fsys, name)
		if err != nil {
			return err
		}
		if info.IsDir() && !c.isRecursive {
			return fmt.Errorf("%s is a directory", name)
		}
		if c.isDryRun {
			fmt.Fprintf(sh.Stdout, "dry-run: remove %s\n", name)
			return nil
		}
		if err := wfs.RemoveAll(fsys, name); err != nil {
			return err
		}
	}
	return nil
}

func (c *rm) AutoCompleter() fssh.AutoCompleterFunc {
	return c.autoComplete
}

func (c *rm) autoComplete(sh *fssh.Shell, arg string) ([]string, error) {
	if c.isRecursive {
		return sh.PrefixMatcher.MatchDirs(sh, arg)
	}
	return sh.PrefixMatcher.MatchFiles(sh, arg)
}

func (c *rm) Usage(w io.Writer) {
	name := c.Name()
	fmt.Fprintf(w, "Usage:\n  %s ([flags]) [name]\n", name)
	fmt.Fprintln(w, "Flags:")
	c.FlagSet().SetOutput(w)
	c.FlagSet().PrintDefaults()
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintf(w, "  %s FILE\n", name)
	fmt.Fprintf(w, "  %s -rf DIR\n", name)
	fmt.Fprintf(w, "  %s (s3|gs)://BUCKET/FILE\n", name)
	fmt.Fprintf(w, "  %s -rf (s3|gs)://BUCKET/DIR\n", name)
}

func init() {
	fssh.RegisterNewCommandFunc(newRm)
}
