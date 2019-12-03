package main

import (
	"errors"
	"io"
	"testing"
)

func TestPackage(t *testing.T) {
	err := errors.New("")
	if _, ok := err.(io.Reader); ok {
	}
}
