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
			fmt.Fprint(os.Stderr, err)
			return
		}
	}

	var writer adifparser.ADIFWriter
	var writefp *os.File
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
		writer = adifparser.NewADIFWriter(writefp)
	} else {
		writefp = nil
		writer = adifparser.NewADIFWriter(os.Stdout)
	}

	if writer.SetComment("goadifdedupe\n") != nil {
		fmt.Fprint(os.Stderr, err)
		return
	}

	// For not deduping, use this filter API:
	// reader := adifparser.NewADIFReader(fp)

	// WITH deduping
	reader := adifparser.NewDedupeADIFReader(fp)

	for record, err := reader.ReadRecord(); record != nil || err != nil; record, err = reader.ReadRecord() {
		if err != nil {
			if err != io.EOF {
				fmt.Fprint(os.Stderr, err)
			}
			break // when io.EOF break the loop!
		}

		// process things here with the record
		writer.WriteRecord(record)

	}

	// Flush and close the output
	writer.Flush()
	if writefp != os.Stdout {
		writefp.Close()
	}
	fmt.Fprintf(os.Stderr, "Total records: %d\n", reader.RecordCount())
}
