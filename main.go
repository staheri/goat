package main

import (
 	"github.com/staheri/goatlib/db"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"bufio"
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

	// set log
	file, err := os.OpenFile("GOAT_log.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	check(err)
  log.SetOutput(file)

	parseFlags()
	//test1(flagPath)
	paths, err := filepath.Glob("/Users/saeed/gobench/gobench/goker/*/*/*")
	check(err)
	db.Clean()

	f,err := os.Create("output.csv")
  check(err)
  w := bufio.NewWriter(f)

	for i,p := range(paths){
		fmt.Println(i,p)
		experiment(p,1,w)
		w.Flush()
	}



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
