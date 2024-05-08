---
title: "TOUCHLOG"
section: 1
header: User Manual
footer: "touchlog 1.2.2"
date: May 1, 2024
---

<!-- markdownlint-disable-file MD025 -->

# NAME

touchlog - a tool to create log files for a date

# SYNOPSIS

**touchlog** [*-version|-verbose|-outdir [dir]|-date [mmddyyyy]|-help*]

# DESCRIPTION

**touchlog** is a tool to create simple log files for a date. It can be supplied a date in the format of *mmddyyyy* using the *-d* option or use the current date when no input is given. To write to a custom directory, ensure the directory first exists. Then, use the *-f [dir]* option.

# OPTIONS

**-help**
: display help message

**-version**
: display version message

**-date [mmddyyyy]**
: use a supplied date

**-outdir [dir]**
: write to existing inputted directory

**-verbose**
: enable verbose mode

# EXAMPLE

**touchlog**
: a log file is create for today's date

**touchlog -version**
: display version message

**touchlog -help**
: display help message

**touchlog -date 04301998**
: a log file is create for date April 30, 1998

**touchlog -outdir logs**
: a log file is created for today's date in the "logs" folder

**touchlog -date 04301998 -outdir logs**
: a log file is create for date April 30, 1998 in the "logs" folder

# AUTHORS

Written by Sasank 'squatch$' Vishnubhatla

# BUGS

Submit bug reports to Sasank via email: [sasank@vishnubhatlas.net](mailto:sasank@vishnubhatlas.net)
