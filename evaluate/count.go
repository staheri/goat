package evaluate

import (
  "github.com/staheri/goatlib/trace"
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
		desc := EventDescriptions[e.Type]
    if contains(ctgDescriptions[catCHNL].Members, "Ev"+desc.Name){
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
