package main

import (
	"github.com/romanyx/erris"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(erris.Analyzer) }
