// Generates tables and reports
package evaluate

import(
  "fmt"
  "os"
  _"github.com/staheri/goatlib/instrument"
  "github.com/jedib0t/go-pretty/table"
)


func SummaryTable(rx *RootExperiment){
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




type Report struct{
  tool          string
  pdl           []string
  gdl           []string
  abort         []string
  crash         []string
  other         []string
  failed        []string
  numExec       float64
  timeExec      float64
}

const(
  BUILTINDL      =iota
  GOLEAK
  LOCKDL
  GOAT_MULTI
  GOAT_UNI
  GOAT_DELAY1
  GOAT_DELAY2
  GOAT_DELAY3
  SUBEX_COUNT
)
