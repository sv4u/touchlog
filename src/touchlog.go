package src

import (
	"flag"
	"fmt"
	"os"
)

const author string = "Sasank 'squatch$' Vishnubhatla"
const version string = "1.0-dev"

const log_format string = "> month: %s\n> day: %s\n> year: %s\n\n|> events\n\n|> emotions\n\n|> things to remember\n"

func Read_Args(buildTime string) bool {
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

		os.Exit(0)
	}

	if *outDirPtr == "" {
		// TODO print error message
		// exit afterwards
	}

	handle_date(datePtr)
	normalize_outdir(outDirPtr)

	return true
}

func handle_date(datePtr *string) {}

func normalize_outdir(outDirPtr *string) {}
