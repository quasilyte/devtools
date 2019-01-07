# go-appcover

**go-appcover** simplifies the process of collecting main/application coverage report.

## Overview

This tool automates the process described in these articles:
* [Go coverage with external tests](https://blog.cloudflare.com/go-coverage-with-external-tests/)
* [Measure end-to-end tests coverage](https://dshipenok.github.io/external_coverage/)

It also goes a few extra miles in order to solve other problems, like `os.Exit` that your application might perform.

In short, it makes it easy to execute `main` package in a coverage-collecting mode.
You need 0 code to make it work, only a `main` package that you would like to instrument.

## Install

```bash
go get -u -v github.com/Quasilyte/devtools/cmd/go-appcover
```

## Quick start

Given a `cmd/foomain` main package that uses `github.com/foondation/foo` as a dependency,
you can collect the coverage profile of your application by doing:

```bash
# Enter a directory of the main package first.
cd $(go env GOPATH)/src/cmd/foomain

# A. Collect only cmd/foomain coverage info.
go-appcover run -coverpkg=cmd/foomain

# B. Include github.com/foondation/foo into the profile.
go-appcover run -coverpkg=cmd/foomain,github.com/foondation/foo

# C. The laziest way. Includes all dependencies into the result.
go-appcover run -coverpkg=all
```

Choose `A`, `B` or `C` option based on your goals.

If no errors occur, you will have `_appcover.out` file inside the current directory.
That file contains the coverage info of your application.

```bash
go tool cover -html=_appcover.out
```

If you have unit tests that you want to merge with that, use [gocovmerge](https://github.com/wadey/gocovmerge).

```bash
gocovmerge _appcover.out unittests.out > total.out
```

## Usage

```bash
$ go-appcover --help
Usage: go-appcover command [go test -c args...]
command is one of the below:
* build - create test binary in /tmp/_appcover, don't run it
* run - build and run test binary, merge partial coverage profiles
```

All arguments, except the first positional one (the command) are forwarded
to `go test -c` command to build the test binary.
You want to pass at least `-coverpkg` to that.
Another useful argument is `-covermode=count`.

### build command

Build command creates a special test file with `TestMain` that is
compiled to a test binary that is capable of testing your `main` function.

The binary is stored under your system temporary directory, name is always
`_appcover`. For most unix systems you get a `/tmp/_appcover` as a result.

The test binary will collect coverage info into `_appcover1.out` and `_appcover2.out`.
These files are a result of the exectution.
You can use either of them or, if one of them is empty, non-empty one.
Prefer the one that was most recently modified.
Or you can try using `gocovmerge` to join them (but don't do that with `-covermode=count`).

### run command

Does everything that `build` command does, plus runs the test binary.

After test binary exits, `_appcover1.out` and `_appcover2.out` are
merged into `_appcover.out`, so that you don't have to do it manually.

## Hints

### Testing interactive CLI apps that require stdin

Instead of using `run` command, use `build`.

Run `/tmp/_appcover` binary on your own and you'll have an opportunity
to use stdin to control the application while collecting the coverage info.

### Testing apps that can do os.Exit

`go-appcover` creates a test binary that writes partial results every few seconds.

If your application exits, you still get the results.
