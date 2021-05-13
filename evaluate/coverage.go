// Implements Experiments and their methods
package evaluate

import(
  //"bufio"
  "github.com/staheri/goatlib/instrument"
  "github.com/staheri/goatlib/trace"
  "github.com/staheri/goatlib/traceops"
  "fmt"
  // "os"
  "strconv"
  "strings"
  // "path/filepath"
  // _"time"
  // "os/exec"
  "sort"

)


type ConcUsageStruct struct{
  ConcUsage            []*instrument.ConcurrencyUsage       `json:"concUsage"`
  ConcUsageMap         map[string]int      // key: cu.string, val: cu index
  ConcUsageStackMap    map[string]int      // key: fkey for cu, val: cu index
  ConcUsageStackFekys  []string            // list of all fkeys
}

// map concurrency usage index to its respective string representation
func (gex *GoatExperiment) InitConcMap(){
  gex.ConcUsage.ConcUsageMap = make(map[string]int)
  gex.ConcUsage.ConcUsageStackMap = make(map[string]int)
  for i,cu := range(gex.ConcUsage.ConcUsage){
    gex.ConcUsage.ConcUsageMap[cu.String()]=i
  }
}

// after creating lstack (updating gstack), update concUsage with their respective fkeys
func (gex *GoatExperiment) UpdateConcUsage(stacks map[uint64][]*trace.Frame, lstack map[uint64]string){
  for stack_id,frms := range(stacks){
		// iterate over frames
		for _,frm := range(frms){
			// iterate concUsage
			for idx,cu := range(gex.ConcUsage.ConcUsage){
				// check file and line
				//fmt.Printf("CHECK OK\nCU File:%s\nStack file:%s\n",cu.OrigLoc.Filename,frm.File)
				if cu.OrigLoc.Filename == frm.File {
					//fmt.Println("file ok")
					if cu.OrigLoc.Line == frm.Line{
						//fmt.Println("line ok")
						if _,ok := lstack[stack_id] ; !ok{
              panic("coverage frame is not in the lstack")
            }
            gex.ConcUsage.ConcUsageStackMap[lstack[stack_id]] = idx
            gex.ConcUsage.ConcUsageStackFekys = append(gex.ConcUsage.ConcUsageStackFekys,lstack[stack_id])
					}
				}
			}
		}
	}
}



type GlobalStack struct{
  FrameMap        map[int]*trace.Frame // key: unique id, value: frame
  FrameSMap       map[string]int       // key: fkey, value: unique id
}

// Updates GStack and returns compatible LStack
func (gex *GoatExperiment) UpdateGStack(stack map[uint64][]*trace.Frame) map[uint64]string {
  // new local stack to return
  lstack := make(map[uint64]string)

  // iterate over stack frames to update the global stack
  for stack_id,frms := range(stack){
    fkeySlice := []string{}
    for _,frm := range(frms){
      frameKey := traceops.ToKey(frm)
      if _,ok := gex.GStack.FrameSMap[frameKey]; !ok{
        gex.GStack.FrameMap[len(gex.GStack.FrameMap)] = frm
        gex.GStack.FrameSMap[frameKey] = len(gex.GStack.FrameMap)-1
      }
      fkeySlice = append(fkeySlice,strconv.Itoa(gex.GStack.FrameSMap[frameKey]))
    }
    lstack[stack_id]=strings.Join(fkeySlice,".")
  }

  return lstack
}


// Nodes of GGTree
type GGInfo struct{
  id              int    // unique id
  createFkey      string // frame key of create stack
  CoverageMap     map[int]*Coverage // global structure to store general coverages. key: cuIndex, val: coverage instance
}

type GGTree struct{
  Node           *GGInfo
  Children       []*GGTree
}


type Coverage struct{
  blocked        int  // send, recv, select, lock, wait
  blocking       int  // lock
  unblocking     int  // send, recv, select, unlock, add, sig, bcast, close
  none           int  // all except lock
  selecti        []int
}

func (gex *GoatExperiment) InitGGTree(parseResult *trace.ParseResult,lstack map[uint64]string) {
  gtree := traceops.GetGTree(parseResult) // obtain local gtree
  gcount := 0 // init counter
  covMap := make(map[int]*Coverage) // placeholder for concusage
  gginfo := &GGInfo{id:gcount,createFkey:lstack[gtree.Node.CreateStack_id],CoverageMap:covMap} // create node
  gcount++ // increment counter


  gex.GGTree = &GGTree{Node:gginfo} // assign root node to gex.GGTree

  tovisit := []*traceops.GTree{gtree}
  tovisitg := []*GGTree{gex.GGTree}
  for ;len(tovisit)!=0;{
		cur := tovisit[0]
    curg := tovisitg[0]
    for _,child := range(cur.Children){ // iterate over local gtree childs to create global ggtree nodes based on them
      covMap2 := make(map[int]*Coverage)
      gginf := &GGInfo{id:gcount,createFkey:lstack[child.Node.CreateStack_id],CoverageMap:covMap2}
      gcount++
      ggt := &GGTree{Node:gginf}
      curg.Children = append(curg.Children,ggt)
      tovisit = append(tovisit,child)
      tovisitg = append(tovisitg,ggt)
    }
    tovisit = tovisit[1:]
    tovisitg = tovisitg[1:]
  }
}

func (gex *GoatExperiment) CheckUpdateGGTree(parseResult *trace.ParseResult, lstack map[uint64]string) {
  gtree := traceops.GetGTree(parseResult)

  tovisit := []*traceops.GTree{gtree}
  tovisitg := []*GGTree{gex.GGTree}

  cuStackKeys := gex.ConcUsage.ConcUsageStackFekys
  cuStack := gex.ConcUsage.ConcUsageStackMap

  for ;len(tovisit)!=0;{
		cur := tovisit[0]
    curg := tovisitg[0]
    // check if their fkey is the same
    if lstack[cur.Node.CreateStack_id] != curg.Node.createFkey{
      panic("incompatible fkey for current gtree and global gtree")
    }
    if len(cur.Children) != len(curg.Children){
      panic("incompatible children counts")
    }
    for idx,e := range(cur.Node.Events){
  		//fmt.Println(e.String())
  		ed := trace.EventDescriptions[e.Type]
  		// check for HB unblock
  		// check for concurrency usage

  		if contains(cuStackKeys,lstack[e.StkID]){
  			//fmt.Printf("***CONC***\n-------\n")
  			switch gex.ConcUsage.ConcUsage[cuStack[lstack[e.StkID]]].Type{
  			case instrument.LOCK, instrument.UNLOCK, instrument.RUNLOCK, instrument.RLOCK:
  				if !strings.HasPrefix(ed.Name,"Mu"){
  					continue
  				}
          // LOCK
          if gex.ConcUsage.ConcUsage[cuStack[lstack[e.StkID]]].Type == instrument.LOCK{
            if e.Args[1] == 0{
              fmt.Println(e.String())
              fmt.Println("LOCK: Blocked")
              if cm,ok := curg.Node.CoverageMap[cuStack[lstack[e.StkID]]];ok{
                cm.blocked++
              }else{
                curg.Node.CoverageMap[cuStack[lstack[e.StkID]]]=&Coverage{blocked:1}
              }
            }else if e.Args[1] == 1{
              fmt.Println(e.String())
              fmt.Println("LOCK: Blocking")
              if cm,ok := curg.Node.CoverageMap[cuStack[lstack[e.StkID]]];ok{
                cm.blocking++
              }else{
                curg.Node.CoverageMap[cuStack[lstack[e.StkID]]]=&Coverage{blocking:1}
              }
            }
          }
          //UNLOCK
          if gex.ConcUsage.ConcUsage[cuStack[lstack[e.StkID]]].Type == instrument.UNLOCK{
            // check if its next event is unblock
            if trace.EventDescriptions[cur.Node.Events[idx+1].Type].Name == "GoUnblock"{
              fmt.Println(e.String())
              fmt.Println(cur.Node.Events[idx+1].String())
              fmt.Println("UNLOCK: Unblocking")
              if cm,ok := curg.Node.CoverageMap[cuStack[lstack[e.StkID]]];ok{
                cm.unblocking++
              } else{
                curg.Node.CoverageMap[cuStack[lstack[e.StkID]]]=&Coverage{unblocking:1}
              }
            }else{
              fmt.Println(e.String())
              fmt.Println("UNLOCK: None")
              if cm,ok := curg.Node.CoverageMap[cuStack[lstack[e.StkID]]];ok{
                cm.none++
              } else{
                curg.Node.CoverageMap[cuStack[lstack[e.StkID]]]=&Coverage{none:1}
              }
            }
          }
  			case instrument.SEND, instrument.RECV, instrument.CLOSE:
  				if !strings.HasPrefix(ed.Name,"Ch"){
  					continue
  				}

          // CLOSE
          if gex.ConcUsage.ConcUsage[cuStack[lstack[e.StkID]]].Type == instrument.CLOSE{
            // check if its next event is unblock
            if trace.EventDescriptions[cur.Node.Events[idx+1].Type].Name == "GoUnblock"{
              fmt.Println(e.String())
              fmt.Println(cur.Node.Events[idx+1].String())
              fmt.Println("CLOSE: Unblocking")
              if cm,ok := curg.Node.CoverageMap[cuStack[lstack[e.StkID]]];ok{
                cm.unblocking++
              } else{
                curg.Node.CoverageMap[cuStack[lstack[e.StkID]]]=&Coverage{unblocking:1}
              }
            }else{
              fmt.Println(e.String())
              fmt.Println("CLOSE: None")
              if cm,ok := curg.Node.CoverageMap[cuStack[lstack[e.StkID]]];ok{
                cm.none++
              } else{
                curg.Node.CoverageMap[cuStack[lstack[e.StkID]]]=&Coverage{none:1}
              }
            }
          } else{// SEND/RECV
            if e.Args[1] == 0{
              fmt.Println(e.String())
              fmt.Println("SEND/RECV: Blocked")
              if cm,ok := curg.Node.CoverageMap[cuStack[lstack[e.StkID]]];ok{
                cm.blocked++
              }else{
                curg.Node.CoverageMap[cuStack[lstack[e.StkID]]]=&Coverage{blocked:1}
              }
            }else if trace.EventDescriptions[cur.Node.Events[idx+1].Type].Name == "GoUnblock"{
              fmt.Println(e.String())
              fmt.Println(cur.Node.Events[idx+1].String())
              fmt.Println("SEND/RECV: Unblocking")
              if cm,ok := curg.Node.CoverageMap[cuStack[lstack[e.StkID]]];ok{
                cm.unblocking++
              } else{
                curg.Node.CoverageMap[cuStack[lstack[e.StkID]]]=&Coverage{unblocking:1}
              }
            } else if e.Args[1] != 2{
              fmt.Println("SEND/RECV: None")
              if cm,ok := curg.Node.CoverageMap[cuStack[lstack[e.StkID]]];ok{
                cm.none++
              } else{
                curg.Node.CoverageMap[cuStack[lstack[e.StkID]]]=&Coverage{none:1}
              }
            }
          }
  			case instrument.SELECT:
  				if !strings.HasPrefix(ed.Name,"Select"){
  					continue
  				}

  			}
  			// var row []interface{}
  			// row = append(row,e.StkID)
  			// //row = append(row,concStackTable[e.StkID].OrigLoc.Filename)
  			// row = append(row,cuStack[lstack[e.StkID]].OrigLoc.Function)
  			// row = append(row,cuStack[lstack[e.StkID]].OrigLoc.Line)
  			// row = append(row,ed.Name)
  			// row = append(row,e.G)
  			// row = append(row,instrument.ConcTypeDescription[cuStack[lstack[e.StkID]].Type])
  			// row = append(row,GetPositionDesc(e))
  			// t.AppendRow(row)

        // for each node, we are iterating over its events.
        // we only care about events that their type matches events and stack matches ConcUsage
        // now we update curg.Node.CoverageMap[cuStack[lstack[e.StkID]]].
  		}
  	}

    tovisit = append(tovisit,cur.Children...)
    tovisitg = append(tovisitg,curg.Children...)
    tovisit = tovisit[1:]
    tovisitg = tovisitg[1:]
  }
}



func (gex *GoatExperiment) PrintGlobals(){
  fmt.Println("ConcUsage Struct: ConcUsage")
  for i,cu := range(gex.ConcUsage.ConcUsage){
    fmt.Printf("%d: %s\n",i,cu.String())
  }
  /*fmt.Println("ConcUsage Struct: ConcUsageMap")
  for k,v := range(gex.ConcUsage.ConcUsageMap){
    fmt.Printf("%v: %v\n",k,v)
  }
  fmt.Println("ConcUsage Struct: ConcUsageStackMap")
  for k,v := range(gex.ConcUsage.ConcUsageStackMap){
    fmt.Printf("%v: %v\n",k,v)
  }

  fmt.Println("Global Stack: Frame Map")
  for k,v := range(gex.GStack.FrameMap){
    fmt.Printf("%v: %v\n",k,traceops.ToString(v))
  }

  fmt.Println("Global Stack: Frame SMap")
  for k,v := range(gex.GStack.FrameSMap){
    fmt.Printf("%v: %v\n",k,v)
  }*/

}


func (cov *Coverage)ToString() string{
  s := ""
  s = s + fmt.Sprintf("blocked: %v, ",cov.blocked)
  s = s + fmt.Sprintf("blocking: %v, ",cov.blocking)
  s = s + fmt.Sprintf("unblocking: %v, ",cov.unblocking)
  s = s + fmt.Sprintf("none: %v",cov.none)
  return s
}

func (gi *GGInfo) ToString(concUsage []*instrument.ConcurrencyUsage) string{
  s := fmt.Sprintf("**** GGINFO: %d ****\n",gi.id)
  s = s + fmt.Sprintf("\tcreateFkey: %v\n",gi.createFkey)
  s = s + fmt.Sprintf("\tCoverageMap:\n")
  // sort map
  concUsageIndex  := []int{}
  for i,_ := range(gi.CoverageMap){
    concUsageIndex = append(concUsageIndex,i)
  }
  sort.Ints(concUsageIndex)
  for _,i := range(concUsageIndex){
    s = s + fmt.Sprintf("\t\t[%v]: %v\n",concUsage[i].String(),gi.CoverageMap[i].ToString())
  }
  s = s + fmt.Sprintf("********************\n")
  return s
}


func PrintGGTree(root *GGTree,concUsage []*instrument.ConcurrencyUsage){
  tovisit := []*GGTree{root}
  for ;len(tovisit)!=0;{
		cur := tovisit[0]
    fmt.Println(cur.ToString(concUsage))
    for _,child := range(cur.Children){ // iterate over local gtree childs to create global ggtree nodes based on them
      tovisit = append(tovisit,child)
    }
    tovisit = tovisit[1:]
  }
}

func (t *GGTree) ToString(concUsage []*instrument.ConcurrencyUsage) string {
  s := fmt.Sprintf("-----------\nNode ID: %d\n",t.Node.id)
  s = s + t.Node.ToString(concUsage)
  s = s + fmt.Sprintf("Children IDs:[ ")
  for _,child := range(t.Children){
    s = s + fmt.Sprintf("%v, ",child.Node.id)
  }
  s = s + fmt.Sprintf("]\n-----------\n")
  return s
}


/*func (gex *GoatExperiment) UpdateCoverageTable(lstack map[uint64]string,events []*trace.Events){
  // maintaining a local g structure
  localGmap := make(map[string][]uint64) // key: fkey, value: [] of local gids
  localGmap_rev := make(map[uint64]string) // key: fkey, value: [] of local gids
  gs,gmap := traceops.GetGoroutineInfo(parseResult)
  localGmap[lstack[gs.Main.CreateStack_id]]=[]int{gs.Main.gid}
  localGmap_rev[gs.Main.gid] = lstack[gs.Main.CreateStack_id]
  for _,gapp := range(gs.App){
    localGmap[lstack[gapp.CreateStack_id]] = append(localGmap[lstack[gapp.CreateStack_id]],gapp.gid)
    localGmap_rev[gapp.gid] = lstack[gapp.CreateStack_id]
  }

  // compare it agains global g structure
  // update if something is new
  for fkey,gids := range(localGmap){
    if ggs,ok := gex.GGMapp[fkey];ok{
      if len(ggs) != len(gids){
        // there are some goroutines in the localGmap that are created
      }
    }
  }
  // add main g to the global g structure


    // add app gs
}*/
