package main

import (
	"go-variable-cleaner"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(go-variable-cleaner.Analyzer) }

