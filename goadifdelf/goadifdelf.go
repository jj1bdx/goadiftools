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

	fieldstodelete := flag.Args()

	if cerr := writer.SetComment("goadifdelf\n"); cerr != nil {
		fmt.Fprintln(os.Stderr, cerr)
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

		// Delete specified fields
		for i := range fieldstodelete {
			// DeleteField return value intentionally ignored: deleting a
			// missing field is a no-op and not an error here.
			record.DeleteField(strings.ToLower(fieldstodelete[i]))
		}
		if werr := writer.WriteRecord(record); werr != nil {
			fmt.Fprintln(os.Stderr, werr)
			os.Exit(1)
		}

	}

	if ferr := writer.Flush(); ferr != nil {
		fmt.Fprintln(os.Stderr, ferr)
		os.Exit(1)
	}
}
