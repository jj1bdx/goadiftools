// goadifdelf: remove specified ADIF fields
// by Kenji Rikitake, JJ1BDX
// Usage: goadifdelf [-f infile] [-o outfile] field_names...

package main

import (
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
			"goadifdelf: remove specified ADIF fields")
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [-f infile] [-o outfile] field_names...\n", execname)
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

	fieldstodelete := flag.Args()

	if writer.SetComment("goadifdelf\n") != nil {
		fmt.Fprint(os.Stderr, err)
		return
	}

	// For deduping, use this filter API:
	// reader := adifparser.NewDedupeADIFReader(fp)

	reader := adifparser.NewADIFReader(fp)
	for record, err := reader.ReadRecord(); record != nil || err != nil; record, err = reader.ReadRecord() {
		if err != nil {
			if err != io.EOF {
				fmt.Fprint(os.Stderr, err)
			}
			break // when io.EOF break the loop!
		}

		// Delete specified fields
		for i := range fieldstodelete {
			// Do not use retuen values
			record.DeleteField(strings.ToLower(fieldstodelete[i]))
		}
		writer.WriteRecord(record)

	}

	// Flush and close the output
	writer.Flush()
	if writefp != os.Stdout {
		writefp.Close()
	}
}
