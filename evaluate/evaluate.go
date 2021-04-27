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



func EvaluateBlocking(configFile string, thresh int) {
  identifier := "blocking"
  causes := ReadGoKerConfig(identifier)

  colorReset := "\033[0m"
  colorRed := "\033[31m"
  colorGreen := "\033[32m"


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
      SummaryTable(mainExp)
    }

  }
}


func EvaluateNonBlocking(configFile string,thresh int) {
  identifier := "nonblocking"
  causes := ReadGoKerConfig(identifier)

  colorReset := "\033[0m"
  colorRed := "\033[31m"
  colorGreen := "\033[32m"

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
      SummaryTable(mainExp)
    }

  }
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
          IDD := fmt.Sprintf("%v_%v_i%v",ex.Target.BugName,ex.ID,strings.Join(ex.Args,"_"))
          for i:=0 ; i < thresh ; i++{
            fmt.Printf("Test %v on %v (input:%s) (%d/%d)\nIDD:%v\n",ex.Target.BugName,ex.ID,strings.Join(ex.Args,"_"),i+1,thresh,IDD)
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
    	t.AppendHeader(table.Row{"Experiment","Input","#GRTN","#CHNL","#EVENTS","OVERHEAD"})

      var totalTimeNative  time.Duration
      var totalTimeET      time.Duration
      var totalTimeECT     time.Duration
      var totalEvET        int
      var totalEvECT       int
      var totalGECT        int
      var totalGET         int
      var totalChET        int
      var totalChECT       int

      for _,n := range(ns){
        key := fmt.Sprintf("%v_%v_i%v",target.BugName,"ECT_native",strconv.Itoa(n))
        fmt.Printf("KEY: %v\n",key)
        for _,r:= range(mainExp.Exps[key].(*ECTExperiment).Results){
          totalTimeNative = totalTimeNative + r.Time
        }
        key = fmt.Sprintf("%v_%v_i%v",target.BugName,"ECT_ET",strconv.Itoa(n))
        for _,r:= range(mainExp.Exps[key].(*ECTExperiment).Results){
          totalTimeET = totalTimeET + r.Time
          totalEvET = totalEvET + r.EventsLen
          totalGET = totalGET + r.TotalG
          totalChET = totalChET + r.TotalCh
        }
        key = fmt.Sprintf("%v_%v_i%v",target.BugName,"ECT_ECT",strconv.Itoa(n))
        for _,r:= range(mainExp.Exps[key].(*ECTExperiment).Results){
          totalTimeECT = totalTimeECT + r.Time
          totalEvECT = totalEvECT + r.EventsLen
          totalGECT = totalGECT + r.TotalG
          totalChECT = totalChECT + r.TotalCh
        }
        var rowET       []interface{}
        var rowECT       []interface{}
        //key = fmt.Sprintf("%v_%v_i%v)",target.BugName,"ET",strconv.Itoa(n))
        rowET = append(rowET,target.BugName+"_ET")
        rowET = append(rowET,n)
        rowET = append(rowET,totalGET/thresh)
  			rowET = append(rowET,totalChET/thresh)
  			rowET = append(rowET,totalEvET/thresh)
  			rowET = append(rowET,(float64(totalTimeET.Milliseconds())/float64(thresh))/(float64(totalTimeNative.Milliseconds())/float64(thresh)))
        t.AppendRow(rowET)
        rowECT = append(rowECT,target.BugName+"_ECT")
        rowECT = append(rowECT,n)
        rowECT = append(rowECT,totalGECT/thresh)
  			rowECT = append(rowECT,totalChECT/thresh)
  			rowECT = append(rowECT,totalEvECT/thresh)
  			rowECT = append(rowECT,(float64(totalTimeECT.Milliseconds())/float64(thresh))/(float64(totalTimeNative.Milliseconds())/float64(thresh)))
        t.AppendRow(rowECT)
      }
      t.Render()
    }
  }

  // change link of GO
  cmd := exec.Command("sudo","ln","-nsf","/usr/local/myGo.1.15.6/","/usr/local/go")
  err = cmd.Run()
  check(err)
}
