// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	"honnef.co/go/codesearch/fs"
	"honnef.co/go/codesearch/index"
	"honnef.co/go/codesearch/regexp"
)

var usageMessage = `usage: csearch [-c] [-f fileregexp] [-h] [-i] [-l] [-n] regexp

Csearch behaves like grep over all indexed files, searching for regexp,
an RE2 (nearly PCRE) regular expression.

The -c, -h, -i, -l, and -n flags are as in grep, although note that as per Go's
flag parsing convention, they cannot be combined: the option pair -i -n 
cannot be abbreviated to -in.

The -f flag restricts the search to files whose names match the RE2 regular
expression fileregexp.

Csearch relies on the existence of an up-to-date index created ahead of time.
To build or rebuild the index that csearch uses, run:

	cindex path...

where path... is a list of directories or individual files to be included in the index.
If no index exists, this command creates one.  If an index already exists, cindex
overwrites it.  Run cindex -help for more.

Csearch uses the index stored in $CSEARCHINDEX or, if that variable is unset or
empty, $HOME/.csearchindex.
`

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(2)
}

type globsFlag []string

func (s *globsFlag) String() string {
	return strings.Join(*s, " ")
}

func (s *globsFlag) Set(v string) error {
	if !fs.ValidGlob(v) {
		return fmt.Errorf("%q is an invalid glob", v)
	}
	*s = append(*s, v)
	return nil
}

var (
	fInclude globsFlag
	fExclude globsFlag

	fFilter          = flag.String("f", "", "search only files with names matching this regexp")
	fCaseInsensitive = flag.Bool("i", false, "case-insensitive search")
	fVerbose         = flag.Bool("verbose", false, "print extra information")
	fBrute           = flag.Bool("brute", false, "brute force - search all files in index")
	cpuProfile       = flag.String("cpuprofile", "", "write cpu profile to this file")

	matches bool
)

func init() {
	flag.Var(&fInclude, "include", "Only search files whose base name matches `glob`")
	flag.Var(&fExclude, "exclude", "Skip files whose base name matches `glob`")
}

func Main() {
	g := regexp.Grep{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	g.AddFlags()

	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		usage()
	}

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	pat := "(?m)" + args[0]
	if *fCaseInsensitive {
		pat = "(?i)" + pat
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		log.Fatal(err)
	}
	g.Regexp = re
	var fre *regexp.Regexp
	if *fFilter != "" {
		fre, err = regexp.Compile(*fFilter)
		if err != nil {
			log.Fatal(err)
		}
	}
	q := index.RegexpQuery(re.Syntax)
	if *fVerbose {
		log.Printf("query: %s\n", q)
	}

	ix := index.Open(index.File())
	ix.Verbose = *fVerbose
	var post []uint32
	if *fBrute {
		post = ix.PostingQuery(&index.Query{Op: index.QAll})
	} else {
		post = ix.PostingQuery(q)
	}
	if *fVerbose {
		log.Printf("post query identified %d possible files\n", len(post))
	}

	fnames := make([]uint32, 0, len(post))
	for _, fileid := range post {
		name := ix.Name(fileid)
		if fre != nil && fre.MatchString(name, true, true) < 0 {
			continue
		}
		if len(fInclude) > 0 && !fs.MatchAny(fInclude, name) {
			continue
		}
		if len(fExclude) > 0 && fs.MatchAny(fExclude, name) {
			continue
		}
		fnames = append(fnames, fileid)
	}

	if *fVerbose {
		log.Printf("filename regexp matched %d files\n", len(fnames))
	}
	post = fnames

	for _, fileid := range post {
		name := ix.Name(fileid)
		f, err := fs.Open(name)
		if err != nil {
			// XXX
			continue
		}
		g.Reader(f, name)
		_ = f.Close()
	}

	matches = g.Match
}

func main() {
	Main()
	if !matches {
		os.Exit(1)
	}
	os.Exit(0)
}
