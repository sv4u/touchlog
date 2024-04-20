package src

import (
	"flag"
	"fmt"
)

const author string = "Sasank 'squatch$' Vishnubhatla"
const version string = "1.0-dev"

const log_format string = "> month: %s\n> day: %s\n> year: %s\n\n|> events\n\n|> emotions\n\n|> things to remember\n"

func Touchlog(buildTime string) bool {
	return read_args(buildTime)
}

func read_args(buildTime string) bool {
	datePtr := flag.String("date", "mmddyyyy", "a logfile is created with the supplied date")
	outDirPtr := flag.String("outdir", "dir", "write the logfile to inputted directory")
	versionPtr := flag.Bool("version", false, "display the version information")

	flag.Parse()

	if *versionPtr {
		// print version information
		fmt.Println("touchlog")
		fmt.Println("Author: ", author)
		fmt.Println("Version: ", version)
		fmt.Println("Build: ", buildTime)

		return true
	}

	if *outDirPtr == "" {
		// TODO print error message

		return false
	}

	handle_date(datePtr)
	normalize_outdir(outDirPtr)
	write_log(datePtr, outDirPtr)

	return true
}

func handle_date(datePtr *string) {
	if *datePtr == "" {
		// TODO handle using today's date
	} else {
		// TODO parsing date from string
	}
}

func normalize_outdir(outDirPtr *string) {}

func write_log(datePtr *string, outDirPtr *string) bool {
	return true
}
