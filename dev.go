package main

import(
  "github.com/staheri/goatlib/instrument"
  "github.com/staheri/goatlib/traceops"
  "github.com/staheri/goat/evaluate"
  "fmt"
  "os"
  //"strconv"
  "strings"
  //"bufio"
  "path/filepath"
)

/*func experiment(path string,iter int,w *bufio.Writer) map[int][]db.Report{
  // for each test, execute ITER times with only trace
  // then execute ITER times with concurrency usage and depth 0
  // then execute ITER times with concurrency usage and depth 1
  // then execute ITER times with concurrency usage and depth 2
  // then execute ITER times with concurrency usage and depth 3
  // iapp is the instrumented version of target program

  ret := make(map[int][]db.Report)
  var passCnt, failCnt int
  var reports []db.Report
  iapp := instrument.Instrument(path,true) // traceOnly: true
  for i := 0 ; i < iter ; i++{
    dbName := fmt.Sprintf("%s_B%d_I%d",iapp.Name,0,i)
    // execute
    events,err := iapp.ExecuteTrace()
    er := handle(err)
    if er != ""{
      _,err = fmt.Fprintf(w,"%s,%s,%s\n",iapp.Name,dbName,"crash")
      failCnt++
      continue
    }
    // store
    dbi := db.Store(events, dbName)
    fmt.Printf("B:0 - test %d/%d (%s)\n",i+1,iter,dbName)
    report := db.Checker(dbi) // longReport = false
    dbi.Close()
    if report.GlobalDL {
      _,err = fmt.Fprintf(w,"%s,%s,%s\n",iapp.Name,dbName,"fail,gdl")
      failCnt++
    } else if report.Leaked != 0{
      _,err = fmt.Fprintf(w,"%s,%s,%s,%d\n",iapp.Name,dbName,"fail,pdl",report.Leaked)
      failCnt++
    }else{
      _,err = fmt.Fprintf(w,"%s,%s,%s\n",iapp.Name,dbName,"pass")
      passCnt++
    }
    reports = append(reports,report)
  }
  fmt.Printf("Pass/Fail: %d/%d\n",passCnt,failCnt)
  ret[0]=reports

  // identify concusage and instrument
  iapp = instrument.Instrument(path,false) // traceOnly: false

  for b:= 1 ; b<4 ; b++{
    reports = nil
    passCnt = 0
    failCnt = 0
    for i := 0 ; i < iter ; i++{
      // set bound
      os.Setenv("GOATRSBOUND",strconv.Itoa(b))

      dbName := fmt.Sprintf("%s_B%d_I%d",iapp.Name,b,i)

      // execute
      events,err := iapp.ExecuteTrace()
      er := handle(err)
      if er != ""{
        _,err = fmt.Fprintf(w,"%s,%s,%s\n",iapp.Name,dbName,"crash")
        failCnt++
        continue
      }
      // store

      dbi := db.Store(events, dbName)
      fmt.Printf("B:%d - test %d/%d (%s)\n",b,i+1,iter,dbName)
      report := db.Checker(dbi) // longReport = false
      dbi.Close()
      if report.GlobalDL {
        _,err = fmt.Fprintf(w,"%s,%s,%s\n",iapp.Name,dbName,"fail,gdl")
        failCnt++
      } else if report.Leaked != 0{
        _,err = fmt.Fprintf(w,"%s,%s,%s,%d\n",iapp.Name,dbName,"fail,pdl",report.Leaked)
        failCnt++
      }else{
        _,err = fmt.Fprintf(w,"%s,%s,%s\n",iapp.Name,dbName,"pass")
        passCnt++
      }
      reports = append(reports,report)
    }
    fmt.Printf("Pass/Fail: %d/%d\n",passCnt,failCnt)
    ret[b]=reports
  }
  return ret
}*/



// generate the pdfs from success and fail trace
func customVis(bugpath, tool_id string, withStack bool){
  reportFolder := "results/blocking_conf_blocking_all_1000"
  visualFolder := "/Volumes/DATA/goatws/visuals"
  //outFolder := "execVis"
  identifier := "blocking"

  causes := evaluate.ReadGoKerConfig(identifier)
  //fmt.Println(causes)
  // extract bug info
  bugFullName := instrument.GobenchAppNameFolder(bugpath) // bugType_BugAppName_BugCommitID
  bugName := strings.Split(bugFullName,identifier+"_")[1]
  if bugName == ""{
    panic("wrong bugName")
  }
  mainExp := &evaluate.RootExperiment{}
  // create bug now
  fmt.Println("BugName:",bugName,", path: ",bugpath,", fullname: ",bugFullName)
  target := &evaluate.Bug{bugName,bugpath,identifier,causes[bugName][0],causes[bugName][1]}
  mainExp.Bug = target
  mainExp.Exps = make(map[string]evaluate.Ex)
  mainExp.ReportJSON = fmt.Sprintf("%s/blocking_%s_T100.json",reportFolder,bugName)
  // create experiment from json files
  if !checkFile(mainExp.ReportJSON){
    panic("unable to read json report")
  }
  mainExp.Exps = evaluate.ReadResults(mainExp.ReportJSON)
  fmt.Println("File Found")

  gex := mainExp.Exps[tool_id].(*evaluate.GoatExperiment)
  if gex.LastFailedTrace != ""{
    outp := fmt.Sprintf("%v/FAIL_%v",visualFolder,strings.Split(filepath.Base(gex.LastFailedTrace),".trace")[0])
    traceops.ExecVis(gex.LastFailedTrace,filepath.Join(gex.PrefixDir,"bin",gex.BinaryName),outp,withStack)
  }
  if gex.LastSuccessTrace != ""{
    outp := fmt.Sprintf("%v/SUCC_%v",visualFolder,strings.Split(filepath.Base(gex.LastSuccessTrace),".trace")[0])
    traceops.ExecVis(gex.LastSuccessTrace,filepath.Join(gex.PrefixDir,"bin",gex.BinaryName),outp,withStack)
  }


  // find the binary file
  // find the trace file (failed)
  // find the trace file (success)
  // create execVis for them

}


func check(err error){
	if err != nil{
		panic(err)
	}
}

func handle(err error) string{
  if err != nil{
    fmt.Println(err)
    s := fmt.Sprintf("%v",err)
    return strings.Split(s,"\n")[0]
  }
  return ""
}

// If s contains e
func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}


func checkFile(filename string) bool {
  fi, err := os.Stat(filename)
  return !os.IsNotExist(err) && fi.Size()!=0
}
