package main

import (
	"varcleaner"

	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(varcleaner.Analyzer) }
