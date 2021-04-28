// Implements Experiments and their methods
package evaluate

import(
  "bufio"
  "github.com/staheri/goatlib/instrument"
  "fmt"
  "os"
  "strconv"
  "path/filepath"
  "time"

)

const RESDIR = "/Volumes/DATA/goatws/results"
const TO = 2 // second
const CPU = 0
const MAXPROCS = "4"
const EVENT_BOUND = 20000000
const WORKDIR = ""
const ORIGINAL_GO = "go.1.15.6"
const NEW_GO = "myGo.1.15.6"

type InstFunc func(string,string) string
type DetectFunc func([]byte) (bool,string)

// Interface for Experiments
type Ex interface{
  Init(bool)                 // init the experiments (bool: race)
  Instrument()               // instrument the program
  Build(bool)                // build the program (bool: race)
  Execute(int,bool)    *Result    // execute the instrumented program (int: #iteration)
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
  TotalG        int                `json:"totalg,omitempty"` // for goat
  TotalCh       int                `json:"totalch,omitempty"` // for goat
  StackSize     int                `json:"stackSize,omitempty"` // for goat
  EventsLen     int                `json:"eventsLen,omitempty"` // for goat
  Detected      bool                   `json:"detected"`
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

  ws := os.Getenv("GOATWS")
  if ws == "" {
    panic("GOATWS is not set!")
  }
  switch gex.Bound {
  case -1:
    if race{
      panic("Goat_m and Goat_n are not compatible with race")
    }
    gex.ID = "goat_m"
    gex.Instrumentor = goat_trace_inst
  case 0:
    if race{
      panic("Goat_m and Goat_n are not compatible with race")
    }
    gex.ID = "goat_u"
    gex.Instrumentor = goat_trace_inst

  default:
    if race{
      gex.ID = "goat_race_d"+strconv.Itoa(int(gex.Bound))
      gex.Instrumentor = goat_critic_inst
    }else{
      gex.ID = "goat_d"+strconv.Itoa(int(gex.Bound))
      gex.Instrumentor = goat_delay_inst
    }
  }

  fmt.Printf("%s: Init...\n",gex.ID)
  if race{
    predir = filepath.Join(ws,gex.Target.BugType,gex.Target.BugName,"goat_race")
  }else{
    predir = filepath.Join(ws,gex.Target.BugType,gex.Target.BugName,"goat_"+gex.GetMode())
  }

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

  // Set environment variables for GOAT experiments
  _b := strconv.Itoa(int(gex.Timeout))
  os.Setenv("GOATTO",_b)
  fmt.Println("set GOATTO",_b)

}

// Instrument Goat-experiment
func (gex *GoatExperiment) Instrument() {
  fmt.Printf("%s: Instrument...\n",gex.ID)
  files, err := filepath.Glob(gex.PrefixDir+"/src/*.go")
  check(err)
  if len(files) > 0{
    // this experiment has already been instrumented, so no need
    return
  }
  destination := filepath.Join(gex.PrefixDir,"src")
  critic := gex.Instrumentor(gex.Target.BugDir,destination)
  if gex.Bound < 1 && critic != ""{
    panic("mismatch bound & instrument")
  }
  // write critic to a file
  if critic != ""{
    f,err := os.Create(gex.PrefixDir+"/criticalPoints.json")
    check(err)
    _,err = f.WriteString(critic)
    f.Close()
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
  predir := filepath.Join(ws,tex.Target.BugType,tex.Target.BugName,tex.ToolID)
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

  tex.OutPath = predir+"/out/"+tex.ToolID+".out"
  f,err := os.Create(tex.OutPath)
  check(err)
  tex.OutBuf = bufio.NewWriter(f)
  tex.Timeout = TO
  tex.Cpu = CPU
}

// Instrument ToolExperiment
func (tex *ToolExperiment) Instrument()  {
  fmt.Printf("%s: Instrument...\n",tex.ToolID)
  destination := filepath.Join(tex.PrefixDir,"src")
  if tex.Instrumentor(tex.Target.BugDir,destination) != ""{
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
