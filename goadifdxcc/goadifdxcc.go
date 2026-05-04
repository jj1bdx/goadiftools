// goadifdxcc: add DXCC related fields
// by Kenji Rikitake, JJ1BDX
// Usage: goadifdxcc [-f infile] [-o outfile]

package main

import (
	"flag"
	"fmt"
	"github.com/jj1bdx/adifparser"
	"github.com/jj1bdx/godxcc"
	"io"
	"os"
	"strconv"
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
			"goadifdxcc: add DXCC related fields using godxcc")
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [-f infile] [-o outfile]\n", execname)
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		fmt.Fprintf(flag.CommandLine.Output(),
			"How goadifdxcc works:\n"+
				"For each record, fetch the correspoding DXCC database data\n"+
				"with the content of the ADIF field 'call'.\n"+
				"Then for each ADIF field of country, cqz, ituz, cont, dxcc:\n"+
				"fill in the field with the DXCC database data if the field is empty.\n")
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

	// Initialize godxcc
	godxcc.LoadCty()

	if cerr := writer.SetComment("goadifdxcc\n"); cerr != nil {
		fmt.Fprintln(os.Stderr, cerr)
		os.Exit(1)
	}

	reader := adifparser.NewADIFReader(fp)
	for record, err := reader.ReadRecord(); record != nil || err != nil; record, err = reader.ReadRecord() {
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, err)
			}
			break // when io.EOF break the loop!
		}

		// Get callsign entry
		call, err := record.GetValue("call")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		// Fetch DXCC database data
		dxccdata := godxcc.DXCCGetRecord(strings.ToUpper(call))

		// For each ADIF field of country, cqz, ituz, cont, dxcc:
		// fill in the field with the DXCC database data if the fieldis empty
		// If already filled, do nothing
		_, err = record.GetValue("country")
		if err == adifparser.ErrNoSuchField {
			record.SetValue("country", dxccdata.Waecountry)
		}
		_, err = record.GetValue("cqz")
		if err == adifparser.ErrNoSuchField {
			record.SetValue("cqz", strconv.Itoa(dxccdata.Waz))
		}
		_, err = record.GetValue("ituz")
		if err == adifparser.ErrNoSuchField {
			record.SetValue("ituz", strconv.Itoa(dxccdata.Ituz))
		}
		_, err = record.GetValue("cont")
		if err == adifparser.ErrNoSuchField {
			record.SetValue("cont", dxccdata.Cont)
		}
		_, err = record.GetValue("dxcc")
		if err == adifparser.ErrNoSuchField {
			record.SetValue("dxcc", strconv.Itoa(dxccdata.Entitycode))
		}

		// Write the record
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
