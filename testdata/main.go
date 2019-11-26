package main

import (
	"errors"
	"io"
)

func main() {
	fn()
}

func fn() {
	err := errors.New("")
	if err == io.EOF {
	}
}
