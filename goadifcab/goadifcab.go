// goadifcab: output Cabrillo QSO log entries for given ADIF records
// by Kenji Rikitake, JJ1BDX
// Usage: goadifcab [-f infile] [-o outfile]
// Note: for HF (160m - 10m) contests only
// Required ADIF fields:
//  station_callsign, call, band, mode,
//  qso_date, time_on, rst_sent, rst_rcvd,
//  stx_string, srx_string
// Optional ADIF fields:
//  freq: will be parsed and reflected

package main

import (
	"flag"
	"fmt"
	"github.com/jj1bdx/adifparser"
	"io"
	"os"
	"strconv"
)

func main() {
	var infile = flag.String("f", "", "input file (stdout in none)")
	var outfile = flag.String("o", "", "output file (stdout if none)")

	var fp *os.File
	var err error

	flag.Usage = func() {
		execname := os.Args[0]
		fmt.Fprintln(flag.CommandLine.Output(),
			"goadifcab: output Cabrillo QSO log entries for given ADIF records")
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [-f infile] [-o outfile]\n", execname)
		fmt.Fprintf(flag.CommandLine.Output(),
			"Note: for HF (160m - 10m) contests only\n")
		fmt.Fprintf(flag.CommandLine.Output(),
			"Required ADIF fields:\n")
		fmt.Fprintf(flag.CommandLine.Output(),
			" station_callsign, call, band, mode,\n")
		fmt.Fprintf(flag.CommandLine.Output(),
			" qso_date, time_on, rst_sent, rst_rcvd,\n")
		fmt.Fprintf(flag.CommandLine.Output(),
			" stx_string, srx_string\n")
		fmt.Fprintf(flag.CommandLine.Output(),
			"Optional ADIF fields:\n")
		fmt.Fprintf(flag.CommandLine.Output(),
			" freq: will be parsed and reflected\n")
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

	var writer io.Writer
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
		writer = io.Writer(writefp)
	} else {
		writefp = nil
		writer = io.Writer(os.Stdout)
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

		// Get station_callsign entry
		station_callsign, err := record.GetValue("station_callsign")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		// Get call entry
		call, err := record.GetValue("call")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		// Get band entry
		band, err := record.GetValue("band")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		// Get band entry
		mode, err := record.GetValue("mode")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		// Get freq entry
		// If not existed, leave it as null string
		freq, err := record.GetValue("freq")
		if err == adifparser.ErrNoSuchField {
			freq = ""
		} else if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		// Get time_on and qso_date entries
		adifdate, err := record.GetValue("qso_date")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		adiftime, err := record.GetValue("time_on")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		// Get rst_sent entry
		rst_sent, err := record.GetValue("rst_sent")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		// Get rst_rcvd entry
		rst_rcvd, err := record.GetValue("rst_rcvd")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		// Get stx_string entry
		stx_string, err := record.GetValue("stx_string")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		// Get srx_string entry
		srx_string, err := record.GetValue("srx_string")
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}

		// Convert qso time to partial data strings
		adifyear, err := strconv.Atoi(adifdate[0:4])
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		adifmonth, err := strconv.Atoi(adifdate[4:6])
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		adifday, err := strconv.Atoi(adifdate[6:8])
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		adifhour, err := strconv.Atoi(adiftime[0:2])
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}
		adifminute, err := strconv.Atoi(adiftime[2:4])
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			continue
		}

		var freqnum uint
		// Convert band to base freq
		// Note: for contests only, exclude WARC bands
		switch {
		case band == "10m":
			freqnum = 28000
		case band == "15m":
			freqnum = 21000
		case band == "20m":
			freqnum = 14000
		case band == "40m":
			freqnum = 7000
		case band == "80m":
			freqnum = 3500
		case band == "160m":
			freqnum = 1800
		default:
			fmt.Fprint(os.Stderr, "Unknown band\n")
			continue
		}
		if freq != "" {
			freqval, err := strconv.ParseFloat(freq, 64)
			if err != nil {
				fmt.Fprint(os.Stderr, err)
				continue
			}
			freqnum = uint(freqval * 1000)
		}

		var cabmode string
		// Convert mode to cabrillo mode
		// Note: for contests only
		switch {
		case mode == "CW":
			cabmode = "CW"
		case mode == "SSB":
			cabmode = "PH"
		case mode == "FM":
			cabmode = "FM"
		case mode == "RTTY":
			cabmode = "RY"
		case mode == "FT8":
			cabmode = "DG"
		case mode == "MFSK":
			cabmode = "DG"
		default:
			fmt.Fprint(os.Stderr, "Unknown mode\n")
			continue
		}

		// print output record
		fmt.Fprintf(writer, "QSO: %5d %s ", freqnum, cabmode)
		fmt.Fprintf(writer, "%04d-%02d-%02d %02d%02d ",
			adifyear, adifmonth, adifday, adifhour, adifminute)
		fmt.Fprintf(writer, "%-13s %-3s %-6s %-13s %-3s %-6s\n",
			station_callsign, rst_sent, stx_string,
			call, rst_rcvd, srx_string)
	}

	// Flush and close the output
	// writer.Flush()
	if writefp != os.Stdout {
		writefp.Close()
	}
}
