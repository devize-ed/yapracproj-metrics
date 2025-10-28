package analyze

import (
	"errors"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Errors for the analyzer
var (
	errFuncFatal = errors.New("log.Fatal detected outside of main function")
	errFuncExit  = errors.New("os.Exit detected outside of main function")
	errFuncPanic = errors.New("panic detected, usage of panic is forbidden")
)

// PanicExitAnalyzer is the analyzer for the panicexit package
var PanicExitAnalyzer = &analysis.Analyzer{
	Name: "panicexit",
	Doc:  "Reports any use of Go `panic`; Reports calls to `log.Fatal` and/or `os.Exit` outside of `main/main.go`",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Iterate over the files in the package
	for _, f := range pass.Files {
		// Check if the package is main
		isPackageMain := f.Name.Name == "main"

		// Track the name of the current function
		var functionName string

		ast.Inspect(f, func(n ast.Node) bool {
			// Check the type of the node
			switch x := n.(type) {
			case *ast.FuncDecl:
				// Update functionName when encounter a function declaration
				functionName = x.Name.Name

			case *ast.CallExpr:
				// Check if the function is panic
				if isFuncPanic(x) != nil {
					pass.Report(analysis.Diagnostic{Pos: x.Pos(), Message: errFuncPanic.Error()})
				}

				// Check if the function is log.Fatal or os.Exit
				if err := isFuncFatalOrExit(x); err != nil {
					if !isPackageMain {
						pass.Report(analysis.Diagnostic{Pos: x.Pos(), Message: err.Error() + " ; forbidden usage outside of main package"})
					} else {
						// Check if the function is main
						if functionName != "main" {
							// Determine the location of the function
							var locationName string
							// If the function name is empty, it means the function is at the package level
							if functionName == "" {
								locationName = "package level"
							} else {
								// If the function name is not empty, it means the function is a regular function
								locationName = "function " + functionName
							}
							pass.Report(analysis.Diagnostic{Pos: x.Pos(), Message: err.Error() + " ; forbidden usage in " + locationName + " outside of main function"})
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}

// isFuncPanic checks if the function is panic
func isFuncPanic(call *ast.CallExpr) error {
	if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
		return errFuncPanic
	}
	return nil
}

// isFuncFatalOrExit checks if the function is log.Fatal or os.Exit
func isFuncFatalOrExit(call *ast.CallExpr) error {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		// Check for log.Fatal
		if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "log" && sel.Sel.Name == "Fatal" {
			return errFuncFatal
		}
		// Check for os.Exit
		if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
			return errFuncExit
		}
	}
	return nil
}
