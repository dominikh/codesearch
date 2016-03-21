package fs

import (
	"io"
	"os"

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
