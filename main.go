package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const author string = "Sasank 'squatch$' Vishnubhatla"

const log_format string = "> month: %v\n> day: %v\n> year: %v\n\n|> events\n\n|> emotions\n\n|> things to remember\n"

const debug_flags int = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile | log.Lmsgprefix

var buildTime string
var version string

func main() {
	Touchlog()
}

var verbosity bool
var buf bytes.Buffer
var nilbuf bytes.Buffer
var debug = log.New(&nilbuf, "touchlog-verbose > ", debug_flags)
var errlog = log.New(&buf, "touchlog-error > ", debug_flags)
var print = log.New(&buf, "", 0)

// Touchlog parses the user input from the command line and then creates a logfile for the desired
// date.
func Touchlog() bool {
	defer fmt.Print(&buf)
	datePtr := flag.String("date", "", "a logfile is created with the supplied date")
	outDirPtr := flag.String("outdir", "", "write the logfile to inputted directory")
	versionPtr := flag.Bool("version", false, "display the version information")
	verbosePtr := flag.Bool("verbose", false, "enable verbosity mode")

	flag.Parse()

	// store the verbosity setting
	verbosity = *verbosePtr
	if verbosity {
		debug = log.New(&buf, "touchlog-verbose > ", debug_flags)
	}

	if *versionPtr {
		debug.Println("printing version information")

		// print version information
		print.Println("touchlog")
		print.Println("Author:  ", author)
		print.Println("Version: ", version)
		print.Println("Build:   ", buildTime)

		return true
	}

	debug.Println("checking output directory")
	if *outDirPtr == "" {
		debug.Printf("%p points to empty string\n", outDirPtr)
		errlog.Println("output directory is empty")

		return false
	}

	month, day, year, result := Handle_date(datePtr)
	if !result {
		return false
	}

	debug.Printf("mmddyyyy -> %v%v%v\n", month, day, year)

	result = Normalize_outdir(outDirPtr)
	if !result {
		return false
	}

	filename := *datePtr + ".log"

	debug.Printf("filename to use: %s", filename)
	debug.Printf("normalized outdir: %s", *outDirPtr)

	Write_log(filename, outDirPtr, month, day, year)

	return true
}

func pad(val int, length int) string {
	debug.Printf("pad(%v, %v)\n", val, length)

	str := strconv.Itoa(val)

	for len(str) < length {
		str = "0" + str
	}

	debug.Printf("padded %v to %v length -> %s", val, length, str)

	return str
}

// Handle_date takes a potential date input in the form of mmddyyyy and parses it into the proper
// month day and year strings.
//
// Once processing is complete, Handle_date returns month, day, year, true.
// If an error occurs during processing, Handle_date logs the errors and returns "", "", "", false.
func Handle_date(datePtr *string) (month string, day string, year string, success bool) {
	tmp := *datePtr

	debug.Printf("handle_date(%p)\n", datePtr)
	debug.Printf("handle_date(%s)\n", tmp)

	if *datePtr == "" {
		// using today's date
		date := time.Now()

		debug.Printf("Using today's date: %v\n", date)

		year = pad(date.Year(), 4)
		month = pad(int(date.Month()), 2)
		day = pad(date.Day(), 2)

		tmp = month + "-" + day + "-" + year
	} else {
		// parsing date from string
		// expected format: mmddyyyy
		// expected length: 8
		if len(tmp) != 8 {
			errlog.Printf("invalid input date: %s\n", tmp)
			errlog.Println("expected format: mmddyyyy")
			errlog.Println("expected length: 8")

			success = false

			return
		}

		// assumption: length of tmp is 8
		// start slicing
		_month := tmp[0:2]
		_day := tmp[2:4]
		_year := tmp[4:8]

		debug.Printf("parsing month: %s\n", _month)
		debug.Printf("parsing day: %s\n", _day)
		debug.Printf("parsing year: %s\n", _year)

		debug.Println("making sure month, day, and year are numbers")

		__month, err := strconv.Atoi(_month)
		if err != nil {
			errlog.Print(err)

			success = false

			return
		}

		__day, err := strconv.Atoi(_day)
		if err != nil {
			errlog.Print(err)

			success = false

			return
		}

		__year, err := strconv.Atoi(_year)
		if err != nil {
			errlog.Print(err)

			success = false

			return
		}

		month = pad(__month, 2)
		day = pad(__day, 2)
		year = pad(__year, 4)

		tmp = month + "-" + day + "-" + year
	}

	*datePtr = tmp

	debug.Printf("handle_date -> %p\n", datePtr)
	debug.Printf("handle_date -> %s\n", tmp)

	success = true

	return
}

// Normlaize_outdir takes a pointer to a string representing a directory and noramlizes it so that
// writing to the outputted directory can be seamless.
//
// If the inputted directory is successfully normalized, Normlaize_outdir returns true.
// Otherwise, the error is logged and Normalize_outdir returns false.
func Normalize_outdir(outDirPtr *string) bool {
	tmp := *outDirPtr

	debug.Printf("normalize_outdir(%p)\n", outDirPtr)
	debug.Printf("normalize_outdir(%s)\n", tmp)

	// join on nothing to clean up path
	tmpPath := filepath.Join(tmp)
	tmpPath, err := filepath.Abs(tmpPath)
	if err != nil {
		errlog.Print(err)

		return false
	}

	debug.Printf("tmpPath := %s", tmpPath)

	// store back in outDirPtr to reuse pointer
	*outDirPtr = tmpPath

	return true
}

// Write_log takes a filename, a pointer to a string representing a directory, the month, the day,
// the year, writes a logfile to the requested directory.
//
// If the logfile is successfully written, Write_log returns true.
// Otherwise, the error is logged and Write_log returns false.
func Write_log(filename string, outDirPtr *string, month string, day string, year string) bool {
	debug.Printf("write_log(%v, %v)\n", filename, outDirPtr)

	logfile := filepath.Join(*outDirPtr, filename)
	f, err := os.Create(logfile)
	if err != nil {
		errlog.Print(err)

		return false
	}

	debug.Printf("os.Create(%v) -> %v", logfile, f)

	defer f.Close()
	debug.Printf("defer %v.Close()", f)

	log_data := fmt.Sprintf(log_format, month, day, year)

	n, err := f.WriteString(log_data)
	if err != nil {
		errlog.Print(err)

		return false
	}

	debug.Printf("wrote %d bytes\n", n)

	err = f.Sync()
	if err != nil {
		errlog.Print(err)

		return false
	}

	return true
}
