package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// noOsExitAnalyzer создает анализатор, который запрещает использование os.Exit
// в функции main пакета main.
func noOsExitAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "noOsExit",
		Doc:  "запрещает использование os.Exit в функции main пакета main",
		Run:  runNoOsExit,
	}
}

// runNoOsExit выполняет анализ: находит функцию main и проверяет, вызывается ли os.Exit.
func runNoOsExit(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() != "main" {
			continue
		}

		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != "main" {
				continue
			}

			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
						pass.Reportf(call.Pos(), "использование os.Exit запрещено в функции main")
					}
				}
				return true
			})
		}
	}
	return nil, nil
}
