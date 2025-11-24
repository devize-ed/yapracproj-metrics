package main

import (
	"github.com/devize-ed/yapracproj-metrics.git/internal/analyze"

	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyze.PanicExitAnalyzer)
}
