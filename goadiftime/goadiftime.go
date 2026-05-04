// goadiftime: sort and filter ADIF file by time
// by Kenji Rikitake, JJ1BDX
// Usage: goadiftime [-f infile] [-o outfile] [-r]
//        [-starttime RFC3339-time] [-endtime RFC3339-time]
// RFC3339-time example: 2022-10-11T12:33:45Z
// Time of ADIF record determined by: qso_date and time_on
//
// Time filtering conditions:
// if starttime and endtime both are specified:
// the condition is: starttime <= record time <= endtime
// if only starttime is specified:
// the condition is: starttime <= record time
// if only endtime is specified:
// the condition is: record time <= endtime
//
// Sorting conditions:
// when with -n option or -n=true:
//   the output is not sorted
// when without -n option or -n=false (default):
//   when without -r option or -r=false (default):
//   the output is sorted by time increasing order
//   when with -r option or -r=true:
//   the output is sorted by time decreasing order

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/jj1bdx/adifparser"
	"io"
	"os"
	"sort"
	"strconv"
	"time"
)

type recordWithTime struct {
	date   time.Time
	record adifparser.ADIFRecord
}

func main() {
	var infile = flag.String("f", "", "input file (stdout in none)")
	var outfile = flag.String("o", "", "output file (stdout if none)")
	var reverse bool
	flag.BoolVar(&reverse, "r", false, "reverse sort (new to old)")
	var nosorting bool
	flag.BoolVar(&nosorting, "n", false, "no sorting with this flag")
	var starttime = flag.String("starttime", "", "start time in RFC3339")
	var endtime = flag.String("endtime", "", "end time in RFC3339")

	var fp *os.File
	var err error

	flag.Usage = func() {
		execname := os.Args[0]
		fmt.Fprintln(flag.CommandLine.Output(),
			"goadiftime: sort and filter ADIF file by time")
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s  [-f infile] [-o outfile] [-r] "+
				"[-starttime RFC3339-time] [-endtime RFC3339-time]\n",
			execname)
		flag.PrintDefaults()
		details :=
			"RFC3339-time example: 2022-10-11T12:33:45Z\n" +
				"Time of ADIF record determined by: qso_date and time_on\n" +
				"\n" +
				"Time filtering conditions:\n" +
				"if starttime and endtime both are specified:\n" +
				"the condition is: starttime <= record time <= endtime\n" +
				"if only starttime is specified:\n" +
				"the condition is: starttime <= record time\n" +
				"if only endtime is specified:\n" +
				"the condition is: record time <= endtime\n" +
				"\n" +
				"Sorting conditions:\n" +
				"when with -n option or -n=true:\n" +
				"  the output is not sorted\n" +
				"when without -n option or -n=false (default):\n" +
				"  when without -r option or -r=false (default):\n" +
				"  the output is sorted by time increasing order\n" +
				"  when with -r option or -r=true:\n" +
				"  the output is sorted by time decreasing order\n"
		fmt.Fprint(flag.CommandLine.Output(), details)
	}

	records := []recordWithTime{}

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

	var startTime time.Time
	var endTime time.Time
	starttimeexists := *starttime != ""
	if starttimeexists {
		parsedStartTime, err := time.Parse(time.RFC3339, *starttime)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		startTime = parsedStartTime.UTC()
	}

	endtimeexists := *endtime != ""
	if endtimeexists {
		parsedEndTime, err := time.Parse(time.RFC3339, *endtime)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		endTime = parsedEndTime.UTC()
	}
	if starttimeexists && endtimeexists &&
		startTime.After(endTime) {
		fmt.Fprintln(os.Stderr, errors.New("starttime is after endtime"))
		os.Exit(2)
	}

	if cerr := writer.SetComment("goadiftime\n"); cerr != nil {
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

		// Validate qso_date / time_on lengths before slicing to avoid
		// an index-out-of-range panic on a truncated/malformed record.
		if len(adifdate) < 8 || len(adiftime) < 4 {
			fmt.Fprintf(os.Stderr,
				"malformed qso_date/time_on: %q %q\n",
				adifdate, adiftime)
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
		// Seconds are optional (HHMM or HHMMSS). The original ">4"
		// guard admitted len==5 and panicked at adiftime[4:6]; require
		// the full HHMMSS form, and surface a parse error rather than
		// silently zeroing — silent timestamp corruption is a
		// correctness hazard for a time-filter tool.
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

		passstart := !starttimeexists ||
			(recordtime.After(startTime) || recordtime.Equal(startTime))
		passend := !endtimeexists ||
			(recordtime.Before(endTime) || recordtime.Equal(endTime))
		if passstart && passend {
			if nosorting {
				// Streaming fast path: when the user opts out of
				// sorting, write directly and never accumulate the
				// slice — bounded memory regardless of input size.
				if werr := writer.WriteRecord(record); werr != nil {
					fmt.Fprintln(os.Stderr, werr)
					os.Exit(1)
				}
			} else {
				recordandtime := recordWithTime{recordtime, record}
				records = append(records, recordandtime)
			}
		}
	}

	if !nosorting {
		if reverse {
			sort.Slice(records,
				func(i, j int) bool {
					return records[i].date.After(records[j].date)
				})
		} else {
			sort.Slice(records,
				func(i, j int) bool {
					return records[i].date.Before(records[j].date)
				})
		}
		for i := range records {
			if werr := writer.WriteRecord(records[i].record); werr != nil {
				fmt.Fprintln(os.Stderr, werr)
				os.Exit(1)
			}
		}
	}

	if ferr := writer.Flush(); ferr != nil {
		fmt.Fprintln(os.Stderr, ferr)
		os.Exit(1)
	}

}
