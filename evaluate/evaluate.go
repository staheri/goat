// Main file for evaluation
package evaluate

import(
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
  "encoding/json"
  "github.com/staheri/goatlib/instrument"
  _"github.com/staheri/goatlib/traceops"
  "strings"
  "strconv"
  "github.com/jedib0t/go-pretty/table"
  "time"
)


type RootExperiment struct{
  Bug         *Bug
  Exps        map[string]Ex      `json:"exps"`
  ReportJSON  string             `json:"reportJSON"`
}


func EvaluateCoverage(configFile string, thresh int) {
  identifier := "blocking"
  causes := ReadGoKerConfig(identifier)
  //fmt.Println(causes)

  colorReset := "\033[0m"
  colorRed := "\033[31m"
  colorGreen := "\033[32m"

  // a map to hold each RootExperiment
  //allBugs := make(map[string]*RootExperiment)


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
      if bugName == ""{
        panic("wrong bugName")
      }
      mainExp := &RootExperiment{}
      // create bug now
      fmt.Println("BugName:",bugName,", path: ",p,", fullname: ",bugFullName)
      target := &Bug{bugName,p,identifier,causes[bugName][0],causes[bugName][1]}
      mainExp.Bug = target
      mainExp.Exps = make(map[string]Ex)

      var exes []Ex
      //exes := []interface{}{}
      //exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:-1})
      exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:6})

      for _,ex := range(exes){
        // pre-set
        ex.Init(false)
        ex.Instrument()
        // after instrument, we have the concusage and concusageMap
        ex.Build(false)
        IDD := ""
        iteration:for i:=0 ; i < thresh ; i++{
          switch ex.(type){
          case *GoatExperiment:
            gex := ex.(*GoatExperiment)
            IDD = gex.ID
            fmt.Printf("Test %v on %v (%d/%d)\n",gex.Target.BugName,gex.ID,i+1,thresh)
            res := gex.Execute(i,false)
            // after the first execute, we can say
            gex.Results = append(gex.Results,res)
            // Update Coverage
            //gex.Coverage.Update(res.CoverageReport)

            if res.Detected{
            	fmt.Println(string(colorRed),res.Desc,string(colorReset))
            }else{
              fmt.Println(string(colorGreen),"PASS",string(colorReset))
            }
            //time.Sleep(5*time.Second)
          case *ToolExperiment:
            tex := ex.(*ToolExperiment)
            IDD = tex.ToolID
            fmt.Printf("Test %v on %v (%d/%d)\n",tex.Target.BugName,tex.ToolID,i+1,thresh)
            res := tex.Execute(i,false)
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
    }
  }
}

func EvaluateBlocking(configFile string, thresh int) {
  identifier := "blocking"
  causes := ReadGoKerConfig(identifier)
  fmt.Println(causes)

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
      if bugName == ""{
        panic("wrong bugName")
      }
      mainExp := &RootExperiment{}
      // create bug now
      fmt.Println("BugName:",bugName,", path: ",p,", fullname: ",bugFullName)
      target := &Bug{bugName,p,identifier,causes[bugName][0],causes[bugName][1]}
      mainExp.Bug = target
      mainExp.Exps = make(map[string]Ex)
      mainExp.ReportJSON = fmt.Sprintf("%s/%s_%s_T%v.json",reportDir,identifier,bugName,thresh)


      if checkFile(mainExp.ReportJSON){ // if report exist
        // create experiment from json files
        mainExp.Exps = ReadResults(mainExp.ReportJSON)
        fmt.Println("File Found")
        // we want to replay what has happened below
        // iterate over exps
        // obtain results
        /*for _,ex := range(mainExp.Exps){
          iteration_replay:for i:=0 ; i < thresh ; i++{
            switch ex.(type){
            case *GoatExperiment:
              gex := ex.(*GoatExperiment)
              // assert i <= len gex.Results
              if len(gex.Results) <= i{
                panic("inconsistent replay")
              }
              res := gex.Results[i]
              // lets not rely on previous experiment
              // double check the trace
              traceops.ReplayDeadlockChecker(res.TracePath, filepath.Join(gex.PrefixDir,"bin",gex.BinaryName))
              if res.Detected{
              	fmt.Println(string(colorRed),res.Desc,string(colorReset))
                break iteration_replay
              }else{
                fmt.Println(string(colorGreen),"PASS",string(colorReset))
              }
            case *ToolExperiment:
              tex := ex.(*ToolExperiment)
              // assert i <= len gex.Results
              if len(tex.Results) <= i{
                panic("inconsistent replay")
              }
              res := tex.Results[i]
              if res.Detected{
              	fmt.Println(string(colorRed),res.Desc,string(colorReset))
                break iteration_replay
              }else{
                fmt.Println(string(colorGreen),"PASS",string(colorReset))
              }
            }
          }
        }*/
      } else{
        var exes []Ex
        //exes := []interface{}{}
        exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:-1})
        exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:0})
        exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:1})
        exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:2})
        exes = append(exes,&GoatExperiment{Experiment: Experiment{Target:target},Bound:3})
        exes = append(exes,&ToolExperiment{Experiment: Experiment{Target:target},ToolID:"lockDL"})
        exes = append(exes,&ToolExperiment{Experiment: Experiment{Target:target},ToolID:"goleak"})
        exes = append(exes,&ToolExperiment{Experiment: Experiment{Target:target},ToolID:"builtinDL"})
        for _,ex := range(exes){
          // pre-set
          ex.Init(false)
          ex.Instrument()
          ex.Build(false)
          IDD := ""
          iteration:for i:=0 ; i < thresh ; i++{
            switch ex.(type){
            case *GoatExperiment:
              gex := ex.(*GoatExperiment)
              IDD = gex.ID
              fmt.Printf("Test %v on %v (%d/%d)\n",gex.Target.BugName,gex.ID,i+1,thresh)
              res := gex.Execute(i,false)
              gex.Results = append(gex.Results,res)
              if res.Detected{
              	fmt.Println(string(colorRed),res.Desc,string(colorReset))
                break iteration
              }else{
                fmt.Println(string(colorGreen),"PASS",string(colorReset))
              }
            case *ToolExperiment:
              tex := ex.(*ToolExperiment)
              IDD = tex.ToolID
              fmt.Printf("Test %v on %v (%d/%d)\n",tex.Target.BugName,tex.ToolID,i+1,thresh)
              res := tex.Execute(i,false)
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

func EvaluateNonBlocking(configFile string,thresh int) {
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
              res := gex.Execute(i,true)
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
              res := tex.Execute(i,true)
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

func EvaluateOverhead(configFile string, thresh int, ns []int) {
  identifier := "overhead"

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
      fmt.Println("PATH:",p)
      // extract bug info
      bugFullName := instrument.GobenchAppNameFolder(p) // bugType_BugAppName_BugCommitID
      bugName := strings.Split(bugFullName,identifier+"_")[1]
      // create bug now
      target := &Bug{BugName:bugName,BugDir:p,BugType:identifier}

      mainExp := &RootExperiment{}
      mainExp.Bug = target
      mainExp.Exps = make(map[string]Ex)
      mainExp.ReportJSON = fmt.Sprintf("%s/%s_%s_T%v.json",reportDir,identifier,bugName,thresh)


      if checkFile(mainExp.ReportJSON){ // if report exist
        // create experiment from json files
        mainExp.Exps = ReadResults(mainExp.ReportJSON)
        for k,ex:= range(mainExp.Exps){
          fmt.Printf("checkFile Key Exp: %v\n",k)
          fmt.Printf("checkFile Result: %v\n",ex.(*ECTExperiment).Results)
        }
        fmt.Println("File Found")
      } else{
        var exes []Ex
        //exes := []interface{}{}
        for _,n := range(ns){
          exes = append(exes,&ECTExperiment{Experiment: Experiment{Target:target},ID:"ECT_native",Args:[]string{strconv.Itoa(n)}})
          exes = append(exes,&ECTExperiment{Experiment: Experiment{Target:target},ID:"ECT_ET",Args:[]string{strconv.Itoa(n)}})
          exes = append(exes,&ECTExperiment{Experiment: Experiment{Target:target},ID:"ECT_ECT",Args:[]string{strconv.Itoa(n)}})
        }
        for _,exx := range(exes){
          // pre-set
          ex := exx.(*ECTExperiment)
          ex.Init(false)  // race = false
          ex.Instrument() // race = false
          ex.Build(false) // race = false
          IDD := fmt.Sprintf("%v_%v_i%v",ex.ID,ex.Target.BugName,strings.Join(ex.Args,"_"))
          for i:=0 ; i < thresh ; i++{
            fmt.Printf("Test %v on %v (input:%s) (%d/%d)\nIDD:%v\n",ex.ID,ex.Target.BugName,strings.Join(ex.Args,"_"),i+1,thresh,IDD)
            //IDD = fmt.Sprintf("%v_%v_input:%s) (%d/%d)\n",ex.Target.BugName,ex.ID,strings.Join(ex.Args,"_"),i+1,thresh)
            res := ex.Execute(i,false)
            fmt.Printf("\tTime: %v\n",res.Time)
            ex.Results = append(ex.Results,res)
          }
          fmt.Println(ex)
          fmt.Printf("Add ex to mainExp.Exps[%v]\n",IDD)
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
      // calculate average time of native
      t := table.NewWriter()
    	t.SetOutputMirror(os.Stdout)
    	t.AppendHeader(table.Row{"Experiment","Input","#GRTN","#CHNL","#EVENTS","Trace Size(B)","OVERHEAD"})
      //
      // var totalTimeNative  time.Duration
      // var totalTimeET      time.Duration
      // var totalTimeECT     time.Duration
      // var totalEvET        int
      // var totalEvECT       int
      // var totalTraceET     int
      // var totalTraceECT    int
      // var totalGECT        int
      // var totalGET         int
      // var totalChET        int
      // var totalChECT       int

      for _,n := range(ns){
        var totalTimeNative  time.Duration
        var totalTimeET      time.Duration
        var totalTimeECT     time.Duration
        var totalEvET        int
        var totalEvECT       int
        var totalTraceET     int
        var totalTraceECT    int
        var totalGECT        int
        var totalGET         int
        var totalChET        int
        var totalChECT       int
        key := fmt.Sprintf("%v_%v_i%v","ECT_native",target.BugName,strconv.Itoa(n))
        fmt.Printf("KEY_Native: %v\n",key)
        for _,r:= range(mainExp.Exps[key].(*ECTExperiment).Results){
          totalTimeNative = totalTimeNative + r.Time
          fmt.Printf("\tResult: %v\n",r)
        }
        key = fmt.Sprintf("%v_%v_i%v","ECT_ET",target.BugName,strconv.Itoa(n))
        fmt.Printf("KEY_ET: %v\n",key)
        for _,r:= range(mainExp.Exps[key].(*ECTExperiment).Results){
          totalTimeET = totalTimeET + r.Time
          totalEvET = totalEvET + r.EventsLen
          totalGET = totalGET + r.TotalG
          totalChET = totalChET + r.TotalCh
          totalTraceET = totalTraceET + r.TraceSize
          fmt.Printf("\tResult: %v\n",r)
        }
        fmt.Printf("KEY_ECT: %v\n",key)
        key = fmt.Sprintf("%v_%v_i%v","ECT_ECT",target.BugName,strconv.Itoa(n))
        for _,r:= range(mainExp.Exps[key].(*ECTExperiment).Results){
          totalTimeECT = totalTimeECT + r.Time
          totalEvECT = totalEvECT + r.EventsLen
          totalGECT = totalGECT + r.TotalG
          totalChECT = totalChECT + r.TotalCh
          totalTraceECT = totalTraceECT + r.TraceSize
          fmt.Printf("\tResult: %v\n",r)
        }
        var rowET       []interface{}
        var rowECT       []interface{}
        //key = fmt.Sprintf("%v_%v_i%v)",target.BugName,"ET",strconv.Itoa(n))
        rowET = append(rowET,target.BugName+"_ET")
        rowET = append(rowET,n)
        rowET = append(rowET,float64(totalGET)/float64(thresh))
  			rowET = append(rowET,float64(totalChET)/float64(thresh))
  			rowET = append(rowET,totalEvET/thresh)
        rowET = append(rowET,totalTraceET/thresh)
  			rowET = append(rowET,(float64(totalTimeET.Milliseconds())/float64(thresh))/(float64(totalTimeNative.Milliseconds())/float64(thresh)))
        t.AppendRow(rowET)
        rowECT = append(rowECT,target.BugName+"_ECT")
        rowECT = append(rowECT,n)
        rowECT = append(rowECT,totalGECT/thresh)
  			rowECT = append(rowECT,totalChECT/thresh)
  			rowECT = append(rowECT,totalEvECT/thresh)
        rowECT = append(rowECT,totalTraceECT/thresh)
  			rowECT = append(rowECT,(float64(totalTimeECT.Milliseconds())/float64(thresh))/(float64(totalTimeNative.Milliseconds())/float64(thresh)))
        t.AppendRow(rowECT)
      }
      t.Render()
      t.RenderCSV()



      // calculate average time of native

      t = table.NewWriter()
    	t.SetOutputMirror(os.Stdout)
    	t.AppendHeader(table.Row{"Experiment","Input","i","#GRTN","#CHNL","#EVENTS","Trace Size(B)","Time"})

      for _,n := range(ns){
        key := fmt.Sprintf("%v_%v_i%v","ECT_native",target.BugName,strconv.Itoa(n))
        ex := mainExp.Exps[key].(*ECTExperiment)
        for i,r:= range(ex.Results){
          var row []interface{}
          row = append(row,target.BugName+"_native")
          row = append(row,n)
          row = append(row,i)
          row = append(row,"n/a")
          row = append(row,"n/a")
          row = append(row,"n/a")
          row = append(row,"n/a")
          row = append(row,r.Time)
          t.AppendRow(row)
        }
        key = fmt.Sprintf("%v_%v_i%v","ECT_ET",target.BugName,strconv.Itoa(n))
        ex = mainExp.Exps[key].(*ECTExperiment)
        for i,r:= range(ex.Results){
          var row []interface{}
          row = append(row,target.BugName+"_ET")
          row = append(row,n)
          row = append(row,i)
          row = append(row,r.TotalG)
          row = append(row,r.TotalCh)
          row = append(row,r.EventsLen)
          row = append(row,r.TraceSize)
          row = append(row,r.Time)
          t.AppendRow(row)
          //fmt.Printf("%v (%v): TOTAL G: %v\n",key,i,r.TotalG)
          //traceops.ReplayDispGMAP(r.TracePath, filepath.Join(ex.PrefixDir,"bin",ex.BinaryName))
        }

        key = fmt.Sprintf("%v_%v_i%v","ECT_ECT",target.BugName,strconv.Itoa(n))
        ex = mainExp.Exps[key].(*ECTExperiment)
        for i,r:= range(ex.Results){
          var row []interface{}
          row = append(row,target.BugName+"_ECT")
          row = append(row,n)
          row = append(row,i)
          row = append(row,r.TotalG)
          row = append(row,r.TotalCh)
          row = append(row,r.EventsLen)
          row = append(row,r.TraceSize)
          row = append(row,r.Time)
          t.AppendRow(row)
          //fmt.Printf("%v (%v): TOTAL G: %v\n",key,i,r.TotalG)
          //traceops.ReplayDispGMAP(r.TracePath, filepath.Join(ex.PrefixDir,"bin",ex.BinaryName))
        }
      }
      //t.Render()

    }
  }

  // change link of GO
  cmd := exec.Command("sudo","ln","-nsf","/usr/local/myGo.1.15.6/","/usr/local/go")
  err = cmd.Run()
  check(err)
}
