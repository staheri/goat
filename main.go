package main

import (
	"github.com/staheri/goatlib/hello"
 	_"github.com/staheri/goatlib/instrument"
	"flag"
	"fmt"
)



var (
	flagPath            string
	flagArgs            []string
	flagVerbose         bool
	validCategories    = []string{"CHNL", "GCMM", "GRTN", "MISC", "MUTX", "PROC", "SYSC", "WGCV", "SCHD", "BLCK"}
	validPrimeCmds     = []string{"word", "hac", "rr", "diff", "dineData", "cleanDB", "dev", "hb", "gtree", "cgraph", "resg","leakChecker"}
	validTestSchedCmds = []string{"test","execVis"}
	validSrc           = []string{"native", "x", "latest", "schedTest"}
)

func main(){
	fmt.Println("Initializing GOAT V.0.1 ...")
	parseFlags()

	test1(flagPath)

	hello.Hello("Saeed!")



	// SingleSource
	//      instrument
	//            concurrency usage
	//            tracing
	//            sched
	//            covearge guiding
	//      execute
	//            build
	//            run
	//            collect trace
	//            store
	//      measuring the coverage
	//      other reports
	//
	// Benchmark
	//   iterate over benchmark

}



func parseFlags() {
	//srcDescription := "native: execute the app and collect from scratch, latest: retrieve data from latest execution, x: retrieve data from specific execution (requires -x option)"
	// Parse flags
	flag.StringVar(&flagPath, "path", "", "Target application (*.go)")
	flag.BoolVar(&flagVerbose, "verb", false, "Print verbose info")

	flag.Parse()

	flagArgs = flag.Args()
}
