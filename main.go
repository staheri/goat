package main

import (
  "github.com/staheri/goat/evaluate"
	"flag"
	"fmt"
	"log"
	"os"
	//"path/filepath"
	_"bufio"
)


var (
	flagPath            string
  flagConf            string
  flagFreq            int
  flagD               int
	flagCoverage        bool
  flagRace            bool
  flagJsonTrace       bool
  flagArgs            []string
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

  if flagPath != "" {
    evaluate.EvaluateSingle(flagPath,flagFreq,flagD,flagRace, flagJsonTrace)
  } else{
    if flagConf != ""{
      evaluate.EvaluateComparison(flagConf,flagFreq, flagJsonTrace)
      if flagCoverage {
        evaluate.EvaluateCoverage(flagConf,flagFreq,flagRace, flagJsonTrace)
      }
    }else{
      panic("GoAT: wrong args")
    }
  }
  //evaluate.TAB_counts()
  //evaluate.EvaluateBlocking(flagPath,100)
  evaluate.EvaluateNonBlocking(flagPath,500, flagJsonTrace) // path to config , frequence,

  //evaluate.EvaluateOverhead(flagPath,100,[]int{1,2,4,16,64,256,512,1024,2048})
  //

  //checkVis()
  //checkChecker()
  //checkJson()

  //customVis(flagPath,flagTool,true)
  //evaluate.EvaluateCoverage(flagPath,1000,false)
  evaluate.EvaluateComparison(flagPath,1, flagJsonTrace) // race = false
  //evaluate.EvaluateComparison(flagPath,2,true) // race = true
  //evaluate.TraceSnippet(flagPath)

}



func parseFlags() {
	//srcDescription := "native: execute the app and collect from scratch, latest: retrieve data from latest execution, x: retrieve data from specific execution (requires -x option)"
	// Parse flags
	flag.StringVar(&flagPath, "path", "", "target folder")
  flag.StringVar(&flagConf, "eval_conf", "", "config file with benchmark paths in it")
  flag.IntVar(&flagD, "d", 0, "number of delays")
  flag.IntVar(&flagFreq, "freq", 1, "frequency of executions")
  flag.BoolVar(&flagCoverage,"cov",false,"include coverage report in evaluation")
  flag.BoolVar(&flagRace,"race",false,"enable race detection")
  flag.BoolVar(&flagJsonTrace,"json_trace",false,"enable collection of execution traces in readable JSON format.")

	flag.Parse()

	flagArgs = flag.Args()
}
