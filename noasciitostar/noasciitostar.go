// noasciitostar:
// convert UTF-8 no-ASCII character (i.e., 2 or more bytes/character)
// in the file by the same byte length of ASCII '*' letters
// by Kenji Rikitake, JJ1BDX
// Usage: noasciitostar [-f infile] [-o outfile]
//
// This is a skeleton code set for adding further processing
//
// Coding convention:
// Use reader for reading each record
// Use writer for writing each record

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

func main() {
	var infile = flag.String("f", "", "input file (stdout in none)")
	var outfile = flag.String("o", "", "output file (stdout if none)")

	var fp *os.File
	var err error

	flag.Usage = func() {
		execname := os.Args[0]
		fmt.Fprintln(flag.CommandLine.Output(),
			"noasciitostar: convert UTF-8 no-ASCII character\n"+
				"(i.e., 2 or more bytes/character) in the file\n"+
				"by the same byte length of ASCII '*' letters\n")
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [-f infile] [-o outfile]\n", execname)
		flag.PrintDefaults()
	}

	flag.Parse()

	if *infile == "" {
		fp = os.Stdin
	} else {
		fp, err = os.Open(*infile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer fp.Close()
	}
	reader := bufio.NewReader(fp)

	var writefp *os.File
	var writer *bufio.Writer
	if *outfile != "" {
		// O_EXCL: atomic create-if-absent that refuses to follow a
		// final-component symlink. Replaces the Stat/Create TOCTOU.
		writefp, err = os.OpenFile(*outfile,
			os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer writefp.Close()
		writer = bufio.NewWriter(writefp)
	} else {
		writefp = nil
		writer = bufio.NewWriter(os.Stdout)
	}

	for r, n, err := reader.ReadRune(); err == nil; r, n, err = reader.ReadRune() {
		if n > 1 {
			for i := 0; i < n; i++ {
				// For non-ASCII letter,
				// convert it to a star of the same byte length
				if werr := writer.WriteByte('*'); werr != nil {
					fmt.Fprintln(os.Stderr, werr)
					os.Exit(1)
				}
			}
		} else {
			if _, werr := writer.WriteRune(r); werr != nil {
				fmt.Fprintln(os.Stderr, werr)
				os.Exit(1)
			}
		}
	}

	if ferr := writer.Flush(); ferr != nil {
		fmt.Fprintln(os.Stderr, ferr)
		os.Exit(1)
	}
}
