package src

import (
	"flag"
)

// TODO write touchlog static string

func read_args() {
	datePtr := flag.String("date", "mmddyyyy", "a logfile is created with the supplied date")
	outDirPtr := flag.String("outdir", "dir", "write the logfile to inputted directory")
	versionPtr := flag.Bool("version", false, "display the version information")

	flag.Parse()

	if *versionPtr {
		// TODO print version information
		// exit afterwards
	}

	if *outDirPtr == "" {
		// TODO print error message
		// exit afterwards
	}

	handle_date(datePtr)
	normalize_outdir(outDirPtr)

}

func handle_date(datePtr *string) {}

func normalize_outdir(outDirPtr *string) {}
