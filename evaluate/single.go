package evaluate

import(
  "fmt"
  "os"
  "path/filepath"
  "encoding/json"
  "github.com/staheri/goatlib/trace"
  "github.com/staheri/goatlib/traceops"
  "strings"
  "strconv"
  "github.com/jedib0t/go-pretty/table"
)

func EvaluateSingle(path string, freq, d int, isRace bool){
  colorReset := "\033[0m"
  colorRed := "\033[31m"
  colorGreen := "\033[32m"

  exp := ""
  if isRace{
    exp = "goat_race_d"+strconv.Itoa(d)
  } else{
    exp = "goat_d"+strconv.Itoa(d)
  }

  // obtain MAXPROCS
  MAXPROCS := os.Getenv("GOATMAXPROCS")
  if MAXPROCS == "" {
    panic("GOATMAXPROCS is not set!")
  } else{
    mp,err := strconv.Atoi(MAXPROCS)
    check(err)
    if mp < 1 || mp > 64 {
      panic("GOATMAXPROCS out of range!")
    }
  }

  // init table
  tall := table.NewWriter()
  tall.SetOutputMirror(os.Stdout)
  tall.AppendHeader(table.Row{"Bug","Delay Bound","1st Fail","Cov1 Leaps", "Cov2 Leaps","Errors","Cov1 > Cov2","Final Coverage"})

  fmt.Println("p: ",path)
  bugName := filepath.Base(path)
  //bugType := bugInfo[0]
  //bugName := bugInfo[1] + "_" + bugInfo[2]
  //bugFullName := bugInfo[0] + "_" + bugInfo[1] + "_" + bugInfo[2]

  if bugName == ""{
    panic("wrong bugName")
  }
  mainExp := &RootExperiment{}
  // create bug now
  fmt.Println("BugName:",bugName,", path: ",path)

  target := &Bug{bugName,path,"single","x","y"}
  mainExp.Bug = target
  mainExp.Exps = make(map[string]Ex)

  // figure exp report file
  ws := os.Getenv("GOATWS")
  if ws == "" {
    panic("GOATWS is not set!")
  }
  // we have to re-discover gex prefixdir to see if we have ready results
  expReport := ws + "/"
  expReport = expReport + "p"+MAXPROCS+"/"
  expReport = expReport + "single_"+ bugName + "/"

  if strings.HasPrefix(exp,"goat_d"){
    d,err := strconv.Atoi(strings.Split(exp,"_d")[1])
    check(err)
    if d < 1 {
      expReport = expReport + "goat_trace/"
    } else{
      expReport = expReport + "goat_delay/"
    }
    //ex = &GoatExperiment{Experiment: Experiment{Target:target},Bound:d}
  } else if strings.HasPrefix(exp,"goat_race"){
    d,err := strconv.Atoi(strings.Split(exp,"_d")[1])
    check(err)
    if d < 1 {
      expReport = expReport + "goat_race_trace/"
    } else{
      expReport = expReport + "goat_race_delay/"
    }
    //ex = &GoatExperiment{Experiment: Experiment{Target:target},Bound:d}
  } else{
    panic("wrong exp")
  }
  expReport = expReport + "results/"
  expReport = expReport + "p"+MAXPROCS+"_"+bugName+"_"+exp+"_T"+strconv.Itoa(freq)+"_"+TERMINATION+".json"

  gex := &GoatExperiment{}
  if checkFile(expReport){ // if report exist
    gex = ReadExperimentResults_goat(expReport)
    gex.Target = target
    // gex lacks coverage data (lstack, gstack, concurrency usage,...)

    // missing init stuff
    // setup global stack
    fmap := make(map[int]*trace.Frame)
    fsmap := make(map[string]int)
    gstack := &GlobalStack{fmap,fsmap}
    gex.GStack = gstack

    // missing instrument stuff
    concUsage := ReadConcUsage(gex.PrefixDir+"/concUsage.json")
    if concUsage != nil{
      gex.ConcUsage = &ConcUsageStruct{ConcUsage:concUsage}
      gex.InitConcMap()
    }

    // missing build stuff
    dest := filepath.Join(gex.PrefixDir,"bin")
    files, err := filepath.Glob(dest+"/*"+gex.GetMode())
    check(err)
    if len(files) != 0{ // check if binary exist
      gex.BinaryName = filepath.Base(files[0]) // assign the first found binary to current gex binaryPath
    }

    // missing execute stuff (from results)
    newResults := []*Result{}
    // iterate over results
    for _,res := range(gex.Results){
      result := res
      // get the local stack
      if res.Desc == "CRASH"|| res.Desc == "NONE" || strings.HasPrefix(res.Desc,"PANIC"){
        // it does not have any trace
        newResults = append(newResults,result)
        continue
      }
      fmt.Println("Reading trace ",result.TracePath)
      fmt.Println("\tRES: ",res.Desc)
      parseRes := traceops.ReadParseTrace(result.TracePath, filepath.Join(gex.PrefixDir,"bin",gex.BinaryName))

      // print events
      // for i,e := range(parseRes.Events){
      //   fmt.Printf("****\nG%v (idx:%v)\n%v\n",e.G,i,e.String())
      // }

      result.LStack = gex.UpdateGStack(parseRes.Stacks)
      gex.UpdateConcUsage(parseRes.Stacks,result.LStack)
      gex.UpdateGGTree(parseRes,result.LStack)
      gex.UpdateCoverageGGTree(parseRes,result.LStack)
      gex.UpdateCoverageReport()
      result.Coverage1 = gex.PrintCoverageReport(true)
      result.Coverage2 = gex.PrintCoverageReport(false)
      newResults = append(newResults,result)
    }
    gex.Results = newResults
  } else{
    gex = &GoatExperiment{Experiment: Experiment{Target:target},Bound: d}
    gex.Init(isRace)
    gex.Instrument()
    gex.Build(isRace)

iteration:for i:=0 ; i < freq ; i++{
      fmt.Printf("Test %v on %v (%d/%d)\n",gex.Target.BugName,gex.ID,i+1,freq)
      res := gex.Execute(i,isRace)
      if res.Detected{
        fmt.Println(string(colorRed),res.Desc,string(colorReset))
        gex.Results = append(gex.Results,res)
        switch TERMINATION {
        case "hitBug":

          break iteration
        case "ignoreGDL":
          if res.Desc == "GDL"{
            break iteration
          }
        default:
        }
      }else{
        if res.Desc == "CRPT_TRACE"{
          // we want to re-run
          i = i - 1
          continue
        }
        fmt.Println(string(colorGreen),"PASS",string(colorReset))
        gex.Results = append(gex.Results,res)
      }
    }

    // store gex into json
    // check if expReport path matches gex.PrefixDir
    if filepath.Dir(expReport) != gex.PrefixDir+"/results" {
      panic(fmt.Sprintf("mismatch expReport dir\n%v\n%v\n)",filepath.Dir(expReport),gex.PrefixDir+"/results"))
    }
    // write to json file
    rep,err := os.Create(expReport)
    check(err)
    newdat ,err := json.MarshalIndent(gex,"","    ")
    check(err)
    _,err = rep.WriteString(string(newdat))
    check(err)
    rep.Close()
  }

  // generate execViz and gtree
  outp := ""
  if gex.LastFailedTrace != ""{
    outp = fmt.Sprintf("%v/FAIL_%v",gex.PrefixDir+"/visual",strings.Split(filepath.Base(gex.LastFailedTrace),".trace")[0])
    traceops.ExecVis(gex.LastFailedTrace,filepath.Join(gex.PrefixDir,"bin",gex.BinaryName),outp,false)
    traceops.ExecVis(gex.LastFailedTrace,filepath.Join(gex.PrefixDir,"bin",gex.BinaryName),outp,true)
    traceops.ExecGtree(gex.LastFailedTrace,filepath.Join(gex.PrefixDir,"bin",gex.BinaryName),outp)
  }
  if gex.LastSuccessTrace != "" {
    outp = fmt.Sprintf("%v/SUCC_%v",gex.PrefixDir+"/visual",strings.Split(filepath.Base(gex.LastSuccessTrace),".trace")[0])
    traceops.ExecVis(gex.LastSuccessTrace,filepath.Join(gex.PrefixDir,"bin",gex.BinaryName),outp,false)
    traceops.ExecVis(gex.LastSuccessTrace,filepath.Join(gex.PrefixDir,"bin",gex.BinaryName),outp,true)
    traceops.ExecGtree(gex.LastSuccessTrace,filepath.Join(gex.PrefixDir,"bin",gex.BinaryName),outp)
  }

  
  mainExp.Exps[gex.ID] = gex
  tall.AppendRow(CoverageSummary(gex))
  tall.AppendSeparator()
  tall.RenderCSV()
  tall.Render()
}
