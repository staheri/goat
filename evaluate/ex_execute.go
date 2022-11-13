// Implementation of execution for GOAT and other tools

package evaluate

import(
  "fmt"
  "os"
  "strconv"
  "os/exec"
  "strings"
  "github.com/staheri/goatlib/instrument"
  "github.com/staheri/goatlib/traceops"
  "github.com/staheri/goatlib/trace"
  "log"
  "bytes"
  "path/filepath"
  "time"
  _"io"
  "encoding/json"
  "io/ioutil"
)


// Execute and analyze Goat-experiment
func (gex *GoatExperiment) Execute(i int, race bool) *Result {
  // Deadlock detection
  // placeholder for results
    result := &Result{}
    var parseRes *trace.ParseResult

    // Set environment variables for GOAT experiments
    _b := strconv.Itoa(int(gex.Bound))
    os.Setenv("GOATRSBOUND",_b)

    // FileName name to store events
    traceName := fmt.Sprintf("%s_B%v_I%d",gex.Target.BugName,_b,i)
    result.TracePath = filepath.Join(gex.TraceDir,traceName)+".trace"
    ///////////////////////////////////////////////////
    // check if trace exist first
    // if not, Execute the (instrumented & built) application
    // Measure time
    ///////////////////////////////////////////////////
    if checkFile(result.TracePath){ // if trace exist
      // read trace
      // obtain trace
    	trc,size,err := traceops.ReadTrace(result.TracePath)
    	check(err)
      result.TraceSize = size
      parseRes,err = trace.ParseTrace(trc,filepath.Join(gex.PrefixDir,"bin",gex.BinaryName))
    	check(err)
      // read trace time
      result.Time,_ = time.ParseDuration(traceops.ReadTime(filepath.Join(gex.PrefixDir,"traceTimes",traceName)+".time"))

      if race { // results are already out there
        return result
      }
    } else{ // trace does not exist, now execute trace

      execRes,err := instrument.ExecuteTrace(filepath.Join(gex.PrefixDir,"bin",gex.BinaryName))
      //fmt.Printf("\tAfter ExecuteTrace:\n\t\trace: %v\n\t\tlen(trace): %v\n",execRes.RaceMessage,execRes.TraceBuffer.Len(),)
      result.Time = execRes.ExecTime
      // write trace time
      traceops.WriteTime(fmt.Sprintf("%v",result.Time),filepath.Join(gex.PrefixDir,"traceTimes",traceName)+".time")
      // Handle runtime errors & empty trace
      if err != nil{
        if execRes != nil{
          // CRASH - Runtime error
          //fmt.Printf("\tProbably Crash (before):\n\t\trace: %v\n\t\tlen(trace): %v\n",execRes.RaceMessage,execRes.TraceBuffer.Len())
          _,err1 := fmt.Fprintf(gex.OutBuf,"-----\nRun # %d\n----\n%v\n%v\n",i,"Runtime error: ",execRes.TraceBuffer.String())
          check(err1)
          gex.OutBuf.Flush()
          if race {
            fmt.Printf("\tProbably Crash (after):\n\t\trace: %v\n\t\tlen(trace): %v\n",execRes.RaceMessage,execRes.TraceBuffer.Len())
            result.Detected,result.Desc = gex.Detector(append([]byte(execRes.RaceMessage),execRes.TraceBuffer.Bytes()...))
          } else{
            result.Detected = true
            result.Desc = "CRASH"
          }
          return result
        } else{
          // Empty Trace
          _,err1 := fmt.Fprintf(gex.OutBuf,"-----\nRun # %d\n----\n%v\n",i,"Empty Trace")
          check(err1)
          gex.OutBuf.Flush()
          result.Detected = false
          result.Desc = "NONE"
          return result
        }
      }
      ////////////////////////
      // Process Traces
      ///////////////////////
      // Write to file
      if race{
        // trace is not there so results are not there either
        bytesOfTrace := execRes.TraceBuffer.Bytes()
        fmt.Printf("\tAfter Parse:\n\t\tlen(trace): %v\n\t\tlen(trace_back): %v\n",execRes.TraceBuffer.Len(),len(bytesOfTrace))
        parseRes, err = trace.ParseTrace(execRes.TraceBuffer, filepath.Join(gex.PrefixDir,"bin",gex.BinaryName))
        if err!= nil{
          // problem with parsing the trace, re-runt
          // there is a runtime error (panic) w/ or w/o race within the trace buffer
          fmt.Println("ERROR in parsing trace")
          result.Detected,result.Desc = gex.Detector(append([]byte(execRes.RaceMessage),bytesOfTrace...))
          return result
        }
        check(err)
        //fmt.Printf("\tAfter Parse:\n\t\tlen(trace): %v\n\t\tlen(trace_back): %v\n",execRes.TraceBuffer.Len(),len(bytesOfTrace))
        result.LStack = gex.UpdateGStack(parseRes.Stacks)
        gex.UpdateConcUsage(parseRes.Stacks,result.LStack)
        gex.UpdateGGTree(parseRes,result.LStack)
        gex.UpdateCoverageGGTree(parseRes,result.LStack)
        gex.UpdateCoverageReport()
        result.Coverage1 = gex.PrintCoverageReport(true)
        result.Coverage2 = gex.PrintCoverageReport(false)

        // analysis trace
        // measure coverage
        // detect result
        // return result
        //if execRes.RaceMessage != ""{
          //fmt.Println("Out race:",execRes.RaceMessage)
          //fmt.Println("Len trace_back:",len(bytesOfTrace))
          //result.Detected,result.Desc = gex.Detector(append([]byte(execRes.RaceMessage),bytesOfTrace...))
        //} else{
          //fmt.Println("Len trace_back:",len(bytesOfTrace))
        //result.Detected,result.Desc = gex.Detector(bytesOfTrace)
        result.Detected,result.Desc = gex.Detector(append([]byte(execRes.RaceMessage),bytesOfTrace...))
        //}

        //fmt.Println("Len trace_back:",len(bytesOfTrace))
        traceBytes_n := traceops.WriteTrace(bytesOfTrace,result.TracePath)
        //traceops.TraceToJSON(parseRes,filepath.Join(gex.TraceDir,traceName)+".json")
        result.TraceSize = traceBytes_n
        fmt.Printf("\tWrite to Trace File: %s \n\tSize: %d bytes\n",traceName,traceBytes_n)
        return result
      }

      traceBytes_n := traceops.WriteTrace(execRes.TraceBuffer.Bytes(),result.TracePath)
      result.TraceSize = traceBytes_n
      fmt.Printf("\tWrite to Trace File: %s \n\tSize: %d bytes\n",traceName,traceBytes_n)
      // Parse trace
      parseRes, err = trace.ParseTrace(execRes.TraceBuffer, filepath.Join(gex.PrefixDir,"bin",gex.BinaryName))
      check(err)
      //traceops.TraceToJSON(parseRes,filepath.Join(gex.TraceDir,traceName)+".json")


    } // parseRes is either obtained from execution or pre execution


    file, _ := json.MarshalIndent(parseRes, "", " ")
    _ = ioutil.WriteFile(filepath.Join(gex.TraceDir,traceName)+"-json.trace", file, 0644)

    fmt.Printf("\t# Events: %d\n",len(parseRes.Events))
    // print events
    for _,e := range(parseRes.Events){
      fmt.Println(e)
      fmt.Println("-----------------------------------------------------------")
    }
    // Check length of events
    result.EventsLen = len(parseRes.Events)
    /*if len(parseRes.Events) > EVENT_BOUND && !race{
      result.Detected = false
      result.Desc = "ABORT"
      return result
    }*/

    // Check for deadlocks
    deadlock_report := traceops.DeadlockChecker(parseRes,false) // longReport = false

    // // print events
    // for i,e := range(parseRes.Events){
    //   // ignore GOAT events
    //   if traceops.IsGoatFunction(e.Stk){
    //     continue
    //   }
    //   fmt.Printf("****\nG%v (idx:%v)\n%v\n",e.G,i,e.String())
    // }

    // get the local stack
    result.LStack = gex.UpdateGStack(parseRes.Stacks)
    gex.UpdateConcUsage(parseRes.Stacks,result.LStack)
    gex.UpdateGGTree(parseRes,result.LStack)
    gex.UpdateCoverageGGTree(parseRes,result.LStack)
    gex.UpdateCoverageReport()
    result.Coverage1 = gex.PrintCoverageReport(true)
    result.Coverage2 = gex.PrintCoverageReport(false)

    // Finalize result and return
    result.TotalG = deadlock_report.TotalG

    if deadlock_report.GlobalDL {
      _,err := fmt.Fprintf(gex.OutBuf,"-----\nRun # %d\n----\n%v\n%v\n",i,"GOAT: Global Deadlock",deadlock_report.Message)
      check(err)

      gex.OutBuf.Flush()
      gex.LastFailedTrace = result.TracePath
      if gex.FirstFailedAfter == 0{
        gex.FirstFailedAfter = i
      }

      result.Detected = true
      result.Desc = "GDL"
      return result

    } else if deadlock_report.Leaked != 0{
      _,err := fmt.Fprintf(gex.OutBuf,"-----\nRun # %d\n----\n%v\n%v\n",i,"GOAT: Partial Deadlock, Leaked Goroutines:"+strconv.Itoa(deadlock_report.Leaked),deadlock_report.Message)
      check(err)

      gex.OutBuf.Flush()
      gex.LastFailedTrace = result.TracePath
      if gex.FirstFailedAfter == 0{
        gex.FirstFailedAfter = i
      }
      result.Detected = true
      result.Desc = "PDL-"+strconv.Itoa(deadlock_report.Leaked)
      return result

    }
     // No Deadlock detected -> PASS
    _,err := fmt.Fprintf(gex.OutBuf,"-----\nRun # %d\n----\n%v\n",i,"PASS")
    check(err)

    gex.OutBuf.Flush()
    gex.LastSuccessTrace = result.TracePath

    result.Detected = false
    result.Desc = "PASS"
    return result
}

// Execute and analyze Tool-experiment
func (tex *ToolExperiment) Execute(i int,race bool) *Result {
  // Variables
  var out []byte
  var err error
  var command string
  var commandVals []interface{}

  // placeholder for results
  result := &Result{}

  //if tex.ToolID == "lockDL" || tex.ToolID == "goleak"{
  if tex.ToolID == "lockDL"{
    command = "%v -test.failfast -test.timeout %v"
    commandVals = []interface{}{filepath.Join(tex.PrefixDir,"bin",tex.BinaryName), time.Duration(tex.Timeout)*time.Second}
  } else{
    command = "%v -test.failfast "
    commandVals = []interface{}{filepath.Join(tex.PrefixDir,"bin",tex.BinaryName)}
  }
  //
  if tex.Cpu != 0 {
    command += " -test.cpu %v"
    commandVals = append(commandVals, tex.Cpu)
  }

  // format the command
  args := strings.Split(fmt.Sprintf(command, commandVals...), " ")

  // execute and measure time
  result.Time = MeasureTime(func() {
    out,err = exec.Command(args[0], args[1:]...).CombinedOutput()
    if  err != nil {
  		log.Println("modified program failed:", err," OUT ", string(out))
  	}
  })
  _,err = fmt.Fprintf(tex.OutBuf,"-----\nRun # %d\n----\n%v\n",i,string(out))
  check(err)
  tex.OutBuf.Flush()

  result.Detected,result.Desc = tex.Detector(out)
  return result
}

// Execute and analyze ECT-experiment
func (ex *ECTExperiment) Execute(i int, race bool) *Result {
  // set timeout
  old_TO := os.Getenv("GOATTO")
  os.Setenv("GOATTO","120")

  // placeholder for results
  result := &Result{}

  if ex.ID == "ECT_native"{
    // Variables
    var stderr,stdout  bytes.Buffer

    cmd := exec.Command(filepath.Join(ex.PrefixDir,"bin",ex.BinaryName),ex.Args...)
    cmd.Stderr = &stderr
    cmd.Stdout = &stdout
    time := MeasureTime(func() {
      if err := cmd.Run(); err != nil {
        fmt.Printf("program execution failed\nErr: %v\nStderr: %v\nStdout: %v\n", err, stderr.String(),stdout.String())
      }
    })
    result.Time = time
  }else{ // execute and trace
    // FileName name to store events
    traceName := fmt.Sprintf("%s_%v_%v_I%d",ex.Target.BugName,strings.Join(ex.Args,"_"),ex.ID,i)

    // Execute the (instrumented/built) application
    // Measure time
    execRes,err := instrument.ExecuteTrace(filepath.Join(ex.PrefixDir,"bin",ex.BinaryName),ex.Args...)
    check(err)

    result.TracePath = filepath.Join(ex.TraceDir,traceName)+".trace"
    traceBytes_n := traceops.WriteTrace(execRes.TraceBuffer.Bytes(),result.TracePath)
    result.TraceSize = traceBytes_n
    fmt.Printf("\tTrace File: %s \n\tSize: %d bytes\n",traceName,traceBytes_n)

    parseRes, err := trace.ParseTrace(execRes.TraceBuffer, filepath.Join(ex.PrefixDir,"bin",ex.BinaryName))
    check(err)

    // parseRes holds events and stacktraces of trace
    fmt.Printf("\t# Events: %d\n",len(parseRes.Events))

    result.EventsLen = len(parseRes.Events)
    result.Time = execRes.ExecTime
    cnt := count(parseRes.Events)
    result.TotalG = cnt.G
    result.TotalCh = cnt.Ch
    result.StackSize = len(parseRes.Stacks)
  }
  // set back timeout
  os.Setenv("GOATTO",old_TO)

  return result
}
