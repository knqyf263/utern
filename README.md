# utern


[![GitHub release](https://img.shields.io/github/release/knqyf263/utern.svg)](https://github.com/knqyf263/utern/releases/latest)
[![Build Status](https://travis-ci.org/knqyf263/utern.svg?branch=master)](https://travis-ci.org/knqyf263/utern)
[![Go Report Card](https://goreportcard.com/badge/github.com/knqyf263/utern)](https://goreportcard.com/report/github.com/knqyf263/utern)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](https://github.com/knqyf263/utern/blob/master/LICENSE)

<img src="img/logo.png" width="300">

The “best of best” way to tail AWS CloudWatch Logs from your terminal.

<img src="img/usage.gif" width="700">


# Abstract
`Utern` allows you to tail log events from multiple log groups and log streams on AWS CloudWatch Logs. Each result is color coded for quicker debugging. Inspired by [stern](https://github.com/wercker/stern).

The query is a regular expression so the log group name and stream name can easily be filtered and you don't need to specify the exact name. If a stream is deleted it gets removed from tail and if a new stream is added it automatically gets tailed.

When a log group contains multiple log streams, `Utern` can tail all of them too without having to do this manually for each one. Simply specify the filter to limit what log events to show.

```
$ utern [options] log-group-query
```

So Simple!!

The log-group-query is a regular expression so you could provide "web-\w" to tail web-backend and web-frontend log groups but not web-123.



# Features
- **Multi log groups tailing in parallel**
  - Regular expression
- **Multi log streams tailing in parallel**
  - Regular expression
- **Colorful**
  - Quicker debugging
- Flexible date and time parser
  - Human friendly formats, i.e. 1h20m to indicate 1 hour and 20 minutes ago
  - A full timestamp 2019-01-02T03:04:05Z (RFC3339)
- Powerful built-in filter
  -  https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html
- Fast
  - Written in golang
- Easy installation
  - Pre-built binaries

# Installation
## From source

```console
$ go get -u github.com/knqyf263/utern
```

## Binary (Including Windows)
Go to [the releases page](https://github.com/knqyf263/utern/releases), find the version you want, and download the zip file. Unpack the zip file, and put the binary to somewhere you want (on UNIX-y systems, /usr/local/bin or the like). Make sure it has execution bits turned on.

## Mac OS X / Homebrew
You can use homebrew on OS X.
```
$ brew tap knqyf263/utern
$ brew install knqyf263/utern/utern
```

If you receive an error (`Error: knqyf263/utern/utern 64 already installed`) during `brew upgrade`, try the following command

```
$ brew unlink utern && brew uninstall utern
($ rm -rf /usr/local/Cellar/utern/64)
$ brew install knqyf263/utern/utern
```

## RedHat, CentOS
Download rpm package from [the releases page](https://github.com/knqyf263/utern/releases)
```
$ sudo rpm -ivh https://github.com/knqyf263/utern/releases/download/v0.0.1/utern_0.0.1_Tux_64-bit.rpm
```

## Debian, Ubuntu
Download deb package from [the releases page](https://github.com/knqyf263/utern/releases)
```
$ wget https://github.com/knqyf263/utern/releases/download/v0.0.1/utern_0.0.1_Tux_64-bit.deb
$ sudo dpkg -i utern_0.0.1_Tux_64-bit.deb
```

# Examples
Some examples are shown below.

### List all log groups

```console
$ aws logs describe-log-groups --query "logGroups[].[logGroupName]" --output text
```

### List all log streams

```console
$ aws logs describe-log-streams --log-group-name log-group-name --query "logStreams[].[logStreamName]" --output text

```

### All log streams

```console
$ utern log-group-query
```

### Filter log groups with regular expressions

```console
$ utern "web-\w"
```

### Filter log streams with regular expressions (--stream, -n)

```console
$ utern --stream log-stream-query log-group-query
```

### Filter log streams with a prefix of log stream name (--stream-prefix, -p)
If the log group has many log streams, `--stream-prefix` will be faster than `--stream`.

```console
$ utern --stream-prefix log-stream-prefix log-group-query
```

### Filter log streams with a prefix and regular expressions

```console
$ utern -p log-stream-prefix -n log-stream-query log-group-query
```

### Logs after 1 hour ago (--since, -s)

```console
$ utern --since 1h log-group-query
```

### Logs after 2019-01-02 03:04:05 UTC
RFC3339

```console
$ utern --since 2019-01-02T03:04:05Z log-group-query
```

### Logs from 10 minutes ago to 5 minutes ago

```console
$ utern --since 10m --end 5m log-group-query
```

# Usage

```console
NAME:
   utern - Multi group and stream log tailing for AWS CloudWatch Logs

USAGE:
   utern [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --stream value, -n value         Log stream name (regular expression). Displays all if omitted. If the option "since" is set to recent time, this option usually makes it faster than the option "stream-prefix"
   --stream-prefix value, -p value  Log stream name prefix. If a log group contains many log streams, this option makes it faster.
   --since value, -s value          Return logs newer than a relative duration like 52, 2m, or 3h. (default: "5m")
   --end value, -e value            Return logs older than a relative duration like 0, 2m, or 3h.
   --profile value                  Specify an AWS profile.
   --code value                     Specify MFA token code directly (if applicable), instead of using stdin.
   --region value, -r value         Specify an AWS region.
   --filter value                   The filter pattern to use. For more information, see https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html.
   --timestamps                     Print timestamps
   --event-id                       Print event ID
   --no-log-group                   Suppress display of log group name
   --no-log-stream                  Suppress display of log stream name
   --max-length value               Maximum log message length (default: 0)
   --color                          Force color output even if not a tty
   --help, -h                       show help
   --version, -v                    print the version
```

# Contribute

1. fork a repository: github.com/knqyf263/utern to github.com/you/repo
2. get original code: `go get github.com/knqyf263/utern`
3. work on original code
4. add remote to your repo: git remote add myfork https://github.com/you/repo.git
5. push your changes: git push myfork
6. create a new Pull Request

- see [GitHub and Go: forking, pull requests, and go-getting](http://blog.campoy.cat/2014/03/github-and-go-forking-pull-requests-and.html)

----

# License
MIT

# Author
Teppei Fukuda
