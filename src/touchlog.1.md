---
title: ",TOUCHLOG"
section: 1
header: User Manual
footer: ",touchlog 1.0.0"
date: July 18, 2023
---

# NAME
,touchlog - a tool to create log files for a date

# SYNOPSIS
**,touchlog** [*-h|-v|-d mmddyyyy|-f [dir]*]

# DESCRIPTION
**,touchlog** is a tool to create simple log files for a date. It can be supplied a date in the format of *mmddyyyy* using the *-d* option or use the current date when no input is given. To write to a custom directory, ensure the directory first exists. Then, use the *-f [dir]* option.

# OPTIONS
**-h**
: display help message

**-v**
: display version message

**-d mmddyyyy**
: use a supplied date

**-f [dir]**
: write to existing inputted directory

# EXAMPLE
**,touchlog**
: a log file is create for today's date

**,touchlog -v**
: display version message

**,touchlog -h**
: display help message

**,touchlog -d 04301998**
: a log file is create for date April 30, 1998

**,touchlog -f logs**
: a log file is created for today's date in the "logs" folder

**,touchlog -d 04301998 -f logs**
: a log file is create for date April 30, 1998 in the "logs" folder

# AUTHORS
Written by Sasank 'squatch$' Vishnubhatla

# BUGS
Submit bug reports to Sasank via email: [sasank@vishnubhatlas.net](mailto:sasank@vishnubhatlas.net)
