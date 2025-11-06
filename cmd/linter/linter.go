package main

import (
	"github.com/dmitastr/yp_observability_service/cmd/linter/panicanalyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(panicanalyzer.PanicFatalAnalyzer)
}
