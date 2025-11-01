package panicanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var PanicFatalAnalyzer = &analysis.Analyzer{
	Name: "panicFatalAnalyzer",
	Doc:  "check for usage of panic, log.Fatal and os.Exit outside of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		var isMain bool
		ast.Inspect(f, func(n ast.Node) bool {
			switch expr := n.(type) {
			case *ast.FuncDecl:
				if expr.Name.Name == "main" && f.Name.Name == "main" {
					isMain = true
					return true
				}
				isMain = false
			case *ast.CallExpr:
				if id, ok := expr.Fun.(*ast.Ident); ok && id.Name == "panic" {
					pass.Reportf(n.Pos(), "panic call")
				}
			case *ast.SelectorExpr:
				if isMain {
					return true
				}
				pkg, ok := expr.X.(*ast.Ident)
				if !ok {
					return true
				}
				pkgName := pkg.Name
				funcName := expr.Sel.Name
				if pkgName == "os" && funcName == "Exit" {
					pass.Reportf(n.Pos(), "os.Exit call outside of main")
				} else if pkgName == "log" && funcName == "Fatal" {
					pass.Reportf(n.Pos(), "log.Fatal call outside of main")

				}
			}

			return true
		})
	}
	return nil, nil
}
