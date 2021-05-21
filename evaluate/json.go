// JSON-related functions
package evaluate

import(
  "fmt"
  "os"
  "path/filepath"
  "encoding/json"
  "github.com/staheri/goatlib/instrument"
  "strings"
  "io/ioutil"
  "time"
)
// Read GoKer config JSON file
func ReadGoKerConfig(bugtype string) (map[string][]string) {
  ret := make(map[string][]string)
  fname := filepath.Join(os.Getenv("HOME"),"gobench/gobench/configures/goker/"+bugtype+".json")
  fmt.Println("Reading ",fname)
  var dat map[string]interface{}
  if !checkFile(fname){
    panic("Error reading JSON file: "+fname)
  }
  bf ,_ := ioutil.ReadFile(fname)
  err := json.Unmarshal([]byte(bf),&dat)
  check(err)
  for k,vv := range(dat) {
    v := vv.(map[string]interface{})
    ret[k] = []string{v["type"].(string),v["subtype"].(string)}
  }
  return ret
}


func ReadExperimentResults_tool(fname string) (tex *ToolExperiment){
  fmt.Println("Reading ",fname)
  var fields map[string]interface{}
  if !checkFile(fname){
    panic("Error reading JSON file: "+fname)
  }
  bf ,_ := ioutil.ReadFile(fname)
  err := json.Unmarshal([]byte(bf),&fields)
  check(err)

  tex = &ToolExperiment{}
  tex.Timeout = int(fields["timeout"].(float64))
  tex.Cpu = int(fields["cpu"].(float64))
  tex.PrefixDir = fields["prefixDir"].(string)
  tex.BinaryName = fields["binaryName"].(string)
  tex.OutPath = fields["outPath"].(string)

  resultsList := fields["results"].([]interface{})
  results := []*Result{}
  for _,tres := range(resultsList){
    rs := &Result{}
    res := tres.(map[string]interface{})
    rs.Time = time.Duration(res["time"].(float64))*time.Nanosecond
    if _,ok := res["desc"];ok{
      rs.Desc = res["desc"].(string)
    }
    rs.Detected = res["detected"].(bool)
    results = append(results,rs)
  }
  tex.Results = results
  tex.ToolID = fields["toolid"].(string)
  return tex
}

func ReadExperimentResults_goat(fname string) (exp *GoatExperiment){
  fmt.Println("Reading ",fname)
  var fields map[string]interface{}
  if !checkFile(fname){
    panic("Error reading JSON file: "+fname)
  }
  bf ,_ := ioutil.ReadFile(fname)
  err := json.Unmarshal([]byte(bf),&fields)
  check(err)
  gex := &GoatExperiment{}
  gex.Timeout = int(fields["timeout"].(float64))
  gex.Cpu = int(fields["cpu"].(float64))
  gex.PrefixDir = fields["prefixDir"].(string)
  gex.BinaryName = fields["binaryName"].(string)
  gex.OutPath = fields["outPath"].(string)

  resultsList := fields["results"].([]interface{})
  results := []*Result{}
  for _,tres := range(resultsList){
    rs := &Result{}
    res := tres.(map[string]interface{})
    rs.Time = time.Duration(res["time"].(float64))*time.Nanosecond
    if _,ok := res["desc"];ok{
      rs.Desc = res["desc"].(string)
    }
    if tracepath,ok := res["tracePath"];ok{
      rs.TracePath = tracepath.(string)
    }
    if tracesize,ok := res["traceSize"];ok{
      rs.TraceSize = int(tracesize.(float64))
    }
    if stacksize,ok := res["stackSize"];ok{
      rs.StackSize = int(stacksize.(float64))
    }
    if eventslen,ok := res["eventsLen"];ok{
      rs.EventsLen = int(eventslen.(float64))
    }
    if totalG,ok := res["totalg"];ok{
      rs.TotalG = int(totalG.(float64))
    }
    if totalCh,ok := res["totalch"];ok{
      rs.TotalCh =int(totalCh.(float64))
    }
    rs.Detected = res["detected"].(bool)
    results = append(results,rs)
  }
  gex.Results = results
  gex.ID = fields["goatid"].(string)
  gex.Bound = int(fields["goatBound"].(float64))
  gex.TraceDir = fields["traceDir"].(string)
  gex.LastFailedTrace = fields["lastFailedTrace"].(string)
  gex.LastSuccessTrace = fields["lastSuccessTrace"].(string)
  gex.FirstFailedAfter = int(fields["firstFailedAfter"].(float64))
  return gex
}


//
// GGTree              *GGTree               `json:"ggtree"`
// TotalGG             int                   `json:"totalgg"`
// ConcUsage           *ConcUsageStruct      `json:"concusage"`
// GStack              *GlobalStack          `json:"gstack"`


// Read experiment results JSON file
func ReadResults(fname string) map[string]Ex{
  fmt.Println("Reading ",fname)
  var dat map[string]interface{}
  if !checkFile(fname){
    panic("Error reading JSON file: "+fname)
  }
  bf ,_ := ioutil.ReadFile(fname)
  err := json.Unmarshal([]byte(bf),&dat)
  check(err)
  // load from json
  //dat[Bug]
  // we want to read dat[exps]
  if _,ok := dat["exps"]; !ok {
    panic("Result JSON has no field \"exps\"")
  }

  ret := make(map[string]Ex)
  experiments := dat["exps"].(map[string]interface{})
  for k,v := range(experiments){
    // GoatExperiment
    if strings.HasPrefix(k,"goat"){
      gex := &GoatExperiment{}
      fields := v.(map[string]interface{})
      gex.Timeout = int(fields["timeout"].(float64))
      gex.Cpu = int(fields["cpu"].(float64))
      gex.PrefixDir = fields["prefixDir"].(string)
      gex.BinaryName = fields["binaryName"].(string)
      gex.OutPath = fields["outPath"].(string)

      resultsList := fields["results"].([]interface{})
      results := []*Result{}
      for _,tres := range(resultsList){
        rs := &Result{}
        res := tres.(map[string]interface{})
        rs.Time = time.Duration(res["time"].(float64))*time.Nanosecond
        if _,ok := res["desc"];ok{
          rs.Desc = res["desc"].(string)
        }
        if tracepath,ok := res["tracePath"];ok{
          rs.TracePath = tracepath.(string)
        }
        if tracesize,ok := res["traceSize"];ok{
          rs.TraceSize = int(tracesize.(float64))
        }
        if stacksize,ok := res["stackSize"];ok{
          rs.StackSize = int(stacksize.(float64))
        }
        if eventslen,ok := res["eventsLen"];ok{
          rs.EventsLen = int(eventslen.(float64))
        }
        if totalG,ok := res["totalg"];ok{
          rs.TotalG = int(totalG.(float64))
        }
        if totalCh,ok := res["totalch"];ok{
          rs.TotalCh =int(totalCh.(float64))
        }
        rs.Detected = res["detected"].(bool)
        results = append(results,rs)
      }
      gex.Results = results
      gex.ID = fields["goatid"].(string)
      gex.Bound = int(fields["goatBound"].(float64))
      gex.TraceDir = fields["traceDir"].(string)
      gex.LastFailedTrace = fields["lastFailedTrace"].(string)
      gex.LastSuccessTrace = fields["lastSuccessTrace"].(string)
      gex.FirstFailedAfter = int(fields["firstFailedAfter"].(float64))
      ret[k]=gex

    }else if strings.HasPrefix(k,"prime") || strings.HasPrefix(k,"ECT"){ // ECTExperiment
      ex := &ECTExperiment{}
      fields := v.(map[string]interface{})
      ex.Timeout = int(fields["timeout"].(float64))
      ex.Cpu = int(fields["cpu"].(float64))
      ex.PrefixDir = fields["prefixDir"].(string)
      ex.BinaryName = fields["binaryName"].(string)
      ex.OutPath = fields["outPath"].(string)

      resultsList := fields["results"].([]interface{})
      results := []*Result{}
      for _,tres := range(resultsList){
        rs := &Result{}
        res := tres.(map[string]interface{})
        rs.Time = time.Duration(res["time"].(float64))*time.Nanosecond
        if desc,ok := res["desc"];ok{
          rs.Desc = desc.(string)
        }
        if tracepath,ok := res["tracePath"];ok{
          rs.TracePath = tracepath.(string)
        }
        if tracesize,ok := res["traceSize"];ok{
          rs.TraceSize = int(tracesize.(float64))
        }
        if stacksize,ok := res["stackSize"];ok{
          rs.StackSize = int(stacksize.(float64))
        }
        if eventslen,ok := res["eventsLen"];ok{
          rs.EventsLen = int(eventslen.(float64))
        }
        if totalG,ok := res["totalg"];ok{
          rs.TotalG = int(totalG.(float64))
        }
        if totalCh,ok := res["totalch"];ok{
          rs.TotalCh = int(totalCh.(float64))
        }
        results = append(results,rs)
      }
      ex.Results = results
      argsList := fields["args"].([]interface{})
      args := []string{}
      for _,targ := range(argsList){
        arg := targ.(string)
        args = append(args,arg)
      }
      ex.Args = args
      ex.ID = fields["ID"].(string)
      ex.GoVer = fields["goVer"].(string)
      //ex.TraceDir = fields["traceDir"].(string)
      ret[k]=ex
    } else{ //ToolExperiment
      tex := &ToolExperiment{}
      fields := v.(map[string]interface{})
      tex.Timeout = int(fields["timeout"].(float64))
      tex.Cpu = int(fields["cpu"].(float64))
      tex.PrefixDir = fields["prefixDir"].(string)
      tex.BinaryName = fields["binaryName"].(string)
      tex.OutPath = fields["outPath"].(string)

      resultsList := fields["results"].([]interface{})
      results := []*Result{}
      for _,tres := range(resultsList){
        rs := &Result{}
        res := tres.(map[string]interface{})
        rs.Time = time.Duration(res["time"].(float64))*time.Nanosecond
        if _,ok := res["desc"];ok{
          rs.Desc = res["desc"].(string)
        }
        rs.Detected = res["detected"].(bool)
        results = append(results,rs)
      }
      tex.Results = results
      tex.ToolID = fields["toolid"].(string)
      ret[k]=tex
    }
  }
  return ret
}

// Read concUsage
func ReadConcUsage(fname string) (ret []*instrument.ConcurrencyUsage){
  fmt.Println("Reading ",fname)
  var dat []interface{}
  if !checkFile(fname){
    panic("Error reading JSON file: "+fname)
  }
  bf ,_ := ioutil.ReadFile(fname)
  err := json.Unmarshal([]byte(bf),&dat)
  check(err)
  // load from json
  //dat[Bug]
  // we want to read dat[exps]
  if len(dat) == 0 {
    panic("ConcUsage JSON length is zero")
  }
  //fmt.Println("Len Dat: ", len(dat))
  for _,cu_raw := range(dat){
    cui := cu_raw.(map[string]interface{})
    CU := &instrument.ConcurrencyUsage{}
    CU.Type = int(cui["type"].(float64))
    cuu := cui["location"].(map[string]interface{})
    CU_Loc := &instrument.CodeLocation{}
    CU_Loc.FileName = cuu["fileName"].(string)
    CU_Loc.Line = int(cuu["line"].(float64))
    CU.Location = CU_Loc
    ret = append(ret,CU)
  }
  return ret
}
