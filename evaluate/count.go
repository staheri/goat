package evaluate

import (
  "github.com/staheri/goatlib/trace"
)

type Count struct{
  G           float64
  Ch          float64
  TotalChOps  float64
}

func countGCH(events []*trace.Event) (int,int){
  gs := make(map[int]int)
  cids := make(map[int]int)
  for _,e := range events{
		desc := EventDescriptions[e.Type]
    if contains(ctgDescriptions[catCHNL].Members, "Ev"+desc.Name){
			// if channel op, finds ID
			cids[int(e.Args[0])] = 1
		}
    gs[int(e.G)] = 1
  }
  return len(gs),len(cids)
}
