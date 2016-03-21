package fs

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"honnef.co/go/codesearch/filter"
)

// Open opens a file and wraps it in a series of file-specific
// filters.
func Open(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return filter.Filter(f, path)
}

// ValidGlob reports whether a glob is valid.
func ValidGlob(glob string) bool {
	_, err := filepath.Match(glob, "")
	return err == nil
}

// Match reports whether the base name of path matches glob.
func Match(glob, path string) bool {
	name := filepath.Base(path)
	ok, _ := filepath.Match(glob, name)
	return ok
}

// MatchAny reports whether the base name of path matches any of the
// globs.
func MatchAny(globs []string, path string) bool {
	for _, glob := range globs {
		if Match(glob, path) {
			return true
		}
	}
	return false
}

// IsInDirs reports whether a path lies under one of the directories.
func IsInDirs(dirs []string, path string) bool {
	for _, dir := range dirs {
		if dir == path {
			return true
		}
		if len(path) < len(dir) {
			continue
		}
		if strings.HasPrefix(path, dir) &&
			(dir[len(dir)-1] == filepath.Separator || path[len(dir)] == filepath.Separator) {
			return true
		}
	}
	return false
}
