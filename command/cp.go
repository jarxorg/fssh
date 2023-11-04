package command

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/jarxorg/fssh"
	"github.com/jarxorg/wfs"
)

type cp struct {
	flagSet     *flag.FlagSet
	isRecursive bool
	isForce     bool
	isDryRun    bool
}

func newCp() fssh.Command {
	return &cp{}
}

func (c *cp) Name() string {
	return "cp"
}

func (c *cp) Description() string {
	return "copy files"
}

func (c *cp) FlagSet() *flag.FlagSet {
	if c.flagSet == nil {
		s := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		s.Usage = func() {}
		s.BoolVar(&c.isRecursive, "r", false, "copy directories recursively")
		s.BoolVar(&c.isForce, "f", false, "forse")
		s.BoolVar(&c.isDryRun, "d", false, "dry run")
		c.flagSet = s
	}
	return c.flagSet
}

func (c *cp) Reset() {
	c.isRecursive = false
	c.isForce = false
	c.isDryRun = false
}

func (c *cp) Exec(sh *fssh.Shell) error {
	args := c.FlagSet().Args()
	if len(args) < 2 {
		c.Usage(sh.Stderr)
		return nil
	}
	from, to := args[0], args[1]
	fromFS, fromName, err := sh.SubFS(from)
	if err != nil {
		return err
	}
	toFS, toName, err := sh.SubFS(to)
	if err != nil {
		return err
	}
	fromInfo, err := fs.Stat(fromFS, fromName)
	if err != nil {
		return err
	}
	if fromInfo.IsDir() {
		return c.copyDir(sh, fromFS, toFS, fromName, toName)
	}
	return c.copyFile(sh, fromFS, toFS, fromName, toName)
}

func (c *cp) copyDir(sh *fssh.Shell, fromFS, toFS wfs.WriteFileFS, fromName, toName string) error {
	if !c.isRecursive {
		if c.isForce {
			return nil
		}
		return fmt.Errorf("%s is a directory (not copied)", fromName)
	}
	toInfo, err := fs.Stat(toFS, toName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if toInfo != nil && !toInfo.IsDir() {
		if c.isForce {
			return nil
		}
		return fmt.Errorf("%s is not a directory (not copied)", toName)
	}
	offset := 0
	if toInfo == nil {
		offset = len(path.Base(fromName))
	}
	return fs.WalkDir(fromFS, fromName, func(fromPath string, d fs.DirEntry, err error) error {
		if err != nil || d == nil {
			return err
		}
		toPath := path.Join(toName, strings.TrimLeft(fromPath[offset:], "/"))
		if d.IsDir() {
			if c.isDryRun {
				fmt.Fprintf(sh.Stdout, "dry-run: mkdir %s\n", toPath)
				return nil
			}
			return toFS.MkdirAll(toPath, os.ModePerm)
		}
		return c.copyFile(sh, fromFS, toFS, fromPath, toPath)
	})
}

func (c *cp) copyFile(sh *fssh.Shell, fromFS, toFS wfs.WriteFileFS, fromName, toName string) error {
	fromInfo, err := fs.Stat(fromFS, fromName)
	if err != nil {
		return err
	}
	toInfo, err := fs.Stat(toFS, toName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if toInfo != nil {
		if !c.isForce {
			fmt.Fprintf(sh.Stderr, "skip copying %s because %s exists\n", fromName, toName)
			return nil
		}
		if toInfo.IsDir() {
			toName = path.Join(toName, path.Base(fromName))
		}
	}
	if c.isDryRun {
		fmt.Fprintf(sh.Stdout, "dry-run: copy %s to %s\n", fromName, toName)
		return nil
	}

	fromFile, err := fromFS.Open(fromName)
	if err != nil {
		return err
	}
	defer fromFile.Close()

	toFile, err := toFS.CreateFile(toName, fromInfo.Mode())
	if err != nil {
		return err
	}
	defer toFile.Close()

	if _, err := io.Copy(toFile, fromFile); err != nil {
		return err
	}
	return nil
}

func (c *cp) AutoCompleter() fssh.AutoCompleter {
	return c.autoComplete
}

func (c *cp) autoComplete(sh *fssh.Shell, arg string) ([]string, error) {
	if c.isRecursive {
		return sh.PrefixMatcher.MatchDirs(sh, arg)
	}
	return sh.PrefixMatcher.MatchFiles(sh, arg)
}

func (c *cp) Usage(w io.Writer) {
	name := c.Name()
	fmt.Fprintf(w, "Usage:\n  %s ([flags]) [from] [to]\n", name)
	fmt.Fprintln(w, "Flags:")
	c.FlagSet().SetOutput(w)
	c.FlagSet().PrintDefaults()
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintf(w, "  %s FROM TO\n", name)
	fmt.Fprintf(w, "  %s LOCAL_FILE (s3|gs)://BUCKET/DIR\n", name)
	fmt.Fprintf(w, "  %s -rf (s3|gs)://BUCKET/DIR LOCAL_DIR\n", name)
}

func init() {
	fssh.RegisterNewCommandFunc(newCp)
}
