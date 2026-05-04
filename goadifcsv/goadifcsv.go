// goadifcsv: pick up specified ADIF fields and output in CSV format
// by Kenji Rikitake, JJ1BDX
// Usage: goadifcsv [-f infile] [-o outfile] field_names...
// Values of non-existing fields are set to empty strings

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/jj1bdx/adifparser"
	"io"
	"os"
	"strings"
)

func main() {
	var infile = flag.String("f", "", "input file (stdout in none)")
	var outfile = flag.String("o", "", "output file (stdout if none)")

	var fp *os.File
	var err error

	flag.Usage = func() {
		execname := os.Args[0]
		fmt.Fprintln(flag.CommandLine.Output(),
			"goadifcsv: remove specified ADIF fields")
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [-f infile] [-o outfile] field_names...\n", execname)
		fmt.Fprintf(flag.CommandLine.Output(),
			"Values of non-existing fields are set to empty strings\n")
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

	var writer *csv.Writer
	var writefp *os.File
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
		writer = csv.NewWriter(writefp)
	} else {
		writefp = nil
		writer = csv.NewWriter(os.Stdout)
	}

	fields := flag.Args()

	// Write field names first
	var fieldnames []string
	for i := range fields {
		fieldnames = append(fieldnames, strings.ToLower(fields[i]))
	}
	if werr := writer.Write(fieldnames); werr != nil {
		fmt.Fprintln(os.Stderr, werr)
		os.Exit(1)
	}

	// For deduping, use this filter API:
	// reader := adifparser.NewDedupeADIFReader(fp)

	reader := adifparser.NewADIFReader(fp)
	for record, err := reader.ReadRecord(); record != nil || err != nil; record, err = reader.ReadRecord() {
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, err)
			}
			break // when io.EOF break the loop!
		}
		// Write a CSV record with chosen fields
		newrecord := []string{}
		for i := range fields {
			newvalue, err := record.GetValue(
				strings.ToLower(fields[i]))
			if err == adifparser.ErrNoSuchField {
				newvalue = ""
			} else if err != nil {
				fmt.Fprintln(os.Stderr, err)
				break
			}
			newrecord = append(newrecord, newvalue)
		}
		if werr := writer.Write(newrecord); werr != nil {
			fmt.Fprintln(os.Stderr, werr)
			os.Exit(1)
		}

	}

	// Flush and surface any deferred write errors so shell pipelines
	// observe a non-zero exit on full disk / broken pipe.
	writer.Flush()
	if werr := writer.Error(); werr != nil {
		fmt.Fprintln(os.Stderr, werr)
		os.Exit(1)
	}
}
