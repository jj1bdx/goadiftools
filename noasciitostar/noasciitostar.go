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
			fmt.Fprint(os.Stderr, err)
			return
		}
	}
	reader := bufio.NewReader(fp)

	var writefp *os.File
	var writer *bufio.Writer
	if *outfile != "" {
		if _, err := os.Stat(*outfile); os.IsNotExist(err) {
			// File does not exist: create it
			writefp, err = os.Create(*outfile)
			if err != nil {
				fmt.Fprint(os.Stderr, err)
				return
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: file %s already exists\n", *outfile)
			return
		}
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
				err := writer.WriteByte('*')
				if err != nil {
					fmt.Fprint(os.Stderr, err)
				}
			}
		} else {
			_, err := writer.WriteRune(r)
			if err != nil {
				fmt.Fprint(os.Stderr, err)
			}
		}
	}

	// Flush and close the output
	writer.Flush()
	if writefp != os.Stdout {
		writefp.Close()
	}
}
