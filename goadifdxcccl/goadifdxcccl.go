// goadifdxcccl: add DXCC/CQ Zone info with Club Log database reference
// by Kenji Rikitake, JJ1BDX
// Usage: goadifdxcc [-f infile] [-o outfile]

package main

import (
	"flag"
	"fmt"
	"github.com/jj1bdx/adifparser"
	"github.com/jj1bdx/gocldb"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
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
				"For each record, fetch the correspoding local Club Log database data\n"+
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

	// Initialize gocldb
	gocldb.LoadCtyXml()
	// Disable debug mode logging of gocldb
	gocldb.DebugLogger.SetOutput(io.Discard)

	if writer.SetComment("goadifdxcccl\n") != nil {
		fmt.Fprint(os.Stderr, err)
		return
	}

	reader := adifparser.NewADIFReader(fp)
	for record, err := reader.ReadRecord(); record != nil || err != nil; record, err = reader.ReadRecord() {
		if err != nil {
			if err != io.EOF {
				fmt.Fprint(os.Stderr, err)
			}
			break // when io.EOF break the loop!
		}

		// Get callsign entry
		call, err := record.GetValue("call")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		// Get time entry from QSO_DATE and TIME_ON fields
		adifdate, err := record.GetValue("qso_date")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
		adiftime, err := record.GetValue("time_on")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		adifyear, err := strconv.Atoi(adifdate[0:4])
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
		adifmonth, err := strconv.Atoi(adifdate[4:6])
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
		adifday, err := strconv.Atoi(adifdate[6:8])
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
		adifhour, err := strconv.Atoi(adiftime[0:2])
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
		adifminute, err := strconv.Atoi(adiftime[2:4])
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
		adifsecond := 0
		if len(adiftime) > 4 {
			adifsecond, err = strconv.Atoi(adiftime[4:6])
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
			if err == adifparser.ErrNoSuchField {
				record.SetValue("cqz", strconv.Itoa(int(result.Cqz)))
			}
			_, err = record.GetValue("cont")
			if err == adifparser.ErrNoSuchField {
				record.SetValue("cont", result.Cont)
			}
			_, err = record.GetValue("dxcc")
			if err == adifparser.ErrNoSuchField {
				record.SetValue("dxcc", strconv.Itoa(int(result.Adif)))
			}
		}

		// Write the record
		writer.WriteRecord(record)

	}

	// Flush and close the output
	writer.Flush()
	if writefp != os.Stdout {
		writefp.Close()
	}
}
