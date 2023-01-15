// Main file for evaluation
package evaluate

import(
  "fmt"
  "os"
  "path/filepath"
  "encoding/json"
  "github.com/staheri/goatlib/instrument"
  "github.com/staheri/goatlib/trace"
  "github.com/staheri/goatlib/traceops"
  "strings"
  "strconv"
  "github.com/jedib0t/go-pretty/table"
)


type RootExperiment struct{
  Bug         *Bug
  Exps        map[string]Ex      `json:"exps"`
  ReportJSON  string             `json:"reportJSON"`
}

var(
  coverage_d1 = []string{"goat_d0"}
  coverage_d2 = []string{"goat_d0","goat_d1"}
  coverage_d3 = []string{"goat_d0","goat_d1","goat_d2"}
  coverage_d5 = []string{"goat_d0","goat_d1","goat_d2","goat_d3","goat_d4"}

  coverage_race_d1 = []string{"goat_race_d0"}
  coverage_race_d2 = []string{"goat_race_d0","goat_race_d1"}
  coverage_race_d3 = []string{"goat_race_d0","goat_race_d1","goat_race_d2"}
  coverage_race_d5 = []string{"goat_race_d0","goat_race_d1","goat_race_d2","goat_race_d3","goat_race_d4"}

  coverage_d10 = []string{"goat_d0","goat_d1","goat_d2","goat_d3","goat_d4","goat_d5","goat_d6","goat_d7","goat_d8","goat_d9"}

  comparison = []string{"goat_d0","goat_d1","goat_d2","goat_d3","goat_d4","builtinDL","goleak","lockDL"}
)

func EvaluateCoverage(configFile string, thresh int, isRace bool, json_trace bool) {
  colorReset := "\033[0m"
  colorRed := "\033[31m"
  colorGreen := "\033[32m"

  dx := []string{}
  if isRace{
    dx = coverage_race_d1
  } else{
    dx = coverage_d5
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


  // a map to hold each RootExperiment
  allBugs := make(map[string]*RootExperiment)

  // read causes
  causes := make(map[string]map[string][]string)
  causes["blocking"] = ReadGoKerConfig("blocking")
  causes["nonblocking"] = ReadGoKerConfig("nonblocking")

  // init table
  tall := table.NewWriter()
  tall.SetOutputMirror(os.Stdout)
  tall.AppendHeader(table.Row{"Bug","Delay Bound","1st Fail","Cov1 Leaps", "Cov2 Leaps","Errors","Cov1 > Cov2","Final Coverage"})

  for _,path := range(ReadLines(configFile)){
    paths, err := filepath.Glob(path)
    check(err)
    // iterate over each bug
    for _,p := range(paths){
      // extract bug info
      fmt.Println("p: ",p)
      bugInfo := strings.Split(instrument.GobenchAppNameFolder(p),"_") // bugType_BugAppName_BugCommitID
      bugType := bugInfo[0]
      bugName := bugInfo[1] + "_" + bugInfo[2]
      bugFullName := bugInfo[0] + "_" + bugInfo[1] + "_" + bugInfo[2]

      if bugName == ""{
        panic("wrong bugName")
      }
      mainExp := &RootExperiment{}
      // create bug now
      fmt.Println("BugName:",bugName,", path: ",p,", fullname: ",bugFullName)

      target := &Bug{bugName,p,bugType,causes[bugType][bugName][0],causes[bugType][bugName][1]}
      mainExp.Bug = target
      mainExp.Exps = make(map[string]Ex)

      for d,exp := range(dx){
        // figure exp report file
        ws := os.Getenv("GOATWS")
        if ws == "" {
          panic("GOATWS is not set!")
        }
        // we have to re-discover gex prefixdir to see if we have ready results
        expReport := ws + "/"
        expReport = expReport + "p"+MAXPROCS+"/"
        expReport = expReport + bugFullName + "/"

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
        expReport = expReport + "p"+MAXPROCS+"_"+bugFullName+"_"+exp+"_T"+strconv.Itoa(thresh)+"_"+TERMINATION+".json"

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

iteration:for i:=0 ; i < thresh ; i++{
            fmt.Printf("Test %v on %v (%d/%d)\n",gex.Target.BugName,gex.ID,i+1,thresh)
            res := gex.Execute(i,isRace, json_trace)
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
              if MAXPROCS == "1" && i > 200 && res.EventsLen > 3000000 {
                break iteration
              }
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
        mainExp.Exps[gex.ID] = gex
        tall.AppendRow(CoverageSummary(gex))
        tall.AppendSeparator()
        tall.RenderCSV()
        tall.Render()

      }
      allBugs[bugFullName] = mainExp
    } // end of inner paths
  }// end of config file

   for _,td := range(dx){
     fmt.Printf("*******\nTool: %v (cov1)\n*******",td)
     Table_Bug_Coverage(allBugs,td,thresh,true)
   }
  //
   for _,td := range(dx){
     fmt.Printf("*******\nTool: %v (cov2)\n*******",td)
     Table_Bug_Coverage(allBugs,td,thresh,false)
   }

}

func EvaluateComparison(configFile string, thresh int, json_trace bool){
  var ex Ex
  colorReset := "\033[0m"
  colorRed := "\033[31m"
  colorGreen := "\033[32m"

  dx := comparison

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

  // a map to hold each RootExperiment
  allBugs := make(map[string]*RootExperiment)

  // read causes
  causes := make(map[string]map[string][]string)
  causes["blocking"] = ReadGoKerConfig("blocking")
  causes["nonblocking"] = ReadGoKerConfig("nonblocking")

  for _,path := range(ReadLines(configFile)){
    paths, err := filepath.Glob(path)
    check(err)
    // iterate over each bug
    for _,p := range(paths){
      // extract bug info
      fmt.Println("p: ",p)
      bugInfo := strings.Split(instrument.GobenchAppNameFolder(p),"_") // bugType_BugAppName_BugCommitID
      bugType := bugInfo[0]
      bugName := bugInfo[1] + "_" + bugInfo[2]
      bugFullName := bugInfo[0] + "_" + bugInfo[1] + "_" + bugInfo[2]

      if bugName == ""{
        panic("wrong bugName")
      }
      mainExp := &RootExperiment{}
      // create bug now
      fmt.Println("BugName:",bugName,", path: ",p,", fullname: ",bugFullName)

      target := &Bug{bugName,p,bugType,causes[bugType][bugName][0],causes[bugType][bugName][1]}
      mainExp.Bug = target
      mainExp.Exps = make(map[string]Ex)

      for _,exp := range(dx){
        // figure exp report file
        ws := os.Getenv("GOATWS")
        if ws == "" {
          panic("GOATWS is not set!")
        }
        // we have to re-discover ex prefixdir to see if we have ready results
        expReport := ws + "/"
        expReport = expReport + "p"+MAXPROCS+"/"
        expReport = expReport + bugFullName + "/"
        if strings.HasPrefix(exp,"goat_d"){
          d,err := strconv.Atoi(strings.Split(exp,"_d")[1])
          check(err)
          if d < 1 {
            expReport = expReport + "goat_trace/"
          } else{
            expReport = expReport + "goat_delay/"
          }
          ex = &GoatExperiment{Experiment: Experiment{Target:target},Bound:d}
        } else if strings.HasPrefix(exp,"goat_race"){
          d,err := strconv.Atoi(strings.Split(exp,"_d")[1])
          check(err)
          if d < 1 {
            expReport = expReport + "goat_race_trace/"
          } else{
            expReport = expReport + "goat_race_delay/"
          }
          ex = &GoatExperiment{Experiment: Experiment{Target:target},Bound:d}
        }else{
          ex = &ToolExperiment{Experiment: Experiment{Target:target},ToolID:exp}
          expReport = expReport + exp + "/"
        }

        expReport = expReport + "results/"
        expReport = expReport + "p"+MAXPROCS+"_"+bugFullName+"_"+exp+"_T"+strconv.Itoa(thresh)+"_"+TERMINATION+".json"

        if checkFile(expReport){ // if report exist
          switch ex.(type){
          case *GoatExperiment:
            gex := ex.(*GoatExperiment)
            gex = ReadExperimentResults_goat(expReport)
            gex.Target = target

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

            // missing execute stuff (from results)
            newResults := []*Result{}
            // iterate over results
            for _,res := range(gex.Results){
              result := res
              // get the local stack
              if res.Desc == "CRASH"|| res.Desc == "NONE"{
                // it does not have any trace
                newResults = append(newResults,result)
                continue
              }
              fmt.Println("Reading Trace ",result.TracePath)
              fmt.Println("Result Desc ",result.Desc)
              parseRes := traceops.ReadParseTrace(result.TracePath, filepath.Join(gex.PrefixDir,"bin",gex.BinaryName))

              // print events
              for _,e := range(parseRes.Events){
                fmt.Println(e)
                fmt.Println("-----------------------------------------------------------")
              }

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
            mainExp.Exps[gex.ID] = gex
          case *ToolExperiment:
            tex := ex.(*ToolExperiment)
            tex = ReadExperimentResults_tool(expReport)
            mainExp.Exps[tex.ToolID] = tex
          }
        } else{
          ex.Init(false)
          ex.Instrument()
          ex.Build(false)
          switch ex.(type){
          case *GoatExperiment:
            gex := ex.(*GoatExperiment)
            iteration:for i:=0 ; i < thresh ; i++{
              fmt.Printf("Test %v on %v (%d/%d)\n",gex.Target.BugName,gex.ID,i+1,thresh)
              res := gex.Execute(i,false, json_trace)
              gex.Results = append(gex.Results,res)
              if res.Detected{
              	fmt.Println(string(colorRed),res.Desc,string(colorReset))
                switch TERMINATION {
                case "hitBug":
                  break iteration
                default:
                }
              }else{
                fmt.Println(string(colorGreen),"PASS",string(colorReset))
              }
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

          case *ToolExperiment:
            tex := ex.(*ToolExperiment)
            iteration2:for i:=0 ; i < thresh ; i++{
              if tex.Target.BugName == "etcd_7492" && (tex.ToolID == "builtinDL" || tex.ToolID == "goleak"){
                res := &Result{Detected:true,Desc:"HANG"}
                tex.Results = append(tex.Results,res)
                fmt.Println(string(colorRed),res.Desc,string(colorReset))
                break iteration2
              }
              fmt.Printf("Test %v on %v (%d/%d)\n",tex.Target.BugName,tex.ToolID,i+1,thresh)
              res := tex.Execute(i,false, json_trace)
              tex.Results = append(tex.Results,res)
              if res.Detected{
              	fmt.Println(string(colorRed),res.Desc,string(colorReset))
                switch TERMINATION {
                case "hitBug":
                  break iteration2
                default:
                }
              }else{
                fmt.Println(string(colorGreen),"PASS",string(colorReset))
              }
            }
            mainExp.Exps[tex.ToolID] = tex
            // store tex into json
            // check if expReport path matches tex.PrefixDir
            if filepath.Dir(expReport) != tex.PrefixDir+"/results" {
              panic(fmt.Sprintf("mismatch expReport dir\n%v\n%v\n)",filepath.Dir(expReport),tex.PrefixDir+"/results"))
            }
            // write to json file
            rep,err := os.Create(expReport)
            check(err)
            newdat ,err := json.MarshalIndent(tex,"","    ")
            check(err)
            _,err = rep.WriteString(string(newdat))
            check(err)
            rep.Close()
          }
        } // gex/tex either executed or read results from json file - mainExp.Exps[tool] assigned
      } // end of loop over tools
      TableSummaryPerBug(mainExp)
      allBugs[bugFullName] = mainExp
    } // end of inner paths
  }// end of config file
  fmt.Println("Total Bugs: ",len(allBugs))

  Table_Bug_Tool(allBugs,ORDER_BUG,"blocking")
  Table_Bug_Tool(allBugs,ORDER_CAUSE,"blocking")
  Table_Bug_Tool(allBugs,ORDER_SUBCAUSE,"blocking")

  // if isRace {
  //   Table_Bug_Tool(allBugs,ORDER_BUG,"nonblocking")
  //   Table_Bug_Tool(allBugs,ORDER_CAUSE,"nonblocking")
  //   Table_Bug_Tool(allBugs,ORDER_SUBCAUSE,"nonblocking")
  // }

}

func EvaluateNonBlocking(configFile string,thresh int, json_trace bool) {
  identifier := "nonblocking"
  causes := ReadGoKerConfig(identifier)

  colorReset := "\033[0m"
  colorRed := "\033[31m"
  colorGreen := "\033[32m"

  // a map to hold each RootExperiment
  allBugs := make(map[string]*RootExperiment)

  // obtain configName
  configName := strings.Split(filepath.Base(configFile),".")[0]

  // obtain result dir
  reportDir := filepath.Join(RESDIR,identifier+"_"+configName+"_"+strconv.Itoa(thresh))
  err := os.MkdirAll(reportDir,os.ModePerm)
  check(err)

  for _,path := range(ReadLines(configFile)){
    paths, err := filepath.Glob(path)
    check(err)
    // iterate over each bug
    for _,p := range(paths){
      // extract bug info
      bugFullName := instrument.GobenchAppNameFolder(p) // bugType_BugAppName_BugCommitID
      bugName := strings.Split(bugFullName,identifier+"_")[1]

      mainExp := &RootExperiment{}
      // create bug now
      target := &Bug{bugName,p,identifier,causes[bugName][0],causes[bugName][1]}
      mainExp.Bug = target
      mainExp.Exps = make(map[string]Ex)
      mainExp.ReportJSON = fmt.Sprintf("%s/%s_%s_T%v.json",reportDir,identifier,bugName,thresh)


      if checkFile(mainExp.ReportJSON){ // if report exist
        // create experiment from json files
        mainExp.Exps = ReadResults(mainExp.ReportJSON)
        fmt.Println("File Found")
      } else{
        var exes []Ex
        //exes := []interface{}{}
        exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:1})
        exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:2})
        exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:3})
        exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:4})
        exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:5})
        exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:6})
        exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:7})
        exes = append(exes,&ToolExperiment{Experiment: Experiment{Target:target},ToolID:"race"})
        for _,ex := range(exes){
          // pre-set
          ex.Init(true)
          ex.Instrument()
          ex.Build(true)
          IDD := ""
          iteration:for i:=0 ; i < thresh ; i++{
            switch ex.(type){
            case *GoatExperiment:
              gex := ex.(*GoatExperiment)
              IDD = gex.ID
              fmt.Printf("Test %v on %v (%d/%d)\n",gex.Target.BugName,gex.ID,i+1,thresh)
              res := gex.Execute(i,true, json_trace)
              gex.Results = append(gex.Results,res)
              if res.Detected{
              	fmt.Println(string(colorRed),res.Desc,string(colorReset))
                break iteration
              } else if res.Desc != ""{ // race not detected but program has paniced
                fmt.Println(string(colorRed),res.Desc,string(colorReset))
                break iteration
              } else {
                fmt.Println(string(colorGreen),"PASS",string(colorReset))
              }
            case *ToolExperiment:
              tex := ex.(*ToolExperiment)
              IDD = tex.ToolID
              fmt.Printf("Test %v on %v (%d/%d)\n",tex.Target.BugName,tex.ToolID,i+1,thresh)
              res := tex.Execute(i,true, json_trace)
              tex.Results = append(tex.Results,res)
              if res.Detected{
              	fmt.Println(string(colorRed),res.Desc,string(colorReset))
                break iteration
              }else{
                fmt.Println(string(colorGreen),"PASS",string(colorReset))
              }
            }
          }
          mainExp.Exps[IDD] = ex
        }
        rep,err := os.Create(mainExp.ReportJSON)
        check(err)
        newdat ,err := json.MarshalIndent(mainExp,"","    ")
        check(err)
        _,err = rep.WriteString(string(newdat))
        check(err)
        rep.Close()
      }
      // all experiments are either done or read from file
      TableSummaryPerBug(mainExp)
      allBugs[bugName] = mainExp
    }
  }
  fmt.Println("Total Bugs: ",len(allBugs))
  Table_Bug_Tool(allBugs,ORDER_BUG,identifier)
  Table_Bug_Tool(allBugs,ORDER_CAUSE,identifier)
  Table_Bug_Tool(allBugs,ORDER_SUBCAUSE,identifier)
}
