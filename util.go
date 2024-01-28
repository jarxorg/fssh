package fssh

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"path"
	"strings"

	gobsargs "github.com/gobs/args"
)

// osUserHomeDir is a simple function that calls os.UserHomeDir for unit tests.
var osUserHomeDir = func() (string, error) {
	return os.UserHomeDir()
}

// IsCurrentPath checks the specified name has ":/"
func IsCurrentPath(name string) bool {
	return !strings.HasPrefix(name, "~") && !strings.Contains(name, ":/")
}

// ParseArgs parses the specified line to args.
func ParseArgs(line string) []string {
	return gobsargs.GetArgs(line)
}

// ParseURI parses the specified uri to protocol, host, filename.
// If the uri starts with ~~ it is replaced with the local current filename.
// If the uri starts with ~, it is replaced with the local home filename.
func ParseURI(uri string) (protocol, host, filename string, err error) {
	if strings.HasPrefix(uri, "~") {
		if strings.HasPrefix(uri[1:], "~") {
			host = "."
			filename = path.Clean(strings.TrimLeft(uri[2:], "/"))
			return
		}
		homeDir, e := osUserHomeDir()
		if e != nil {
			err = e
			return
		}
		host = homeDir
		filename = path.Clean(strings.TrimLeft(uri[1:], "/"))
		return
	}
	u, e := url.Parse(uri)
	if e != nil {
		err = e
		return
	}
	switch u.Scheme {
	case "s3", "gs", "mem":
		protocol = u.Scheme + "://"
		host = u.Host
		filename = path.Clean(strings.TrimLeft(u.Path, "/"))
	case "file":
		host = u.Host
		filename = path.Clean(strings.TrimLeft(u.Path, "/"))
	default:
		host = "."
		filename = path.Clean(strings.TrimLeft(uri, "/"))
	}
	return
}

const (
	unitKb = 1024
	unitMb = 1024 * 1024
	unitGb = 1024 * 1024 * 1024
	unitTb = 1024 * 1024 * 1024 * 1024
	unitPb = 1024 * 1024 * 1024 * 1024 * 1024
)

// DisplaySize returns summary of size.
func DisplaySize(size int64) string {
	if size < unitKb {
		return fmt.Sprintf("%4dB", size)
	}
	if size < unitMb {
		return fmt.Sprintf("%4dK", int64(math.Round(float64(size)/float64(unitKb))))
	}
	if size < unitGb {
		return fmt.Sprintf("%4dM", int64(math.Round(float64(size)/float64(unitMb))))
	}
	if size < unitTb {
		return fmt.Sprintf("%4dG", int64(math.Round(float64(size)/float64(unitGb))))
	}
	if size < unitPb {
		return fmt.Sprintf("%4dT", int64(math.Round(float64(size)/float64(unitTb))))
	}
	return fmt.Sprintf("%4dP", int64(math.Round(float64(size)/float64(unitPb))))
}

// IsGlobPattern checks pattern contains glob pattern.
func IsGlobPattern(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[]")
}

// WithPrefixes set prefix to each items.
func WithPrefixes(items []string, prefix string) []string {
	if prefix != "" {
		for i, item := range items {
			items[i] = prefix + item
		}
	}
	return items
}

// WithSuffixes set suffix to each items.
func WithSuffixes(items []string, suffix string) []string {
	if suffix != "" {
		for i, item := range items {
			items[i] = item + suffix
		}
		return items
	}
	return items
}

// SliceClone clones a slice.
func SliceClone[T comparable](src []T) []T {
	dest := make([]T, len(src))
	copy(dest, src)
	return dest
}
