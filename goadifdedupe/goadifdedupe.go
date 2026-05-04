// goadifdedupe: reformat preserve all ADIF file fields WITH deduping
// by Kenji Rikitake, JJ1BDX
// Usage: goaddifdedupe [-f infile] [-o outfile]
//
// This is a skeleton code set for adding further processing
//
// Coding convention:
// Use reader for reading each record (with DedupeADIFReader)
// Use writer for writing each record (with ADIFWriter)

package main

import (
	"flag"
	"fmt"
	"github.com/jj1bdx/adifparser"
	"io"
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
			"goadifdedupe: reformat preserve all ADIF file fields WITH deduping")
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

	var writer adifparser.ADIFWriter
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
		writer = adifparser.NewADIFWriter(writefp)
	} else {
		writefp = nil
		writer = adifparser.NewADIFWriter(os.Stdout)
	}

	if cerr := writer.SetComment("goadifdedupe\n"); cerr != nil {
		fmt.Fprintln(os.Stderr, cerr)
		os.Exit(1)
	}

	// For not deduping, use this filter API:
	// reader := adifparser.NewADIFReader(fp)

	// WITH deduping
	reader := adifparser.NewDedupeADIFReader(fp)

	for record, err := reader.ReadRecord(); record != nil || err != nil; record, err = reader.ReadRecord() {
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, err)
			}
			break // when io.EOF break the loop!
		}

		// process things here with the record
		if werr := writer.WriteRecord(record); werr != nil {
			fmt.Fprintln(os.Stderr, werr)
			os.Exit(1)
		}

	}

	// Surface flush errors so shell pipelines see a non-zero exit on
	// full disk / broken pipe.
	if ferr := writer.Flush(); ferr != nil {
		fmt.Fprintln(os.Stderr, ferr)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Total records: %d\n", reader.RecordCount())
}
