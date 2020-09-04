package varcleaner

import (
	"fmt"
	"go/ast"
	"go/token"
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
			fmt.Println(unnecessaryVars, replaceConsts)

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

func walk(fd *ast.FuncDecl) ([]string, []string) {
	vars := map[string]int{}   // count number of each variable appearance
	consts := map[string]int{} // count number of unique consts

	// Walk through AST to obtain variable / consts and its number of appearance
	walkDecl(fd, vars, consts)
	fmt.Println(vars, consts)

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

func walkDecl(n ast.Node, vars map[string]int, consts map[string]int) {
	switch n := n.(type) {
	case *ast.GenDecl:
		walkToken(n.Tok, vars, consts)
		for _, s := range n.Specs {
			walkSpec(s, vars, consts)
		}
	case *ast.FuncDecl:
		walkStmt(n.Body, vars, consts)
	}
}

func walkStmt(n ast.Node, vars map[string]int, consts map[string]int) {
	switch n := n.(type) {
	case *ast.DeclStmt:
		walkDecl(n.Decl, vars, consts)
	case *ast.ExprStmt:
		walkExpr(n.X, vars, consts)
	case *ast.SendStmt:
		walkExpr(n.Chan, vars, consts)
		walkExpr(n.Value, vars, consts)
	case *ast.IncDecStmt:
		walkExpr(n.X, vars, consts)
	case *ast.AssignStmt:
		for _, exp := range n.Lhs {
			walkExpr(exp, vars, consts)
		}
		for _, exp := range n.Rhs {
			walkExpr(exp, vars, consts)
		}
	case *ast.GoStmt:
		walkExpr(n.Call, vars, consts)
	case *ast.DeferStmt:
		walkExpr(n.Call, vars, consts)
	case *ast.ReturnStmt:
		for _, e := range n.Results {
			walkExpr(e, vars, consts)
		}
	case *ast.BranchStmt:
		if n.Label != nil {
			walkExpr(n.Label, vars, consts)
		}
	case *ast.BlockStmt:
		for _, s := range n.List {
			walkStmt(s, vars, consts)
		}
	case *ast.IfStmt:
		if n.Init != nil {
			walkStmt(n.Init, vars, consts)
		}
		walkExpr(n.Cond, vars, consts)
		walkStmt(n.Body, vars, consts)
		if n.Else != nil {
			walkStmt(n.Else, vars, consts)
		}
	case *ast.SwitchStmt:
		if n.Init != nil {
			walkStmt(n.Init, vars, consts)
		}
		if n.Tag != nil {
			walkExpr(n.Tag, vars, consts)
		}
		walkStmt(n.Body, vars, consts)
	case *ast.SelectStmt:
		walkStmt(n.Body, vars, consts)
	case *ast.ForStmt:
		if n.Init != nil {
			walkStmt(n.Init, vars, consts)
		}
		if n.Cond != nil {
			walkExpr(n.Cond, vars, consts)
		}
		if n.Post != nil {
			walkStmt(n.Post, vars, consts)
		}
		walkStmt(n.Body, vars, consts)
	case *ast.RangeStmt:
		if n.Key != nil {
			walkExpr(n.Key, vars, consts)
		}
		if n.Value != nil {
			walkExpr(n.Value, vars, consts)
		}
		walkExpr(n.X, vars, consts)
		walkStmt(n.Body, vars, consts)
	case *ast.CaseClause:
		if n.List != nil {
			for _, c := range n.List {
				walkExpr(c, vars, consts)
			}
		}
		if n.Body != nil {
			for _, b := range n.Body {
				walkStmt(b, vars, consts)
			}
		}
	}
}

func walkSpec(spec ast.Spec, vars map[string]int, consts map[string]int) {
	switch spec := spec.(type) {
	case *ast.ValueSpec:
		for _, n := range spec.Names {
			walkExpr(n, vars, consts)
			if spec.Type != nil {
				walkExpr(spec.Type, vars, consts)
			}
			if spec.Values != nil {
				for _, v := range spec.Values {
					walkExpr(v, vars, consts)
				}
			}
		}
	}
}
func walkExpr(exp ast.Expr, vars map[string]int, consts map[string]int) {
	switch exp := exp.(type) {
	case *ast.ParenExpr:
		walkExpr(exp.X, vars, consts)
	case *ast.SelectorExpr:
		walkExpr(exp.X, vars, consts)
		walkExpr(exp.Sel, vars, consts)
	case *ast.IndexExpr:
		walkExpr(exp.X, vars, consts)
		walkExpr(exp.Index, vars, consts)
	case *ast.SliceExpr:
		walkExpr(exp.X, vars, consts)
		if exp.Low != nil {
			walkExpr(exp.Low, vars, consts)
		}
		if exp.High != nil {
			walkExpr(exp.High, vars, consts)
		}
		if exp.Max != nil {
			walkExpr(exp.Max, vars, consts)
		}
	case *ast.TypeAssertExpr:
		walkExpr(exp.X, vars, consts)
		if exp.Type != nil {
			walkExpr(exp.Type, vars, consts)
		}
	case *ast.CallExpr:
		walkExpr(exp.Fun, vars, consts)
		for _, a := range exp.Args {
			walkExpr(a, vars, consts)
		}
	case *ast.StarExpr:
		walkExpr(exp.X, vars, consts)
	case *ast.UnaryExpr:
		walkToken(exp.Op, vars, consts)
		walkExpr(exp.X, vars, consts)
	case *ast.BinaryExpr:
		walkExpr(exp.X, vars, consts)
		walkExpr(exp.Y, vars, consts)
	case *ast.KeyValueExpr:
		walkExpr(exp.Key, vars, consts)
		walkExpr(exp.Value, vars, consts)
	case *ast.BasicLit:
		consts[exp.Value]++
	case *ast.FuncLit:
		walkExpr(exp.Type, vars, consts)
	case *ast.CompositeLit:
		if exp.Type != nil {
			walkExpr(exp.Type, vars, consts)
		}
		for _, e := range exp.Elts {
			walkExpr(e, vars, consts)
		}
	case *ast.Ident:
		if exp.Obj != nil {
			vars[exp.Name]++
		}
	case *ast.Ellipsis:
		if exp.Elt != nil {
			walkExpr(exp.Elt, vars, consts)
		}
	case *ast.FuncType:
		if exp.Params.List != nil {
			for _, f := range exp.Params.List {
				walkExpr(f.Type, vars, consts)
			}
		}
	case *ast.ChanType:
		walkExpr(exp.Value, vars, consts)
	}
}

func walkToken(t token.Token, vars map[string]int, consts map[string]int) {
	switch t {
	case token.IDENT:
		vars[t.String()]++
	case token.INT, token.FLOAT, token.IMAG, token.CHAR, token.STRING:
		consts[t.String()]++
	}
}
