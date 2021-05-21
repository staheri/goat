// Generates tables and reports
package evaluate

import(
  "fmt"
  "os"
  "github.com/jedib0t/go-pretty/table"
  "strings"
  "sort"
)

var btools = []string{"builtinDL","goleak","lockDL","goat_d0","goat_d1","goat_d2","goat_d3"}
var nbtools = []string{"race","goat_race_d0","goat_race_d1","goat_race_d2","goat_race_d3","goat_race_d4","goat_race_d5","goat_race_d6","goat_race_d7"}


func TableSummaryPerBug(rx *RootExperiment){
  t := table.NewWriter()
  t.SetOutputMirror(os.Stdout)
  t.AppendHeader(table.Row{"Experiment","Detect After","Error","Avg. Time","Tot. Time"})
  for k,vv := range(rx.Exps){
    switch vv.(type){
    case *GoatExperiment:
      v := vv.(*GoatExperiment)

      var row []interface{}
      row = append(row,k)
      //fmt.Fprintf(evaluateCSV,"%v,",k)
      row = append(row,len(v.Results))
      //fmt.Fprintf(evaluateCSV,"%v,",v.Result.FailAfter)
      row = append(row,v.Results[len(v.Results)-1].Desc)
      //fmt.Fprintf(evaluateCSV,"%v,",v.Result.Desc)
      tot := 0.0
      for _,tt := range(v.Results){
        tot = tot + float64((tt.Time).Milliseconds())/1000.0
      }
      //row = append(row,float64(base.Sub(ts)/time.Millisecond)/float64(len(v.Result.Time)))
      if len(v.Results) != 0{
        row = append(row,fmt.Sprintf("%.3f",tot/float64(len(v.Results))))
        //fmt.Fprintf(evaluateCSV,"%.3f,",tot/float64(len(v.Result.Time)))
      } else{
        row = append(row,"-")
        //fmt.Fprintf(evaluateCSV,"%v,","-")
      }

      row = append(row,fmt.Sprintf("%.3f",tot))

      t.AppendRow(row)
    case *ToolExperiment:
      v := vv.(*ToolExperiment)
      var row []interface{}
      row = append(row,k)
      //fmt.Fprintf(evaluateCSV,"%v,",k)
      row = append(row,len(v.Results))
      //fmt.Fprintf(evaluateCSV,"%v,",v.Result.FailAfter)
      row = append(row,v.Results[len(v.Results)-1].Desc)
      //fmt.Fprintf(evaluateCSV,"%v,",v.Result.Desc)
      tot := 0.0
      for _,tt := range(v.Results){
        tot = tot + float64((tt.Time).Milliseconds())/1000.0
      }
      //row = append(row,float64(base.Sub(ts)/time.Millisecond)/float64(len(v.Result.Time)))
      if len(v.Results) != 0{
        row = append(row,fmt.Sprintf("%.3f",tot/float64(len(v.Results))))
        //fmt.Fprintf(evaluateCSV,"%.3f,",tot/float64(len(v.Result.Time)))
      } else{
        row = append(row,"-")
        //fmt.Fprintf(evaluateCSV,"%v,","-")
      }

      row = append(row,fmt.Sprintf("%.3f",tot))

      t.AppendRow(row)
    }
  }
  //evaluateCSV.Flush()
  fmt.Println("Bug: ",rx.Bug.BugName)
  t.Render()
}


func Table_Bug_Tool(bugs map[string]*RootExperiment, order int, identifier string){
  // Variables
  dat := make(map[string][]*RootExperiment)
  var key string
  keys := []string{}
  totals := make([]int,TOOL_COUNT)
  var tools []string

  // first pass (categorize)
  for bug,mainExp := range(bugs){
    switch order {
    case ORDER_CAUSE:
      key = mainExp.Bug.BugCause
    case ORDER_SUBCAUSE:
      key = mainExp.Bug.BugSubCause
    default:
      key = bug
    }
    if lb,ok := dat[key];ok{
      dat[key] = append(lb,mainExp)
    } else{
      dat[key] = []*RootExperiment{mainExp}
    }

    if !contains(keys,key){
      keys = append(keys,key)
    }
  }

  if identifier == "blocking"{
    tools = btools
  }else{
    tools = nbtools
  }
  // create table
  t := table.NewWriter()
  t.SetOutputMirror(os.Stdout)

  // create header
  var headerRow []interface{}
  switch order {
  case ORDER_SUBCAUSE:
    headerRow = append(headerRow,"SubCause")
    headerRow = append(headerRow,"Cause")
    headerRow = append(headerRow,"Bug")
  case ORDER_CAUSE:
    headerRow = append(headerRow,"Cause")
    headerRow = append(headerRow,"SubCause")
    headerRow = append(headerRow,"Bug")
  default: // ORDER_BUG or else
    headerRow = append(headerRow,"Bug")
    headerRow = append(headerRow,"Cause")
    headerRow = append(headerRow,"SubCause")
  }
  for _,t := range(tools){
    headerRow = append(headerRow,strings.ToUpper(t))
  }
  t.AppendHeader(headerRow)

  sort.Strings(keys)

  for _,key := range(keys){
    rex := dat[key]
    // sorting
    sort.Slice(rex, func(i,j int) bool{
      return rex[i].Bug.BugName < rex[j].Bug.BugName
    })
    // end sort
    for _,ex := range(rex){
      var row []interface{}
      row = append(row,key)
      switch order {
      case ORDER_SUBCAUSE:
        row = append(row,ex.Bug.BugCause)
        row = append(row,ex.Bug.BugName)
      case ORDER_CAUSE:
        row = append(row,ex.Bug.BugSubCause)
        row = append(row,ex.Bug.BugName)
      default: // ORDER_BUG or else
        row = append(row,ex.Bug.BugCause)
        row = append(row,ex.Bug.BugSubCause)
      }
      // for each tool, check its result
      for i,t := range(tools){
        res := ""
        detected := false
        switch ex.Exps[t].(type){
        case *GoatExperiment:
          exp := ex.Exps[t].(*GoatExperiment)
          res,detected = resultsToStringDescription(exp.Results)
        case *ToolExperiment:
          exp := ex.Exps[t].(*ToolExperiment)
          res,detected = resultsToStringDescription(exp.Results)
        }
        if detected {
          totals[i]++
        }
        row = append(row,res)
      }
      t.AppendRow(row)
    }
    t.AppendSeparator()
  }
  // total row
  var row []interface{}
  row = append(row,"-")
  row = append(row,"-")
  row = append(row,"-")
  for _,tot := range(totals){
    row = append(row,tot)
  }
  t.AppendRow(row)
  t.Render()
  t.RenderCSV()
}


func resultsToStringFailAfter(results []*Result) (string, bool) {
  failafter := len(results)
  return fmt.Sprintf("%d",failafter) , results[failafter-1].Detected
}

func resultsToStringDescription(results []*Result) (string, bool) {
  ret := ""
  failafter := len(results)
  if results[failafter-1].Detected{
    ret = results[failafter-1].Desc
  } else {
    ret = "X"
  }
  return fmt.Sprintf("%s (%d)",ret,failafter) , results[failafter-1].Detected
}

func CoverageSummaryPerExp(ex Ex){
  t := table.NewWriter()
  t.SetOutputMirror(os.Stdout)
  t.AppendHeader(table.Row{"Test","Cov1 (%)","Cov2 (%)","Error"})

  switch ex.(type){
  case *GoatExperiment:
    gex := ex.(*GoatExperiment)
    for i,resg := range(gex.Results){
      var row []interface{}
      row = append(row,fmt.Sprintf("%v on %v (%d)\n",gex.Target.BugName,gex.ID,i+1))
      row = append(row,fmt.Sprintf("%.2f",resg.Coverage1*100))
      row = append(row,fmt.Sprintf("%.2f",resg.Coverage2*100))
      if resg.Detected{
      	row = append(row,resg.Desc)
      }else{
        row = append(row,"-")
      }
      t.AppendRow(row)
    }
  }
  t.Render()

}

func CoverageSummary(ex Ex) ([]interface{}){
  first_fail := 0
  errs := []string{}
  prev_cov1 := 0.0
  prev_cov2 := 0.0
  testName := ""
  delayBound := 0
  coverage1Leaps := []int{}
  coverage2Leaps := []int{}
  c1gc2 := false
  finalCoverage := 0.0
  t := table.NewWriter()
  t.SetOutputMirror(os.Stdout)
  t.AppendHeader(table.Row{"Bug","Delay Bound","Cov1 (%)","Cov2 (%)","Error"})

  switch ex.(type){
  case *GoatExperiment:
    gex := ex.(*GoatExperiment)
    for i,resg := range(gex.Results){
      var row []interface{}
      testName = fmt.Sprintf("%v_%v",gex.Target.BugType,gex.Target.BugName)
      delayBound = gex.Bound
      row = append(row,testName)
      row = append(row,gex.Bound)
      if resg.Coverage1 != prev_cov1{
        coverage1Leaps = append(coverage1Leaps,i+1)
        prev_cov1 = resg.Coverage1
      }

      if resg.Coverage2 != prev_cov2{
        coverage2Leaps = append(coverage2Leaps,i+1)
        prev_cov2 = resg.Coverage2
      }

      finalCoverage = resg.Coverage2
      if resg.Coverage1 > resg.Coverage2 {
        c1gc2=true
        finalCoverage = resg.Coverage1
      }
      row = append(row,fmt.Sprintf("%.2f",resg.Coverage1*100))
      row = append(row,fmt.Sprintf("%.2f",resg.Coverage2*100))
      if resg.Detected{
      	row = append(row,resg.Desc)
        if !contains(errs,resg.Desc){
          errs = append(errs,resg.Desc)
        }
        if first_fail == 0{
          first_fail = i+1
        }
      }else{
        row = append(row,"-")
      }
      t.AppendRow(row)
    }
  }
  t.Render()

  t1 := table.NewWriter()
  t1.SetOutputMirror(os.Stdout)
  t1.AppendHeader(table.Row{"Bug","Delay Bound","1st Fail","Cov1 Leaps", "Cov2 Leaps","Errors","Cov1 > Cov2","Final Coverage"})
  var row []interface{}
  row = append(row,testName)
  row = append(row,delayBound)
  row = append(row,first_fail)
  row = append(row,fmt.Sprintf("%v",coverage1Leaps))
  row = append(row,fmt.Sprintf("%v",coverage2Leaps))
  row = append(row,fmt.Sprintf("%v",errs))
  row = append(row,fmt.Sprintf("%v",c1gc2))
  row = append(row,fmt.Sprintf("%.2f",finalCoverage))
  t1.AppendRow(row)
  t1.Render()
  return row
}


const(
  BUILTINDL      =iota
  GOLEAK
  LOCKDL
  GOAT_M
  GOAT_U
  GOAT_D1
  GOAT_D2
  GOAT_D3
  TOOL_COUNT
)

const(
  ORDER_BUG     = iota
  ORDER_CAUSE
  ORDER_SUBCAUSE
)
