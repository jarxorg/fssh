package command

import (
	"flag"
	"fmt"
	"io"
	"io/fs"

	"github.com/jarxorg/fssh"
)

type ls struct {
	flagSet *flag.FlagSet
	isLong  bool
}

func newLs() fssh.Command {
	return &ls{}
}

func (c *ls) Name() string {
	return "ls"
}

func (c *ls) Description() string {
	return "list directory contents"
}

func (c *ls) FlagSet() *flag.FlagSet {
	if c.flagSet == nil {
		s := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		s.Usage = func() {}
		s.BoolVar(&c.isLong, "l", false, "long format")
		c.flagSet = s
	}
	return c.flagSet
}

func (c *ls) Reset() {
	c.isLong = false
}

func (c *ls) Exec(sh *fssh.Shell) error {
	args := c.FlagSet().Args()
	name := "."
	if len(args) > 0 {
		name = args[0]
	}
	subFs, subName, err := sh.SubFS(name)
	if err != nil {
		return err
	}
	if fssh.IsGlobPattern(subName) {
		matches, err := fs.Glob(subFs, subName)
		if err != nil {
			return err
		}
		for _, match := range matches {
			info, err := fs.Stat(subFs, match)
			if err != nil {
				return err
			}
			c.printInfo(sh, info)
		}
		return nil
	}
	info, err := fs.Stat(subFs, subName)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		c.printInfo(sh, info)
		return nil
	}
	entries, err := fs.ReadDir(subFs, subName)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return err
		}
		c.printInfo(sh, info)
	}
	return nil
}

func (c *ls) printInfo(sh *fssh.Shell, info fs.FileInfo) {
	name := info.Name()
	if c.isLong {
		modTime := info.ModTime().Format("2006-01-02 15:04")
		size := fssh.DisplaySize(info.Size())
		fmt.Fprintf(sh.Stdout, "%s %s %s %s\n", info.Mode(), modTime, size, name)
		return
	}
	tailSlash := ""
	if info.IsDir() {
		tailSlash = "/"
	}
	fmt.Fprintf(sh.Stdout, "%s%s\n", name, tailSlash)

}

func (c *ls) AutoCompleter() fssh.AutoCompleterFunc {
	return c.autoComplete
}

func (c *ls) autoComplete(sh *fssh.Shell, arg string) ([]string, error) {
	return sh.PrefixMatcher.Matches(sh, arg)
}

func (c *ls) Usage(w io.Writer) {
	name := c.Name()
	fmt.Fprintf(w, "Usage:\n  %s ([flags]) ([dir])\n", name)
	fmt.Fprintln(w, "Flags:")
	c.FlagSet().SetOutput(w)
	c.FlagSet().PrintDefaults()
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintf(w, "  %s DIR\n", name)
	fmt.Fprintf(w, "  %s (s3|gs)://BUCKET/DIR\n", name)
}

func init() {
	fssh.RegisterNewCommandFunc(newLs)
}
