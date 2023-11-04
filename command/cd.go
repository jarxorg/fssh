package command

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"path"

	"github.com/jarxorg/fssh"
)

type cd struct {
	flagSet *flag.FlagSet
}

func newCd() fssh.Command {
	return &cd{}
}

func (c *cd) Name() string {
	return "cd"
}

func (c *cd) Description() string {
	return "change directory"
}

func (c *cd) FlagSet() *flag.FlagSet {
	if c.flagSet == nil {
		s := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		s.Usage = func() {}
		c.flagSet = s
	}
	return c.flagSet
}

func (c *cd) Reset() {
}

func (c *cd) Exec(sh *fssh.Shell) error {
	args := c.FlagSet().Args()
	if len(args) == 0 {
		sh.Dir = ""
		sh.UpdatePrompt()
		return nil
	}
	dir := args[0]
	if !fssh.IsCurrentPath(dir) {
		fsys, protocol, baseDir, subDir, err := fssh.NewFS(dir)
		if err != nil {
			return err
		}
		info, err := fs.Stat(fsys, subDir)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("not directory: %s", dir)
		}
		sh.FS = fsys
		sh.Protocol = protocol
		sh.Host = baseDir
		sh.Dir = subDir
		sh.UpdatePrompt()
		return nil
	}
	dir = path.Join(sh.Dir, dir)
	info, err := fs.Stat(sh.FS, dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not directory: %s", dir)
	}
	sh.Dir = dir
	sh.UpdatePrompt()
	return nil
}

func (c *cd) AutoCompleter() fssh.AutoCompleter {
	return c.autoComplete
}

func (c *cd) autoComplete(sh *fssh.Shell, arg string) ([]string, error) {
	return sh.PrefixMatcher.MatchDirs(sh, arg)
}

func (c *cd) Usage(w io.Writer) {
	name := c.Name()
	fmt.Fprintf(w, "Usage:\n  %s [dir]\n", name)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintf(w, "  %s DIR\n", name)
	fmt.Fprintf(w, "  %s (s3|gs)://BUCKET/DIR\n", name)
}

func init() {
	fssh.RegisterNewCommandFunc(newCd)
}
