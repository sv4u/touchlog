# touchlog

A tool to create log files for a date.

## Usage

The `,touchlog` executable has the following options:

- 'noop': a logfile is create with the current date
- '-d mmmddyyyy': a logfile is create with the supplied date
- '-f [dir]': write the logfile to the existing directory given
- '-v': the version information is displayed
- '-h': the help message is displayed

## Installation

Place the `,touchlog` executable into a directory in your `PATH`. Or, you can choose down download the [`install.sh`](./src/install.sh) script and allow `,touchlog` to be installed to `/usr/local/bin`.

## Man Page

The man page for `,touchlog` can be found here: [,touchlog.1](docs/,touchlog.1.md). Please either manually install it to your system or use the [`install-doc.sh`](./src/install-doc.sh) script to install both `,touchlog` and the man page.

## Source

To install the source, please use the [`install-source.sh`](./src/install-source.sh) script.

## License

See the [License](LICENSE).

