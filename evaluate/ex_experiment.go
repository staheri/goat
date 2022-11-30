// Implements Experiments and their methods
package evaluate

import(
  "bufio"
  "github.com/staheri/goatlib/instrument"
  "github.com/staheri/goatlib/trace"
  "fmt"
  "os"
  "strconv"
  "path/filepath"
  "time"
  "encoding/json"

)

const RESDIR = "/home/saeed/goatws/results"
const TO = 30 // second
const CPU = 0
const EVENT_BOUND = 200000000
const WORKDIR = ""
const GOVER_ORIG = "/home/saeed/go-builds/go-orig-1.15.6"
const GOVER_GOAT = "/home/saeed/go-builds/go-goat-1.15.6"
const TERMINATION = "hitBug"
// const TERMINATION = "ignoreGDL"
// const TERMINATION = "thresh"
type InstFunc func(string,string) []*instrument.ConcurrencyUsage
type DetectFunc func([]byte) (bool,string)

// Interface for Experiments
type Ex interface{
  Init(bool)                 // init the experiments (bool: race)
  Instrument()               // instrument the program
  Build(bool)                // build the program (bool: race)
  Execute(int,bool, bool)  *Result    // execute the instrumented program (int: #iteration)
}

// Struct for Bugs
type Bug struct{
  BugName       string                 `json:"bugName"` // Name
  BugDir        string                 `json:"bugDir"` // Dir
  BugType       string                 `json:"bugType"` // Blocking - Nonblocking - overhead
  BugCause      string                 `json:"bugCause"` // Majore Cause
  BugSubCause   string                 `json:"bugSubcause"` // Sub-cause
}

// Basic config for experiments
type ExConfig struct{
  Timeout        int               `json:"timeout"` // time.Second
  Cpu            int               `json:"cpu"`
}

// Struct for all experiments
type Experiment struct{
  ExConfig
  Target         *Bug                  `json:"-"`
  PrefixDir      string                `json:"prefixDir"`
  BinaryName     string                `json:"binaryName"`
  OutPath        string                `json:"outPath"`
  OutBuf         *bufio.Writer         `json:"-"`
  Results        []*Result             `json:"results"`
  Detector       DetectFunc            `json:"-"`
  Instrumentor   InstFunc              `json:"-"`
}

// Struct for Goat experiments
type GoatExperiment struct{
  Experiment
  ID                  string                `json:"goatid"`
  Bound               int                   `json:"goatBound"` // -1: multi, 0: uni,  1 =< : delay
  TraceDir            string                `json:"traceDir"`// for blocking bugs only
  LastFailedTrace     string                `json:"lastFailedTrace"`
  LastSuccessTrace    string                `json:"lastSuccessTrace"`
  FirstFailedAfter    int                   `json:"firstFailedAfter"`
  GGTree              *GGTree               `json:"-"` // (will be reconstructed from replay)
  TotalGG             int                   `json:"-"` // (will be reconstructed from replay)
  ConcUsage           *ConcUsageStruct      `json:"-"` // (will be reconstructed from replay)
  GStack              *GlobalStack          `json:"-"` // (will be reconstructed from replay)
}

// Struct for Tool experiments
type ToolExperiment struct{
  Experiment
  ToolID             string            `json:"toolid"`
}

// Hold results from experimnets
type Result struct{
  Time          time.Duration          `json:"time"`
  Desc          string                 `json:"desc,omitempty"`
  TracePath     string                 `json:"tracePath,omitempty"` // for goat
  TraceSize     int                    `json:"traceSize,omitempty"` // for goat
  TotalG        int                    `json:"totalg,omitempty"` // for goat
  TotalCh       int                    `json:"totalch,omitempty"` // for goat
  StackSize     int                    `json:"stackSize,omitempty"` // for goat
  EventsLen     int                    `json:"eventsLen,omitempty"` // for goat
  Detected      bool                   `json:"detected"` // for goat
  LStack        map[uint64]string      `json:"lstack,omitempty"` // for goat (will be reconstructed from replay)
  Coverage1     float64                `json:"coverage1,omitempty"` // for goat (will be reconstructed from replay)
  Coverage2     float64                `json:"coverage1,omitempty"` // for goat (will be reconstructed from replay)
}

///////////////////////////////////////////////////////
// GOAT Methods that implements Ex interface
///////////////////////////////////////////////////////

// GetMode
func (gex *GoatExperiment) GetMode() string{
  if gex.Bound < 1{
    return "trace"
  }
  return "delay"
}

// Init Goat
func (gex *GoatExperiment) Init(race bool) {
  // Variables
  var predir string
  var preToolName string

  ws := os.Getenv("GOATWS")
  if ws == "" {
    panic("GOATWS is not set!")
  }

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

  if gex.Bound < 0 {
    panic("invalid bound")
  }

  if race{
    preToolName = "goat_race"
  } else{
    preToolName = "goat"
  }

  gex.ID = preToolName+"_d"+strconv.Itoa(int(gex.Bound))

  if gex.Bound == 0{
    gex.Instrumentor = goat_trace_inst
  } else{
    gex.Instrumentor = goat_delay_inst
  }

  fmt.Printf("%s: Init...\n",gex.ID)
  predir = filepath.Join(ws,"p"+MAXPROCS,gex.Target.BugType+"_"+gex.Target.BugName,preToolName+"_"+gex.GetMode())

  err := os.MkdirAll(predir,os.ModePerm)
  check(err)
  gex.PrefixDir = predir

  // also make predir/src predir/bin predir/out predir/aux predir/trace
  err = os.MkdirAll(filepath.Join(predir,"src"),os.ModePerm)
  check(err)
  err = os.MkdirAll(filepath.Join(predir,"bin"),os.ModePerm)
  check(err)
  err = os.MkdirAll(filepath.Join(predir,"out"),os.ModePerm)
  check(err)
  err = os.MkdirAll(filepath.Join(predir,"results"),os.ModePerm)
  check(err)
  err = os.MkdirAll(filepath.Join(predir,"visual"),os.ModePerm)
  check(err)
  err = os.MkdirAll(filepath.Join(predir,"traceTimes"),os.ModePerm)
  check(err)
  gex.TraceDir = filepath.Join(predir,"traces",gex.ID)
  err = os.MkdirAll(gex.TraceDir,os.ModePerm)
  check(err)
  gex.OutPath = predir+"/out/goat_"+gex.ID+".out"
  f,err := os.Create(gex.OutPath)
  check(err)
  gex.OutBuf = bufio.NewWriter(f)
  gex.Timeout = TO
  gex.Cpu = CPU
  if race{
    gex.Detector = race_detector
  }

  // setup global stack
  fmap := make(map[int]*trace.Frame)
  fsmap := make(map[string]int)
  gstack := &GlobalStack{fmap,fsmap}
  gex.GStack = gstack


  // Set environment variables for GOAT experiments
  _b := strconv.Itoa(int(gex.Timeout))
  os.Setenv("GOATTO",_b)
  fmt.Println("set GOATTO",_b)

  os.Setenv("GOATMAXPROCS",MAXPROCS)
  fmt.Println("set GOATMAXPROCS",MAXPROCS)

}

// Instrument Goat-experiment
func (gex *GoatExperiment) Instrument() {
  fmt.Printf("%s: Instrument...\n",gex.ID)
  files, err := filepath.Glob(gex.PrefixDir+"/src/*.go")
  check(err)
  if len(files) > 0{
    // this experiment has already been instrumented, so no need
    // read from file

    concUsage := ReadConcUsage(gex.PrefixDir+"/concUsage.json")
    if concUsage != nil{
      gex.ConcUsage = &ConcUsageStruct{ConcUsage:concUsage}
      gex.InitConcMap()
    } else{
      panic("no concusage")
    }
    return
  }
  destination := filepath.Join(gex.PrefixDir,"src")
  concUsage := gex.Instrumentor(gex.Target.BugDir,destination)
  // write critic to a file
  // instead store coverage in the GoatExperiment
  if concUsage != nil{
    gex.ConcUsage = &ConcUsageStruct{ConcUsage:concUsage}
    gex.InitConcMap()
    // write to json file
    rep,err := os.Create(gex.PrefixDir+"/concUsage.json")
    check(err)
    newdat ,err := json.MarshalIndent(concUsage,"","    ")
    check(err)
    _,err = rep.WriteString(string(newdat))
    check(err)
    rep.Close()
  }
}

// Build Goat-experiment
func (gex *GoatExperiment) Build(race bool) {
  fmt.Printf("%s: Build...\n",gex.ID)
  src := filepath.Join(gex.PrefixDir,"src")
  dest := filepath.Join(gex.PrefixDir,"bin")
  files, err := filepath.Glob(dest+"/*"+gex.GetMode())
  check(err)
  if len(files) != 0{ // check if binary exist
    gex.BinaryName = filepath.Base(files[0]) // assign the first found binary to current gex binaryPath
    return
  }
  gex.BinaryName = instrument.BuildCommand(src,dest,gex.GetMode(),"test",race) // race=false
}

///////////////////////////////////////////////////////
// Tool Methods that implements Ex interface
///////////////////////////////////////////////////////

// Init ToolExperiment
func (tex *ToolExperiment) Init(race bool) {
  fmt.Printf("%s: Init...\n",tex.ToolID)
  ws := os.Getenv("GOATWS")
  if ws == "" {
    panic("GOATWS is not set!")
  }

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

  tex.Detector = race_detector
  switch tex.ToolID{
  case "lockDL":
    tex.Instrumentor = lockDL_inst
    if !race{
      tex.Detector = lockDL_detector
    }
  case "goleak":
    tex.Instrumentor = goleak_inst
    if !race{
      tex.Detector = goleak_detector
    }
  default:
    tex.Instrumentor = builtinDL_inst
    if !race{
      tex.Detector = builtinDL_detector
    }
  }
  predir := filepath.Join(ws,"p"+MAXPROCS,tex.Target.BugType+"_"+tex.Target.BugName,tex.ToolID)
  err := os.MkdirAll(predir,os.ModePerm)
  check(err)
  tex.PrefixDir = predir
  // also make predir/src predir/bin predir/out predir/aux predir/trace
  err = os.MkdirAll(filepath.Join(predir,"src"),os.ModePerm)
  check(err)
  err = os.MkdirAll(filepath.Join(predir,"bin"),os.ModePerm)
  check(err)
  err = os.MkdirAll(filepath.Join(predir,"out"),os.ModePerm)
  check(err)
  err = os.MkdirAll(filepath.Join(predir,"results"),os.ModePerm)
  check(err)

  tex.OutPath = predir+"/out/"+tex.ToolID+".out"
  f,err := os.Create(tex.OutPath)
  check(err)
  tex.OutBuf = bufio.NewWriter(f)
  tex.Timeout = TO
  tex.Cpu = CPU

  os.Setenv("GOATMAXPROCS",MAXPROCS)
  fmt.Println("set GOATMAXPROCS",MAXPROCS)
}

// Instrument ToolExperiment
func (tex *ToolExperiment) Instrument()  {
  fmt.Printf("%s: Instrument...\n",tex.ToolID)
  destination := filepath.Join(tex.PrefixDir,"src")
  if tex.Instrumentor(tex.Target.BugDir,destination) != nil{
    panic("Error instrumenting ToolExperiment")
  }
}

// Build ToolExperiment
func (tex *ToolExperiment) Build(race bool)  {
  fmt.Printf("%s: Build...\n",tex.ToolID)
  src := filepath.Join(tex.PrefixDir,"src")
  dest := filepath.Join(tex.PrefixDir,"bin")
  tex.BinaryName = instrument.BuildCommand(src,dest,tex.ToolID,"test",race)
}
