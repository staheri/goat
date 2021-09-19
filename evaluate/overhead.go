package evaluate

import(
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
  "encoding/json"
  "github.com/staheri/goatlib/instrument"
  "strings"
  "strconv"
  "github.com/jedib0t/go-pretty/table"
  "time"
)



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
  cmd := exec.Command("ln","-nsf",GOVER_GOAT,os.Getenv("GOROOT"))
  err = cmd.Run()
  check(err)
}
