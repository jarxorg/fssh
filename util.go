package fssh

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	gobsargs "github.com/gobs/args"
)

func IsCurrentPath(name string) bool {
	return !strings.Contains(name, ":/")
}

func ParseArgs(line string) []string {
	return gobsargs.GetArgs(line)
}

func parseDir(name string) (protocol, host, dir string, err error) {
	u, e := url.Parse(name)
	if e != nil {
		err = e
		return
	}
	switch u.Scheme {
	case "s3", "gs", "mem":
		protocol = u.Scheme + "://"
		host = u.Host
		dir = path.Clean(strings.TrimLeft(u.Path, "/"))
	case "file":
		host = "."
		dir = path.Clean(u.Path)
	default:
		host = "."
		dir = path.Clean(name)
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

func DisplaySize(size int64) string {
	if size < unitKb {
		return fmt.Sprintf("%4dB", size)
	}
	if size < unitMb {
		return fmt.Sprintf("%4dK", size/unitKb)
	}
	if size < unitGb {
		return fmt.Sprintf("%4dM", size/unitMb)
	}
	if size < unitTb {
		return fmt.Sprintf("%4dG", size/unitGb)
	}
	if size < unitPb {
		return fmt.Sprintf("%4dT", size/unitTb)
	}
	return fmt.Sprintf("%4dP", size/unitPb)
}

func IsPattern(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[]")
}

func WithPrefixes(items []string, prefix, joiner string) []string {
	if prefix != "" {
		for i, item := range items {
			items[i] = prefix + joiner + item
		}
		return items
	}
	return items
}

func WithSuffixes(items []string, joiner, suffix string) []string {
	if suffix != "" {
		for i, item := range items {
			items[i] = item + joiner + suffix
		}
		return items
	}
	return items
}

func ArrayClone[T comparable](src []T) []T {
	dest := make([]T, len(src))
	copy(dest, src)
	return dest
}
