package evaluate

import (
  "github.com/staheri/goatlib/trace"
  "strings"
)




type Count struct{
  G           int
  Ch          int
  TotalChOps  int
}

func count(events []*trace.Event) (cnt Count){
  gs := make(map[int]int)
  cids := make(map[int]int)
  for _,e := range events{
		desc := trace.EventDescriptions[e.Type]
    if strings.HasPrefix(desc.Name,"Ch"){
			// if channel op, finds ID
			cids[int(e.Args[0])] = 1
      cnt.TotalChOps++
		}
    gs[int(e.G)] = 1
  }
  cnt.G = len(gs)
  cnt.Ch = len(cids)
  return cnt
}
