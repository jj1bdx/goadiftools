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
			fmt.Fprint(os.Stderr, err)
			return
		}
	}

	var writer adifparser.ADIFWriter
	var writefp *os.File
	if *outfile != "" {
		writefp, err = os.Create(*outfile)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
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
			fmt.Fprint(os.Stderr, err)
			return
		}
		startTime = parsedStartTime.UTC()
	}

	endtimeexists := *endtime != ""
	if endtimeexists {
		parsedEndTime, err := time.Parse(time.RFC3339, *endtime)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
		endTime = parsedEndTime.UTC()
	}
	if starttimeexists && endtimeexists &&
		startTime.After(endTime) {
		fmt.Fprint(os.Stderr, errors.New("starttime is after endtime"))
		return
	}

	if writer.SetComment("goadiftime\n") != nil {
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

		passstart := !starttimeexists ||
			(recordtime.After(startTime) || recordtime.Equal(startTime))
		passend := !endtimeexists ||
			(recordtime.Before(endTime) || recordtime.Equal(endTime))
		if passstart && passend {
			recordandtime := recordWithTime{recordtime, record}
			records = append(records, recordandtime)
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
	}

	for i := range records {
		writer.WriteRecord(records[i].record)
	}

	// Flush and close output here
	writer.Flush()
	if writefp != os.Stdout {
		writefp.Close()
	}

}
