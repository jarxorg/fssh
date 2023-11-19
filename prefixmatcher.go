package fssh

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

// PrefixMatcher provides a functions to match prefix.
type PrefixMatcher interface {
	// Matches returns files and directories that match the prefix.
	Matches(sh *Shell, prefix string) ([]string, error)
	// Matches returns files that match the prefix.
	MatchFiles(sh *Shell, prefix string) ([]string, error)
	// Matches returns directories that match the prefix.
	MatchDirs(sh *Shell, prefix string) ([]string, error)
	// Reset is called when the shell status was updated.
	Reset()
}

const (
	flgGlobPrefixFiles = 0b01
	flgGlobPrefixDirs  = 0b10
)

type GlobPrefixMatcher struct {
	cachePrefix        string
	cachePrefixMatches []string
}

var _ PrefixMatcher = (*GlobPrefixMatcher)(nil)

// Reset clears internal cache.
func (m *GlobPrefixMatcher) Reset() {
	m.cachePrefix = ""
	m.cachePrefixMatches = nil
}

// Matches returns files and directories that match the prefix.
func (m *GlobPrefixMatcher) Matches(sh *Shell, prefix string) ([]string, error) {
	return m.matches(sh, prefix, flgGlobPrefixFiles|flgGlobPrefixDirs)
}

// Matches returns files that match the prefix.
func (m *GlobPrefixMatcher) MatchFiles(sh *Shell, prefix string) ([]string, error) {
	return m.matches(sh, prefix, flgGlobPrefixFiles)
}

// Matches returns directories that match the prefix.
func (m *GlobPrefixMatcher) MatchDirs(sh *Shell, prefix string) ([]string, error) {
	return m.matches(sh, prefix, flgGlobPrefixDirs)
}

func (m *GlobPrefixMatcher) matches(sh *Shell, prefix string, flgs int) ([]string, error) {
	if m.cachePrefix != "" {
		if prefix == m.cachePrefix {
			return SliceClone(m.cachePrefixMatches), nil
		}
		if strings.HasPrefix(prefix, m.cachePrefix) {
			var newMatches []string
			for _, match := range m.cachePrefixMatches {
				if strings.HasPrefix(match, prefix) {
					newMatches = append(newMatches, match)
				}
			}
			m.cachePrefix = prefix
			m.cachePrefixMatches = newMatches
			return SliceClone(newMatches), nil
		}
	}
	fsys, pattern, err := m.prefixSubFS(sh, prefix)
	if err != nil {
		return nil, err
	}
	matches, err := fs.Glob(fsys, pattern)
	if err != nil {
		return nil, err
	}
	if flgs&flgGlobPrefixFiles != 0 && flgs&flgGlobPrefixDirs != 0 {
		return m.normalizeMatches(sh, prefix, matches)
	}
	if flgs&flgGlobPrefixFiles != 0 {
		var files []string
		for _, match := range matches {
			info, err := fs.Stat(fsys, match)
			if err != nil {
				return nil, err
			}
			if !info.IsDir() {
				files = append(files, match)
			}
		}
		return m.normalizeMatches(sh, prefix, files)
	}
	var dirs []string
	for _, match := range matches {
		info, err := fs.Stat(fsys, match)
		if err != nil {
			return nil, err
		}
		if info.IsDir() && match != "." {
			dirs = append(dirs, match)
		}
	}
	return m.normalizeMatches(sh, prefix, dirs)
}

func (m *GlobPrefixMatcher) prefixSubFS(sh *Shell, prefix string) (FS, string, error) {
	if IsCurrentPath(prefix) {
		if prefix == "" || prefix == "." || prefix == "./" {
			return sh.FS, path.Join(sh.Dir, "./*"), nil
		}
		return sh.FS, path.Join(sh.Dir, prefix+"*"), nil
	}
	fsys, _, _, dir, err := NewFS(prefix)
	if err != nil {
		return nil, "", err
	}
	return fsys, dir + "*", nil
}

func (m *GlobPrefixMatcher) normalizeMatches(sh *Shell, prefix string, matches []string) ([]string, error) {
	if IsCurrentPath(prefix) {
		for i := range matches {
			rel, err := filepath.Rel(sh.Dir, matches[i])
			if err != nil {
				return nil, err
			}
			if strings.HasPrefix(prefix, "./") || strings.HasPrefix(prefix, ".") {
				rel = "./" + rel
			}
			matches[i] = rel
		}
	}
	m.cachePrefix = prefix
	m.cachePrefixMatches = matches
	return SliceClone(matches), nil
}
