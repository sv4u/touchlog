# touchlog

A tool to create log files for a date.

## Usage

The `touchlog` executable has the following options:

- '-date mmmddyyyy': a logfile is create with the supplied date
- '-outdir [dir]': write the logfile to inputted directory
- '-verbose': enable verbosity mode
- '-version': display the version information
- '-help': the help message is displayed

## Installation

Install via go module:

```bash
go get -v gitlab.com/sv4u/touchlog
```

Install source:

TODO

## Man Page

The man page for `touchlog` can be found here: [touchlog.1](touchlog.1.html). Please manually install touchlog to your system to use both `touchlog` and the man page.

## Changelog

To generate a changelog, use [`git-chglog`](https://github.com/git-chglog/git-chglog/). Follow this command:

```bash
git-chglog -o CHANGELOG.md
```

## License

See the [License](LICENSE).
