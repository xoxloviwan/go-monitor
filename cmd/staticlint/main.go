package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"honnef.co/go/tools/quickfix/qf1011"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck/st1013"
)

func main() {
	mychecks := []*analysis.Analyzer{defers.Analyzer, loopclosure.Analyzer}
	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}
	mychecks = append(mychecks, qf1011.Analyzer)
	mychecks = append(mychecks, st1013.Analyzer)
	mychecks = append(mychecks, NoMainExitCheck)

	multichecker.Main(mychecks...)
}

var NoMainExitCheck = &analysis.Analyzer{
	Name: "nomainexit",
	Doc:  "direct call os.Exit() from main() function of main package not allowed",
	Run:  run,
}

func isPkgDot(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	return ok && isIdent(sel.X, pkg) && isIdent(sel.Sel, name)
}

func isIdent(expr ast.Expr, ident string) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == ident
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() != "main" {
			continue
		}

		// ignore testing packages
		next := false
		imports := pass.Pkg.Imports()
		for _, v := range imports {
			if v.Name() == "testing" {
				next = true
				break
			}
		}
		if next {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {

			if fn, ok := node.(*ast.FuncDecl); ok {
				if fn.Name.String() == "main" {
					ast.Inspect(fn.Body, func(n ast.Node) bool {
						if c, ok := n.(*ast.CallExpr); ok {
							if isPkgDot(c.Fun, "os", "Exit") {
								pass.Reportf(c.Pos(), "os.Exit() from main() function of main package not allowed")
							}
						}
						return true
					})
				}
			}
			return true
		})
	}
	return nil, nil
}
