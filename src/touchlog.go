package src

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"
)

const author string = "Sasank 'squatch$' Vishnubhatla"
const version string = "1.0-dev"

const log_format string = "> month: %s\n> day: %s\n> year: %s\n\n|> events\n\n|> emotions\n\n|> things to remember\n"

const debug_flags int = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile | log.Lmsgprefix

var verbosity bool
var buf bytes.Buffer
var debug = log.New(&buf, "touchlog-verbose > ", debug_flags)
var errlog = log.New(&buf, "touchlog-error > ", debug_flags)
var print = log.New(&buf, "", 0)

func vprintf(format string, v ...any) {
	if verbosity {
		debug.Printf(format, v...)
	}
}

func vprintln(a ...any) {
	if verbosity {
		debug.Println(a...)
	}
}

func eprint(v ...any) {
	errlog.Fatal(v...)
}

func eprintf(format string, v ...any) {
	errlog.Fatalf(format, v...)
}

func eprintln(a ...any) {
	errlog.Fatalln(a...)
}

func println(a ...any) {
	print.Println(a...)
}

func Touchlog(buildTime string) bool {
	defer fmt.Print(&buf)
	return read_args(buildTime)
}

func read_args(buildTime string) bool {
	datePtr := flag.String("date", "", "a logfile is created with the supplied date")
	outDirPtr := flag.String("outdir", "", "write the logfile to inputted directory")
	versionPtr := flag.Bool("version", false, "display the version information")
	verbosePtr := flag.Bool("verbose", false, "enable verbosity mode")

	flag.Parse()

	// store the verbosity setting
	verbosity = *verbosePtr

	if *versionPtr {
		vprintln("printing version information")

		// print version information
		println("touchlog")
		println("Author:  ", author)
		println("Version: ", version)
		println("Build:   ", buildTime)

		return true
	}

	vprintln("checking output directory")
	if *outDirPtr == "" {
		vprintf("%p points to empty string\n", outDirPtr)
		eprintln("output directory is empty")

		return false
	}

	handle_date(datePtr)
	normalize_outdir(outDirPtr)

	filename := *datePtr + ".log"

	vprintf("filename to use: %s", filename)
	vprintf("normalized outdir: %s", *outDirPtr)

	write_log(filename, outDirPtr)

	return true
}

func pad(val int, length int) string {
	vprintf("pad(%v, %v)\n", val, length)

	str := strconv.Itoa(val)

	for len(str) < length {
		str = "0" + str
	}

	return str
}

func handle_date(datePtr *string) bool {
	tmp := *datePtr

	vprintf("handle_date(%p)\n", datePtr)
	vprintf("handle_date(%s)\n", tmp)

	if *datePtr == "" {
		// using today's date
		date := time.Now()
		year := pad(date.Year(), 4)
		month := pad(int(date.Month()), 2)
		day := pad(date.Day(), 2)

		tmp = month + "-" + day + "-" + year
	} else {
		// parsing date from string
		// expected format: mmddyyyy
		// expected length: 8
		if len(tmp) != 8 {
			eprintf("invalid input date: %s\n", tmp)
			eprintln("expected format: mmddyyyy")
			eprintln("expected length: 8")

			return false
		}

		// assumption: length of tmp is 8
		// start slicing
		_month := tmp[0:2]
		_day := tmp[2:4]
		_year := tmp[4:8]

		vprintf("parsing month: %s", _month)
		vprintf("parsing day: %s", _day)
		vprintf("parsing year: %s", _year)

		vprintln("making sure month, day, and year are numbers")

		__month, err := strconv.Atoi(_month)
		if err != nil {
			eprint(err)
		}

		__day, err := strconv.Atoi(_day)
		if err != nil {
			eprint(err)
		}

		__year, err := strconv.Atoi(_year)
		if err != nil {
			eprint(err)
		}

		month := pad(__month, 2)
		day := pad(__day, 2)
		year := pad(__year, 4)

		tmp = month + "-" + day + "-" + year
	}

	*datePtr = tmp

	vprintf("handle_date -> %p\n", datePtr)
	vprintf("handle_date -> %s\n", tmp)

	return true
}

func normalize_outdir(outDirPtr *string) {
	tmp := *outDirPtr

	vprintf("normalize_outdir(%p)\n", outDirPtr)
	vprintf("normalize_outdir(%s)\n", tmp)

	// join on nothing to clean up path
	tmpPath := filepath.Join(tmp)
	tmpPath, err := filepath.Abs(tmpPath)
	if err != nil {
		eprintf("could not get absolute path of outdir %s", tmpPath)
	}

	vprintf("tmpPath := %s", tmpPath)

	// store back in outDirPtr to reuse pointer
	*outDirPtr = tmpPath
}

func write_log(filename string, outDirPtr *string) bool {
	return true
}
