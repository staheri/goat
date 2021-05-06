package main

import (
  _"github.com/staheri/goat/evaluate"
	"flag"
	"fmt"
	"log"
	"os"
	//"path/filepath"
	_"bufio"
)



var (
	flagPath            string
  flagTool            string
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
	//paths := []string{flagPath}
	//test1(flagPath)
	//paths, err := filepath.Glob("/Users/saeed/gobench/gobench/goker/blocking/kubernetes/*")
	//check(err)
	//paths2,err := filepath.Glob("/Users/saeed/gobench/gobench/goker/blocking/moby/*")
	//check(err)
	//paths3,err := filepath.Glob("/Users/saeed/gobench/gobench/goker/blocking/serving/*")
	//check(err)
	//paths4,err := filepath.Glob("/Users/saeed/gobench/gobench/goker/blocking/syncthing/*")
	//check(err)
	//paths = append(paths,paths2...)
	//paths = append(paths,paths3...)
	//paths = append(paths,paths4...)


	//f,err := os.Create("output-block-kub-mob-serv-sync.csv")

	//f,err := os.Create("out.csv")
  //check(err)
  //w := bufio.NewWriter(f)

	/*for i,p := range(paths){
		fmt.Println(i,p)
		experiment(p,40,w)
		w.Flush()
	}*/

  //evaluate.TAB_counts()
  //evaluate.EvaluateBlocking(flagPath,100)
  //evaluate.EvaluateNonBlocking(flagPath,500)

  //evaluate.EvaluateOverhead(flagPath,100,[]int{1,2,4,16,64,256,512,1024,2048})
  //

  //checkVis()
  //checkChecker()
  //checkJson()

  customVis(flagPath,flagTool,true)



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
  flag.StringVar(&flagTool, "tool", "goat_m", "tool id")
	flag.BoolVar(&flagVerbose, "verb", false, "Print verbose info")

	flag.Parse()

	flagArgs = flag.Args()
}
