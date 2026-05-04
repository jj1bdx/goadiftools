// goadifdxcccl: add DXCC/CQ Zone info with Club Log database reference
// by Kenji Rikitake, JJ1BDX
// Usage: goadifdxcc [-f infile] [-o outfile]

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jj1bdx/adifparser"
	"github.com/jj1bdx/gocldb"
)

func main() {
	var infile = flag.String("f", "", "input file (stdout in none)")
	var outfile = flag.String("o", "", "output file (stdout if none)")

	var fp *os.File
	var err error

	flag.Usage = func() {
		execname := os.Args[0]
		fmt.Fprintln(flag.CommandLine.Output(),
			"goadifdxcc: add DXCC/CQ Zone fields using gocldb")
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [-f infile] [-o outfile]\n", execname)
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		fmt.Fprintf(flag.CommandLine.Output(),
			"How goadifdxcc works:\n"+
				"For each record, fetch the corresponding local Club Log database data\n"+
				"with the content of the ADIF field 'call'.\n"+
				"Then for each ADIF field of country, cqz, cont, dxcc:\n"+
				"fill in the field with the DXCC database data if the field is empty.\n")
		fmt.Fprintf(flag.CommandLine.Output(),
			"Note well:\n"+
				"gocldb does not handle ITU Zone info\n")
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

	// Initialize gocldb
	gocldb.LoadCtyXml()
	// Disable debug mode logging of gocldb
	gocldb.SetDebugOutput(io.Discard)

	if cerr := writer.SetComment("goadifdxcccl\n"); cerr != nil {
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
		// Get time entry from QSO_DATE and TIME_ON fields
		adifdate, err := record.GetValue("qso_date")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		adiftime, err := record.GetValue("time_on")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		// Validate qso_date / time_on lengths before slicing.
		// ADIF requires QSO_DATE to be 8 digits and TIME_ON to be
		// 4 or 6 digits; without this guard a truncated value would
		// trigger an index-out-of-range panic mid-stream.
		if len(adifdate) < 8 || len(adiftime) < 4 {
			fmt.Fprintf(os.Stderr,
				"malformed qso_date/time_on for %s: %q %q\n",
				call, adifdate, adiftime)
			continue
		}

		adifyear, err := strconv.Atoi(adifdate[0:4])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		adifmonth, err := strconv.Atoi(adifdate[4:6])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		adifday, err := strconv.Atoi(adifdate[6:8])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		adifhour, err := strconv.Atoi(adiftime[0:2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		adifminute, err := strconv.Atoi(adiftime[2:4])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		// Seconds are optional per ADIF (TIME_ON is HHMM or HHMMSS).
		// The original guard "> 4" admitted len == 5 and panicked on
		// adiftime[4:6]; require the full HHMMSS form before slicing,
		// and surface a parse failure instead of silently zeroing.
		adifsecond := 0
		if len(adiftime) >= 6 {
			adifsecond, err = strconv.Atoi(adiftime[4:6])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
		}
		recordtime := time.Date(
			adifyear, time.Month(adifmonth), adifday,
			adifhour, adifminute, adifsecond,
			0, time.UTC)

		// Fetch DXCC database data
		result, err := gocldb.CheckCallsign(strings.ToUpper(call), recordtime)
		if err == nil {

			// For each ADIF field of country, cqz, cont, dxcc:
			// If each field is empty,
			// fill in the field with the DXCC database data
			// If already filled, do nothing
			_, err = record.GetValue("country")
			if err == adifparser.ErrNoSuchField {
				record.SetValue("country", result.Name)
			}
			_, err = record.GetValue("cqz")
			// Do not set CQZ field if the obtained value is zero
			// CQZ value must be a positive integer
			if err == adifparser.ErrNoSuchField {
				if result.Cqz > 0 {
					record.SetValue("cqz", strconv.Itoa(int(result.Cqz)))
				}
			}
			_, err = record.GetValue("cont")
			if err == adifparser.ErrNoSuchField {
				if len(result.Cont) > 0 {
					record.SetValue("cont", result.Cont)
				}
			}
			_, err = record.GetValue("dxcc")
			if err == adifparser.ErrNoSuchField {
				record.SetValue("dxcc", strconv.Itoa(int(result.Adif)))
			}
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
