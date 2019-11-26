# erris

erris is a program for checking that errors are compared or type asserted using go1.13 `errors.Is` and `errors.As` functions.

[![Build Status](https://travis-ci.org/romanyx/erris.png?branch=master)](https://travis-ci.org/romanyx/erris)
[![Report](https://goreportcard.com/badge/github.com/romanyx/erris)](https://goreportcard.com/report/github.com/romanyx/erris)

## Install

```sh
go get -u github.com/romanyx/erris
```

## Use

For basic usage, just give the package path of interest as the first argument:

```sh
erris github.com/romanyx/erris/testdata
```

Outputs:

```sh
github.com/romanyx/erris/testdata/main.go:14:5:	use errors.Is to compare an error
github.com/romanyx/erris/testdata/main_test.go:11:14:	use errors.As to assert an error
```

To check all packages beneath the current directory:

```
erris ./...
```
