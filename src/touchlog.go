package src

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"path/filepath"
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

	vprintf("date to write: %s", *datePtr)
	vprintf("normalized outdir: %s", *outDirPtr)

	write_log(datePtr, outDirPtr)

	return true
}

func handle_date(datePtr *string) {
	tmp := *datePtr

	vprintf("handle_date(%p)\n", datePtr)
	vprintf("handle_date(%s)\n", tmp)

	if *datePtr == "" {
		// TODO handle using today's date
	} else {
		// TODO parsing date from string
	}
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

func write_log(datePtr *string, outDirPtr *string) bool {
	return true
}
