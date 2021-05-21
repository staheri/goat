// Implements Experiments and their methods
package evaluate

import(
  "bufio"
  "github.com/staheri/goatlib/instrument"
  "fmt"
  "os"
  "strconv"
  "path/filepath"
  _"time"
  "os/exec"

)


type ECTExperiment struct{
  Experiment
  Args         []string        `json:"args"`
  ID           string          `json:"ID"`
  GoVer        string          `json:"goVer"`
  TraceDir     string          `json:"traceDir,omitempty"`
}


func (ex *ECTExperiment) Init(race bool){
  // Variables
  var predir string

  ws := os.Getenv("GOATWS")
  if ws == "" {
    panic("GOATWS is not set!")
  }
  switch ex.ID {
  case "ECT_native":
    ex.Instrumentor = builtinDL_inst
    ex.GoVer = GOVER_ORIG
    predir = filepath.Join(ws,ex.Target.BugType,ex.Target.BugName,"native")
  case "ECT_ET":
    ex.Instrumentor = goat_trace_inst
    ex.GoVer = GOVER_ORIG
    predir = filepath.Join(ws,ex.Target.BugType,ex.Target.BugName,"tracing")
    ex.TraceDir = filepath.Join(predir,"traces",ex.ID)
    err := os.MkdirAll(ex.TraceDir,os.ModePerm)
    check(err)
  case "ECT_ECT":
    ex.Instrumentor = goat_trace_inst
    ex.GoVer = GOVER_GOAT
    predir = filepath.Join(ws,ex.Target.BugType,ex.Target.BugName,"tracing")
    ex.TraceDir = filepath.Join(predir,"traces",ex.ID)
    err := os.MkdirAll(ex.TraceDir,os.ModePerm)
    check(err)

  }

  fmt.Printf("%s: Init...\n",ex.ID)


  err := os.MkdirAll(predir,os.ModePerm)
  check(err)
  ex.PrefixDir = predir

  // also make predir/src predir/bin predir/out predir/aux predir/trace
  err = os.MkdirAll(filepath.Join(predir,"src"),os.ModePerm)
  check(err)
  err = os.MkdirAll(filepath.Join(predir,"bin"),os.ModePerm)
  check(err)
  err = os.MkdirAll(filepath.Join(predir,"out"),os.ModePerm)
  check(err)
  ex.OutPath = predir+"/out/"+ex.ID+".out"
  f,err := os.Create(ex.OutPath)
  check(err)
  ex.OutBuf = bufio.NewWriter(f)
  ex.Timeout = TO
  ex.Cpu = CPU

  // Set environment variables for GOAT experiments
  _b := strconv.Itoa(int(ex.Timeout))
  os.Setenv("GOATTO",_b)
  fmt.Println("set GOATTO",_b)

}

// Instrument Goat-experiment
func (ex *ECTExperiment) Instrument() {
  fmt.Printf("%s: Instrument...\n",ex.ID)
  files, err := filepath.Glob(ex.PrefixDir+"/src/*.go")
  check(err)
  if len(files) > 0{
    // this experiment has already been instrumented, so no need
    return
  }
  destination := filepath.Join(ex.PrefixDir,"src")
  if ex.Instrumentor(ex.Target.BugDir,destination) != nil{
    panic("Error instrumenting ToolExperiment")
  }
}

// Build ECT-experiment
func (ex *ECTExperiment) Build(race bool) {
  fmt.Printf("%s: Build...\n",ex.ID)

  // change link of GO
  cmd := exec.Command("ln","-nsf",ex.GoVer,os.Getenv("GOROOT"))
  err := cmd.Run()
  check(err)

  src := filepath.Join(ex.PrefixDir,"src")
  dest := filepath.Join(ex.PrefixDir,"bin")
  files, err := filepath.Glob(dest+"/*"+ex.ID)
  check(err)
  if len(files) != 0{ // check if binary exist
    ex.BinaryName = filepath.Base(files[0]) // assign the first found binary to current gex binaryPath
    return
  }
  ex.BinaryName = instrument.BuildCommand(src,dest,ex.ID,"main",false) // race=false
}
