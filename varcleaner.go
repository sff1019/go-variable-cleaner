package varcleaner

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "varcleaner is a tool to make codes more readable and scalable"

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "varcleaner",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.FuncDecl:
			unnecessaryVars, replaceConsts := walk(n)

			if len(unnecessaryVars) > 0 && len(replaceConsts) > 0 {
				unnecessaryVarsStr := strings.Join(unnecessaryVars, ", ")
				replaceConstsStr := strings.Join(replaceConsts, ", ")
				pass.Reportf(n.Pos(), "No need to define these variables: %v\nUsed same consts multiple times, replace with variable: %v", unnecessaryVarsStr, replaceConstsStr)
			} else if len(unnecessaryVars) > 0 {
				unnecessaryVarsStr := strings.Join(unnecessaryVars, ", ")
				pass.Reportf(n.Pos(), "No need to define these variables: %v", unnecessaryVarsStr)
			} else if len(replaceConsts) > 0 {
				replaceConstsStr := strings.Join(replaceConsts, ", ")
				pass.Reportf(n.Pos(), "Used same consts multiple times, replace with variable: %v", replaceConstsStr)
			}
		}
	})

	return nil, nil
}

type branchVisitor func(n ast.Node) (w ast.Visitor)

// Visit is ...
func (v branchVisitor) Visit(n ast.Node) (w ast.Visitor) {
	return v(n)
}

func walk(fd *ast.FuncDecl) ([]string, []string) {
	vars := map[string]int{}   // count number of each variable appearance
	consts := map[string]int{} // count number of unique consts

	// // Walk through AST to obtain variable / consts and its number of appearance
	var v ast.Visitor
	v = branchVisitor(func(n ast.Node) (w ast.Visitor) {
		switch n := n.(type) {
		case *ast.BasicLit:
			consts[n.Value]++
		case *ast.Ident:
			if n.Obj != nil {
				vars[n.Name]++
			}
		}
		return v
	})
	ast.Walk(v, fd)

	// Check whether it is unnecessary to define the variables
	unnecessaryVars := checkVarNecessity(vars)
	// Check whether consts can be replaced by a variable
	replaceConsts := checkConstsUnnecessity(consts)

	return unnecessaryVars, replaceConsts
}

func checkVarNecessity(vars map[string]int) []string {
	var unnecessary []string
	for k, v := range vars {
		if v == 2 {
			unnecessary = append(unnecessary, k)
		}
	}

	return unnecessary
}

func checkConstsUnnecessity(consts map[string]int) []string {
	var replace []string
	for k, v := range consts {
		if v >= 2 {
			replace = append(replace, k)
		}
	}

	return replace
}
