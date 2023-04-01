// Package staticlint is the static analysis tool for shortener service.
//
// Usage: build package and run it with command: staticlint [package]
//
// Included analyzers:
// aasign - detects useless assignments;
// atomic - checks for common mistakes using the sync/atomic package;
// copylock - checks for locks erroneously passed by value;
// errorsas - checks that the second argument to errors.As is a pointer to a type implementing error;
// httpresponse - checks for mistakes using HTTP responses;
// loopclosure - checks for references to enclosing loop variables from within nested functions;
// lostcancel - checks for failure to call a context cancellation function;
// nilfunc - checks for useless comparisons against nil;
// printf - checks consistency of Printf format strings and arguments;
// shadow - checks for shadowed variables;
// sigchanyzer - detects misuse of unbuffered signal as argument to signal.Notify;
// sortslice - checks for calls to sort.Slice that do not use a slice type as first argument;
// structtag - checks struct field tags are well formed;
// timeformat - checks for the use of time.Format or time.Parse calls with a bad format;
// unmarshal - checks for passing non-pointer or non-interface types to unmarshal and decode functions;
// unreachable - checks for unreachable code;
// unusedresult - checks for unused results of calls to certain pure functions;
// unusedwrite - checks for unused writes to the elements of a struct or array object;
// usesgenerics - checks for usage of generic features added in Go 1.18;
// osExitAnalyser - checks for using os.Exit() in main();
// all staticcheck SA and ST analysers.
package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

const (
	staticcheckClass = "SA"
	stylecheckClass  = "ST"
)

func main() {
	analyzers := []*analysis.Analyzer{
		assign.Analyzer,
		atomic.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		structtag.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
		osExitAnalyser,
	}

	for _, a := range quickfix.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	for _, a := range simple.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	for _, a := range stylecheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	multichecker.Main(analyzers...)
}

var osExitAnalyser = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "checks for using os.Exit() in main()",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			x, ok := n.(*ast.FuncDecl)
			if !ok || x.Name.Name != "main" {
				return true
			}

			ast.Inspect(x, func(n ast.Node) bool {
				y, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				s, ok := y.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				i, ok := s.X.(*ast.Ident)
				if !ok {
					return true
				}

				if i.Name == "os" && s.Sel.Name == "Exit" {
					pass.Reportf(n.Pos(), "os.Exit using in main() is forbidden")
				}

				return true
			})

			return true
		})
	}

	return nil, nil
}
