package fs

import (
	"io"
	"os"
	"path/filepath"

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
