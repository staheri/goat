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
)


// Execute and analyze Goat-experiment
func (gex *GoatExperiment) Execute(i int, race bool) *Result {
  if race{ // Race detection
    // Variables
    var out []byte
    var err error
    var command string
    var commandVals []interface{}

    // placeholder for results
    result := &Result{}

    // Set environment variables for GOAT experiments
    _b := strconv.Itoa(int(gex.Bound))
    os.Setenv("GOATRSBOUND",_b)
    fmt.Println("set GOATRSBOUND",_b)

    ///////////////////////////////////////////////////
    // Execute the (instrumented & built) application
    // Measure time
    ///////////////////////////////////////////////////
    // Prepare commands
    command = "%v -test.failfast"
    commandVals = []interface{}{filepath.Join(gex.PrefixDir,"bin",gex.BinaryName)}
    if gex.Cpu != 0 {
      command += " -test.cpu %v"
      commandVals = append(commandVals, gex.Cpu)
    }

    // Format the command
    args := strings.Split(fmt.Sprintf(command, commandVals...), " ")

    // Execute and Measure time
    result.Time = MeasureTime(func() {
      out,err = exec.Command(args[0], args[1:]...).CombinedOutput()
      fmt.Println("OUT",string(out))
      if  err != nil {
    		log.Println("modified program failed:", err," OUT ", string(out))
    	}
    })
    _,err = fmt.Fprintf(gex.OutBuf,"-----\nRun # %d\n----\n%v\n",i,string(out))
    check(err)
    gex.OutBuf.Flush()

    result.Detected,result.Desc = gex.Detector(out)
    return result

  }else{ // Deadlock detection
    // placeholder for results
    result := &Result{}

    // Set environment variables for GOAT experiments
    _b := strconv.Itoa(int(gex.Bound))
    b := gex.Bound
    os.Setenv("GOATRSBOUND",_b)
    if b < 0{
      os.Setenv("GOATMAXPROCS",MAXPROCS)
      _b = "T"
    }else{
      os.Setenv("GOATMAXPROCS","1")
    }

    // FileName name to store events
    traceName := fmt.Sprintf("%s_B%v_I%d",gex.Target.BugName,_b,i)

    ///////////////////////////////////////////////////
    // Execute the (instrumented & built) application
    // Measure time
    ///////////////////////////////////////////////////
    execRes,err := instrument.ExecuteTrace(filepath.Join(gex.PrefixDir,"bin",gex.BinaryName))

    result.Time = execRes.ExecTime

    // Handle runtime errors & empty trace
    if err != nil{
      if execRes != nil{
        // CRASH - Runtime error
        _,err1 := fmt.Fprintf(gex.OutBuf,"-----\nRun # %d\n----\n%v\n%v\n",i,"Runtime error: ",execRes.TraceBuffer.String())
        check(err1)
        gex.OutBuf.Flush()
        result.Detected = true
        result.Desc = "CRASH"
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
    result.TracePath = filepath.Join(gex.TraceDir,traceName)+".trace"
    traceBytes_n := traceops.WriteTrace(execRes.TraceBuffer.Bytes(),result.TracePath)
    result.TraceSize = traceBytes_n
    fmt.Printf("\tTrace File: %s \n\tSize: %d bytes\n",traceName,traceBytes_n)
    // Parse trace
    parseRes, err := trace.ParseTrace(execRes.TraceBuffer, filepath.Join(gex.PrefixDir,"bin",gex.BinaryName))
    check(err)
    fmt.Printf("\t# Events: %d\n",len(parseRes.Events))

    // Check length of events
    result.EventsLen = len(parseRes.Events)
    if len(parseRes.Events) > EVENT_BOUND {
      result.Detected = false
      result.Desc = "ABORT"
      return result
    }

    // Check for deadlocks
    deadlock_report := traceops.DeadlockChecker(parseRes,false) // longReport = false

    // Finalize result and return
    result.TotalG = deadlock_report.TotalG
    if deadlock_report.GlobalDL {
      _,err = fmt.Fprintf(gex.OutBuf,"-----\nRun # %d\n----\n%v\n%v\n",i,"GOAT: Global Deadlock",deadlock_report.Message)
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
      _,err = fmt.Fprintf(gex.OutBuf,"-----\nRun # %d\n----\n%v\n%v\n",i,"GOAT: Partial Deadlock, Leaked Goroutines:"+strconv.Itoa(deadlock_report.Leaked),deadlock_report.Message)
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
    _,err = fmt.Fprintf(gex.OutBuf,"-----\nRun # %d\n----\n%v\n",i,"PASS")
    check(err)

    gex.OutBuf.Flush()
    gex.LastSuccessTrace = result.TracePath

    result.Detected = false
    result.Desc = "PASS"
    return result
  }
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

  if tex.ToolID == "lockDL"{
    command = "%v -test.failfast -test.timeout %v"
    commandVals = []interface{}{filepath.Join(tex.PrefixDir,"bin",tex.BinaryName), time.Duration(tex.Timeout)*time.Second}
  } else{
    command = "%v -test.failfast"
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
  if ex.ID == "ECT_native"{
    // Variables
    var stderr,stdout  bytes.Buffer

    // placeholder for results
    result := &Result{}

    cmd := exec.Command(filepath.Join(ex.PrefixDir,"bin",ex.BinaryName),ex.Args...)
    cmd.Stderr = &stderr
    cmd.Stdout = &stdout
    time := MeasureTime(func() {
      if err := cmd.Run(); err != nil {
        fmt.Printf("program execution failed\nErr: %v\nStderr: %v\nStdout: %v\n", err, stderr.String(),stdout.String())
      }
    })
    result.Time = time
    return result

  }else{ // execute and trace
    // Variables
    //var trace     []byte
    //var parseRes  *trace.ParseResult

    // placeholder for results
    result := &Result{}

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
    result.TotalG,result.TotalCh = countGCH(parseRes.Events)
    result.StackSize = len(parseRes.Stacks)

    return result
  }
}
