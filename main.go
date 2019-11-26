package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/romanyx/erris/internal/checker"
	"github.com/romanyx/erris/internal/visitor"
)

func main() {
	var (
		withoutTests = flag.Bool("ignoretests", false, "if true, checking of _test.go files is disabled")
	)

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	checker := checker.New(*withoutTests)
	err := checker.CheckPackages(args...)
	if err != nil {
		var issues visitor.Issues
		if !errors.As(err, &issues) {
			log.Println(err)
			os.Exit(1)
		}

		for _, r := range issues {
			fmt.Printf("%s:\t%s\n", r.Pos, r.Text)
		}
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: erris ./...")
}
