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
go install github.com/sv4u/touchlog@latest
```

Install from tarball:

- From [development.sasankvishnubhatla.net/log-suite/touchlog](https://development.sasankvishnubhatla.net/log-suite/touchlog) download the [`touchlog-latest.tar`](https://development.sasankvishnubhatla.net/log-suite/touchlog/touchlog-latest.tar)
- Extract `touchlog-latest.tar` to `/usr/bin/local/touchlog`
- Add `/usr/bin/local/touchlog` to `PATH`

Install source:

TODO

## Man Page

The man page for `touchlog` can be found here: [manpage.md](manpage.md). Please manually install touchlog with source to your system to use both `touchlog` and the man page.

To install the manpage:

TODO

## Changelog

To generate a changelog, use [`git-chglog`](https://github.com/git-chglog/git-chglog/). Follow this command:

```bash
git-chglog -o CHANGELOG.md
```

## License

See the [License](LICENSE).
