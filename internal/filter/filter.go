package filter

import (
	"compress/gzip"
	"io"
	"regexp"
)

type readCloser struct {
	io.Reader
	io.Closer
}

type filter func(rc io.ReadCloser) (io.ReadCloser, error)

var filters = map[*regexp.Regexp]filter{
	regexp.MustCompile(`\.gz$`): filterUncompressGz,
}

func filterUncompressGz(rc io.ReadCloser) (io.ReadCloser, error) {
	r, err := gzip.NewReader(rc)
	return readCloser{r, rc}, err
}

func Filter(rc io.ReadCloser, path string) (io.ReadCloser, error) {
	for rx, f := range filters {
		if rx.Match([]byte(path)) {
			rc2, err := f(rc)
			if err != nil {
				_ = rc.Close()
				return nil, err
			}
			rc = rc2
		}
	}

	return rc, nil
}
