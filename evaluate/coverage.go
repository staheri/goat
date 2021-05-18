// Implements Experiments and their methods
package evaluate

import(
  //"bufio"
  "github.com/staheri/goatlib/instrument"
  "github.com/staheri/goatlib/trace"
  "github.com/staheri/goatlib/traceops"
  	"github.com/jedib0t/go-pretty/table"
  "fmt"
   "os"
  "strconv"
  "strings"
  // "path/filepath"
  // _"time"
  // "os/exec"
  "sort"

)

type ConcUsageStruct struct{
  ConcUsage            []*instrument.ConcurrencyUsage       `json:"concUsage"`
  ConcUsageMap         map[string]int      // key: cu.string, val:  cu index
  ConcUsageStackMap    map[string][]int      // key: fkey for cu, val: []cu index (there might be multiple concurrent usage that shares a common stack frame)
  ConcUsageStackFekys  []string            // list of all fkeys
}

// map concurrency usage index to its respective string representation
func (gex *GoatExperiment) InitConcMap(){
  gex.ConcUsage.ConcUsageMap = make(map[string]int)
  gex.ConcUsage.ConcUsageStackMap = make(map[string][]int)
  for i,cu := range(gex.ConcUsage.ConcUsage){
    gex.ConcUsage.ConcUsageMap[cu.String()]=i
  }
}

// after creating lstack (updating gstack), update concUsage with their respective fkeys
func (gex *GoatExperiment) UpdateConcUsage(stacks map[uint64][]*trace.Frame, lstack map[uint64]string){
  stackConc := make(map[int]int)
  for idx,_ := range(gex.ConcUsage.ConcUsage){
    //fmt.Printf("ConcUsage[%d]: %v\n",idx,cu.String())
    stackConc[idx]=0
  }
  //for cu,idx := range(gex.ConcUsage.ConcUsageMap){
  //  fmt.Printf("ConcUsageMap[%v]: %d\n",cu,idx)
  //}
  for stack_id,frms := range(stacks){
		// iterate over frames
		for _,frm := range(frms){
			// iterate concUsage
			for idx,cu := range(gex.ConcUsage.ConcUsage){
				// check file and line
				//fmt.Printf("CHECK OK\nCU File:%s\nStack file:%s\n",cu.OrigLoc.Filename,frm.File)
				if cu.OrigLoc.Filename == frm.File {
					//fmt.Println("\tfile ok")
					if cu.OrigLoc.Line == frm.Line{
						//fmt.Println("\t\tline ok")
						if _,ok := lstack[stack_id] ; !ok{
              panic("coverage frame is not in the lstack")
            }
            if idxs,ok := gex.ConcUsage.ConcUsageStackMap[lstack[stack_id]];ok{
              if !containsInt(idxs,idx){
                idxs = append(idxs,idx)
                fmt.Printf("Update ConcUsage:CU:%s\nStack:%s\n",cu.String(),traceops.ToString(frm))
              }
              //gex.ConcUsage.ConcUsageStackMap[lstack[stack_id]] = idxs
            } else{
              fmt.Printf("Update ConcUsage:CU:%s\nStack:%s\n",cu.String(),traceops.ToString(frm))
              gex.ConcUsage.ConcUsageStackMap[lstack[stack_id]] = []int{idx}
            }
            //gex.ConcUsage.ConcUsageStackMap[lstack[stack_id]] = idx
            gex.ConcUsage.ConcUsageStackFekys = append(gex.ConcUsage.ConcUsageStackFekys,lstack[stack_id])
            stackConc[idx]=1
					}
				}
			}
		}
	}

  //for idx,val:=range(stackConc){
  //  fmt.Printf("covered[%d]: %d\n",idx,val)
  //}
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
  Children       map[string][]*GGTree // key: createFkey, val: totally ordered goroutines that were created in fkey
}


type Coverage struct{
  blocked        int  // send, recv, select, lock, wait
  blocking       int  // lock
  unblocking     int  // send, recv, select, unlock, add, sig, bcast, close
  no_op          int  // all except lock
  selecti        map[uint64]*Selecti // map by casei
}

type Selecti struct{
  casei        uint64
  kindi        uint64
  cidi         uint64
  selected     int
  blocked      int
  unblocking   int
  no_op        int
}

func (gex *GoatExperiment) UpdateGGTree(parseResult *trace.ParseResult,lstack map[uint64]string) {
  gtree := traceops.GetGTree(parseResult) // obtain local gtree

  if gex.GGTree == nil{
    // GGTree is not inititated yet (first runt) - initiate it with main
    if gex.TotalGG != 0{
      panic("totalGG is not zero")
    }
    covMap := make(map[int]*Coverage) // placeholder for concusage
    gginfo := &GGInfo{id:gex.TotalGG,createFkey:lstack[gtree.Node.CreateStack_id],CoverageMap:covMap} // create node
    gex.GGTree = &GGTree{Node:gginfo} // assign root node to gex.GGTree
    gex.TotalGG++
  }


  // now iterate over gtree
  // add missing nodes

  tovisit := []*traceops.GTree{gtree}
  tovisitg := []*GGTree{gex.GGTree}
  for ;len(tovisit)!=0;{
    cur := tovisit[0]
    curg := tovisitg[0]
    if curg.Node.createFkey != lstack[cur.Node.CreateStack_id] {
      panic("incompatible create_stack_id for corresponding Gtree and GGtree nodes")
    }

    // make sure there is no nil map for curg.Children
    if curg.Children == nil{
      curg.Children = make(map[string][]*GGTree)
    }

    // store childrens of current node in a map (key: createStackFkey,val:[]children created in that location)
    cur_children := make(map[string][]*traceops.GTree)
    for _,child := range(cur.Children){
      if chx,ok := cur_children[lstack[child.Node.CreateStack_id]];ok{
        chx = append(chx,child)
        cur_children[lstack[child.Node.CreateStack_id]] = chx
      } else{
        cur_children[lstack[child.Node.CreateStack_id]] = []*traceops.GTree{child}
      }
    }
    // we store all children of current node in a map (key: createStackFkey,val:[]children created in that location)
    // now iterate over cur_children and check if there is anything missing in the curg

    for cur_fkey,cur_child := range(cur_children){
      if chgx,ok := curg.Children[cur_fkey] ; ok{
        // this fkey has already been added. Now check length
        tchgx := []*GGTree{}
        for i := 0 ; i< len(cur_child) ; i++{
          if i < len(chgx){
            tovisit = append(tovisit,cur_child[i])
            tovisitg = append(tovisitg,chgx[i])
            tchgx = append(tchgx,chgx[i])
          } else{
            covMap := make(map[int]*Coverage) // placeholder for concusage
            gginfo := &GGInfo{id:gex.TotalGG,createFkey:lstack[cur_child[i].Node.CreateStack_id],CoverageMap:covMap} // create node
            gex.TotalGG++
            chg := &GGTree{Node:gginfo}
            tchgx = append(tchgx, chg)
            tovisit = append(tovisit,cur_child[i])
            tovisitg = append(tovisitg,chg)
          }
        }
        curg.Children[cur_fkey]=tchgx
      } else{
        //curg has no children for the fkey
        tchgx := []*GGTree{}
        for _,ch := range(cur_child){
          // create new child
          covMap := make(map[int]*Coverage) // placeholder for concusage
          gginfo := &GGInfo{id:gex.TotalGG,createFkey:lstack[ch.Node.CreateStack_id],CoverageMap:covMap} // create node
          gex.TotalGG++
          chg := &GGTree{Node:gginfo}
          tchgx = append(tchgx,chg)
          tovisit = append(tovisit,ch)
          tovisitg = append(tovisitg,chg)
        }
        curg.Children[cur_fkey]=tchgx
      }
    } // for all children of current node, we have a corresponding node in GGTree (curg)

    tovisit = tovisit[1:]
    tovisitg = tovisitg[1:]
  }
}

func (gex *GoatExperiment) UpdateCoverageGGTree(parseResult *trace.ParseResult, lstack map[uint64]string) {
  gtree := traceops.GetGTree(parseResult)

  tovisit := []*traceops.GTree{gtree}
  tovisitg := []*GGTree{gex.GGTree}

  cuStackKeys := gex.ConcUsage.ConcUsageStackFekys
  cuStack := gex.ConcUsage.ConcUsageStackMap

  for ;len(tovisit)!=0;{
		cur := tovisit[0]
    curg := tovisitg[0]
    fmt.Printf("Iterating over\n\tG: %v\n\tGG:%v\n\tLen(events):%v\n",cur.Node.Gid,curg.Node.id,len(cur.Node.Events))

    for idx,e := range(cur.Node.Events){
  		//fmt.Println(e.String())
  		ed := trace.EventDescriptions[e.Type]
  		// check for HB unblock
  		// check for concurrency usage

  		if contains(cuStackKeys,lstack[e.StkID]){

        for _,cus_idx := range(cuStack[lstack[e.StkID]]){
          //fmt.Printf("***CONC***\n%v\nlstack[event.Stackid]: %v\ncuStack: %v\n-------\n",gex.ConcUsage.ConcUsage[cus_idx].String(),lstack[e.StkID],cuStack[lstack[e.StkID]])
          switch gex.ConcUsage.ConcUsage[cus_idx].Type{
          case instrument.LOCK, instrument.UNLOCK, instrument.RUNLOCK, instrument.RLOCK:
            if !strings.HasPrefix(ed.Name,"Mu"){
              continue
            }
            // LOCK
            if gex.ConcUsage.ConcUsage[cus_idx].Type == instrument.LOCK{
              if e.Args[1] == 0{
                //fmt.Println(e.String())
                //fmt.Println("LOCK: Blocked")
                if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                  cm.blocked++
                }else{
                  curg.Node.CoverageMap[cus_idx]=&Coverage{blocked:1}
                }
              }else if e.Args[1] == 1{
                //fmt.Println(e.String())
                //fmt.Println("LOCK: Blocking")
                if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                  cm.blocking++
                }else{
                  curg.Node.CoverageMap[cus_idx]=&Coverage{blocking:1}
                }
              }
            }
            //UNLOCK
            if gex.ConcUsage.ConcUsage[cus_idx].Type == instrument.UNLOCK{
              // check if its next event is unblock
              if trace.EventDescriptions[cur.Node.Events[idx+1].Type].Name == "GoUnblock"{
                //fmt.Println(e.String())
                //fmt.Println(cur.Node.Events[idx+1].String())
                //fmt.Println("UNLOCK: Unblocking")
                if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                  cm.unblocking++
                } else{
                  curg.Node.CoverageMap[cus_idx]=&Coverage{unblocking:1}
                }
              }else{
                //fmt.Println(e.String())
                //fmt.Println("UNLOCK: None")
                if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                  cm.no_op++
                } else{
                  curg.Node.CoverageMap[cus_idx]=&Coverage{no_op:1}
                }
              }
            }
          case instrument.SEND, instrument.RECV, instrument.CLOSE:
            if !strings.HasPrefix(ed.Name,"Ch"){
              continue
            }

            // CLOSE
            if gex.ConcUsage.ConcUsage[cus_idx].Type == instrument.CLOSE{
              // check if its next event is unblock
              if trace.EventDescriptions[cur.Node.Events[idx+1].Type].Name == "GoUnblock"{
                //fmt.Println(e.String())
                //fmt.Println(cur.Node.Events[idx+1].String())
                //fmt.Println("CLOSE: Unblocking")
                if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                  cm.unblocking++
                } else{
                  curg.Node.CoverageMap[cus_idx]=&Coverage{unblocking:1}
                }
              }else{
                //fmt.Println(e.String())
                //fmt.Println("CLOSE: None")
                if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                  cm.no_op++
                } else{
                  curg.Node.CoverageMap[cus_idx]=&Coverage{no_op:1}
                }
              }
            } else{// SEND/RECV
              if e.Args[1] == 0{
                //fmt.Println(e.String())
                //fmt.Println("SEND/RECV: Blocked")
                if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                  cm.blocked++
                }else{
                  curg.Node.CoverageMap[cus_idx]=&Coverage{blocked:1}
                }
              }else if trace.EventDescriptions[cur.Node.Events[idx+1].Type].Name == "GoUnblock"{
                //fmt.Println(e.String())
                //fmt.Println(cur.Node.Events[idx+1].String())
                //fmt.Println("SEND/RECV: Unblocking")
                if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                  cm.unblocking++
                } else{
                  curg.Node.CoverageMap[cus_idx]=&Coverage{unblocking:1}
                }
              } else if e.Args[1] != 2{
                //fmt.Println("SEND/RECV: None")
                if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                  cm.no_op++
                } else{
                  curg.Node.CoverageMap[cus_idx]=&Coverage{no_op:1}
                }
              }
            }
          case instrument.SELECT:
            if !strings.HasPrefix(ed.Name,"Select"){
              continue
            }
            if ed.Name == "Selecti"{
              //fmt.Println(e.String())
              //fmt.Println("Selecti")
              // initilize
              if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                // assume that before each EvSelect, we encounter its EvSelecti first
                if cm.selecti == nil{
                  panic("a selecti is encountered before init")
                }
                if _,ok2:=cm.selecti[e.Args[0]];ok2{
                  fmt.Printf("the casei %v is already added",e.Args[0])
                  //fmt.Printf("(old_casei: %v, old_kindi: %v\n",si.casei,si.kindi)
                  //fmt.Printf("(new_casei: %v, new_kindi: %v\n",e.Args[0],e.Args[2])
                }else{
                  selecti := &Selecti{casei:e.Args[0],cidi:e.Args[1],kindi:e.Args[2]}
                  cm.selecti[e.Args[0]]=selecti
                }
              } else{
                fmt.Println("\tNewly added")
                newSelectCoverage := &Coverage{}
                newSelectCoverage.selecti= make(map[uint64]*Selecti)
                selecti := &Selecti{casei:e.Args[0],cidi:e.Args[1],kindi:e.Args[2]}
                newSelectCoverage.selecti[e.Args[0]] = selecti
                curg.Node.CoverageMap[cus_idx]=newSelectCoverage
              }
            }
            if ed.Name == "Select"{
              // update
              pos := e.Args[0]
              casi := e.Args[1]
              if pos == 1{
                // non-blocking --> we have to decide: unblocking / noop
                if trace.EventDescriptions[cur.Node.Events[idx+1].Type].Name == "GoUnblock" || trace.EventDescriptions[cur.Node.Events[idx+2].Type].Name == "GoUnblock"{
                  //fmt.Println(e.String())
                  //fmt.Println(cur.Node.Events[idx+1].String())
                  //fmt.Println(cur.Node.Events[idx+2].String())
                  //fmt.Println("Select: Unblocking")
                  if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                    // which case is selected?
                    cm.unblocking++
                    if cmm,okk := cm.selecti[casi] ; okk{
                      cmm.selected++
                      cmm.unblocking++
                    } else{
                      panic("selected case is not inited in selecti")
                    }
                  } else{
                    panic("select is encountered before selecti")
                    //curg.Node.CoverageMap[cus_idx]=&Coverage{unblocking:1}
                  }
                }else{
                  fmt.Println("SELECT: No-op")
                  if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                    cm.no_op++
                    // which case is selected?
                    if cmm,okk := cm.selecti[casi] ; okk{
                      cmm.selected++
                      cmm.no_op++
                    } else{
                      panic("selected case is not inited in selecti")
                    }
                  } else{
                    panic("select is encountered before selecti")
                    //curg.Node.CoverageMap[cus_idx]=&Coverage{no_op:1}
                  }
                }
              } else if pos == 2{ // select was blocked but now it is unblocked
                //blocking
                //fmt.Println(e.String())
                //fmt.Println("Select: Blocked Then Unblocked")
                if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                  // which case is selected?
                  if cmm,okk := cm.selecti[casi] ; okk{
                    cmm.selected++
                    cmm.blocked++
                  } else{
                    panic("selected case is not inited in selecti")
                  }
                } else{
                  panic("select is encountered before selecti")
                  //curg.Node.CoverageMap[cus_idx]=&Coverage{blocked:1}
                }
              } else{ // pos == 0, select is blocked
                if cm,ok := curg.Node.CoverageMap[cus_idx]; ok{
                  cm.blocked++
                } else{
                  panic("select is encountered before selecti")
                }
              }
            }
          case instrument.WAIT:
            if !strings.HasPrefix(ed.Name,"Wg") && !strings.HasPrefix(ed.Name,"CvWait"){
              continue
            }
            // WgWAIT
            if ed.Name == "WgWait"{
              if e.Args[1] == 0{
                //fmt.Println(e.String())
                //fmt.Println("LOCK: Blocked")
                if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                  cm.blocked++
                }else{
                  curg.Node.CoverageMap[cus_idx]=&Coverage{blocked:1}
                }
              }else if e.Args[1] == 1{
                //fmt.Println(e.String())
                //fmt.Println("LOCK: Blocking")
                if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                  cm.no_op++
                }else{
                  curg.Node.CoverageMap[cus_idx]=&Coverage{no_op:1}
                }
              }
            }else{ // CvWait
              if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                cm.no_op++
              }else{
                curg.Node.CoverageMap[cus_idx]=&Coverage{no_op:1}
              }
            }

          case instrument.DONE:
            if !strings.HasPrefix(ed.Name,"WgAdd"){
              continue
            }
            if e.Args[1] > 0 { // WgAdd -> not interested (we want WgDone which is WgAdd(-1))
              continue
            }
            if e.Args[2] > 0 || e.Args[3]==0{ // wg counter is more than 0 OR nobody is waiting --> no_op
              //fmt.Println(e.String())
              //fmt.Println("LOCK: Blocked")
              if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                cm.no_op++
              }else{
                curg.Node.CoverageMap[cus_idx]=&Coverage{no_op:1}
              }
            }else { //
              //fmt.Println(e.String())
              //fmt.Println("LOCK: Blocking")
              if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
                cm.unblocking++
              }else{
                curg.Node.CoverageMap[cus_idx]=&Coverage{unblocking:1}
              }
            }
          case instrument.SIGNAL,instrument.BROADCAST:
            if !strings.HasPrefix(ed.Name,"Cv"){
              continue
            }
            if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
              cm.no_op++
            }else{
              curg.Node.CoverageMap[cus_idx]=&Coverage{no_op:1}
            }
          case instrument.GO:
            if !strings.HasPrefix(ed.Name,"GoCreate"){
              continue
            }
            if cm,ok := curg.Node.CoverageMap[cus_idx];ok{
              cm.no_op++
            }else{
              curg.Node.CoverageMap[cus_idx]=&Coverage{no_op:1}
            }
          }
        } // end switch concUsage type
  		} // end mapping concusage and local stack (if contains(cuStackKeys,lstack[e.StkID]))
  	}


    // figure next children to check
    cur_children := make(map[string][]*traceops.GTree)
    for _,child := range(cur.Children){
      if chx,ok := cur_children[lstack[child.Node.CreateStack_id]];ok{
        chx = append(chx,child)
        cur_children[lstack[child.Node.CreateStack_id]] = chx
      } else{
        cur_children[lstack[child.Node.CreateStack_id]] = []*traceops.GTree{child}
      }
    }
    for cur_fkey,cur_child := range(cur_children){
      tovisit = append(tovisit,cur_child...)
      tovisitg = append(tovisitg,curg.Children[cur_fkey]...)
    }
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


func (cov *Coverage)ToString(cu *instrument.ConcurrencyUsage) (string,string){

  s := ""
  percent := ""
  total := 0
  //t.AppendHeader(table.Row{"Conc Usage","Blocked","Blocking","Unblocking","No-Op"})
  switch cu.Type{
  case instrument.SEND,instrument.RECV:
    s = s + fmt.Sprintf("blocked: %v, ",cov.blocked)
    if cov.blocked > 0 {
      total++
    }
    s = s + fmt.Sprintf("unblocking: %v, ",cov.unblocking)
    if cov.unblocking > 0 {
      total++
    }
    s = s + fmt.Sprintf("no_op: %v",cov.no_op)
    if cov.no_op > 0 {
      total++
    }
    percent = strconv.Itoa(total)+"/3"
  case instrument.CLOSE,instrument.UNLOCK,instrument.ADD:
    s = s + fmt.Sprintf("unblocking: %v, ",cov.unblocking)
    if cov.unblocking > 0 {
      total++
    }
    s = s + fmt.Sprintf("no_op: %v",cov.no_op)
    if cov.no_op > 0 {
      total++
    }
    percent = strconv.Itoa(total)+"/2"
  case instrument.SELECT:
    if cov.selecti != nil{
      // blocking or non-blocking
      blocking := true
      for casei,_ := range(cov.selecti){
        csi := cov.selecti[uint64(casei)]
        if csi.kindi == 3 {// default case
          blocking = false
        }
      }
      if blocking{
        // we want to list all cases
        for casei,_ := range(cov.selecti){
          csi := cov.selecti[uint64(casei)]
          s = s + fmt.Sprintf("\n\t\tcasei: %v, kindi: %v ,blocked: %v, unblocking: %v, no_op:%v, selected: %v",csi.casei,csi.kindi,csi.blocked,csi.unblocking,csi.no_op,csi.selected)
          if csi.no_op > 0 {
            total++
          }
          if csi.unblocking > 0 {
            total++
          }
          if csi.blocked > 0 {
            total++
          }
        }
        percent = strconv.Itoa(total)+"/"+strconv.Itoa(len(cov.selecti)*3)
      } else{
        for casei,_ := range(cov.selecti){
          csi := cov.selecti[uint64(casei)]
          s = s + fmt.Sprintf("\n\t\tcasei: %v, kindi: %v ,unblocking: %v, no_op:%v, selected: %v",csi.casei,csi.kindi,csi.unblocking,csi.no_op,csi.selected)
          if csi.no_op > 0 {
            total++
          }
          if csi.unblocking > 0 {
            total++
          }
          percent = strconv.Itoa(total)+"/"+strconv.Itoa(len(cov.selecti)*2)
        }
      }
    }
  case instrument.LOCK:
    s = s + fmt.Sprintf("blocked: %v, ",cov.blocked)
    s = s + fmt.Sprintf("blocking: %v, ",cov.blocking)
    if cov.blocking > 0 {
      total++
    }
    if cov.blocked > 0 {
      total++
    }
    percent = strconv.Itoa(total)+"/2"
  case instrument.WAIT:
    s = s + fmt.Sprintf("blocked: %v, ",cov.blocked)
    s = s + fmt.Sprintf("no_op: %v",cov.no_op)
    if cov.no_op > 0 {
      total++
    }
    if cov.blocked > 0 {
      total++
    }
    percent = strconv.Itoa(total)+"/2"
  case instrument.SIGNAL,instrument.BROADCAST,instrument.GO:
    s = s + fmt.Sprintf("no_op: %v",cov.no_op)
    if cov.no_op > 0 {
      total++
    }
    percent = strconv.Itoa(total)+"/1"
  }

  return s,percent
}

func (gi *GGInfo) ToString(concUsage []*instrument.ConcurrencyUsage) (string,string){
  covReq := 0
  covCov := 0
  cov := ""
  s := fmt.Sprintf("<GGINFO: %d>\n",gi.id)
  s = s + fmt.Sprintf("\tcreateFkey: %v\n",gi.createFkey)
  s = s + fmt.Sprintf("\tCoverageMap:\n")
  // sort map
  concUsageIndex  := []int{}
  for i,_ := range(gi.CoverageMap){
    concUsageIndex = append(concUsageIndex,i)
  }
  sort.Ints(concUsageIndex)
  for _,i := range(concUsageIndex){
    st,pcnt := gi.CoverageMap[i].ToString(concUsage[i])
    cr,err := strconv.Atoi(strings.Split(pcnt,"/")[1])
    check(err)
    covReq = covReq + cr
    cc,err := strconv.Atoi(strings.Split(pcnt,"/")[0])
    check(err)
    covCov = covCov + cc

    s = s + fmt.Sprintf("\t\t[%v]: %v (%v)\n",concUsage[i].String(),st,pcnt)
  }

  s = s + fmt.Sprintf("</GGINFO>\n")
  cov = strconv.Itoa(covCov)+"/"+strconv.Itoa(covReq)
  return s,cov
}

func (t *GGTree) ToString(concUsage []*instrument.ConcurrencyUsage) string {
  s := fmt.Sprintf("-----------\nNode ID: %d\n",t.Node.id)
  st,pcnt := t.Node.ToString(concUsage)
  s = s + fmt.Sprintf("\n%v\n((( %v )))\n",st,pcnt)
  s = s + fmt.Sprintf("Children IDs:[ ")
  for _,childs := range(t.Children){
    for _,child := range(childs){
      s = s + fmt.Sprintf("%v, ",child.Node.id)
    }
  }
  s = s + fmt.Sprintf("]\n-----------\n")
  return s
}

func PrintGGTree(root *GGTree,concUsage []*instrument.ConcurrencyUsage){
  tovisit := []*GGTree{root}
  for ;len(tovisit)!=0;{
		cur := tovisit[0]
    st := cur.ToString(concUsage)
    fmt.Println(st)
    for _,child := range(cur.Children){ // iterate over local gtree childs to create global ggtree nodes based on them
      tovisit = append(tovisit,child...)
    }
    tovisit = tovisit[1:]
  }
}


func (gi *GGInfo) CovNodePairs(concUsage []*instrument.ConcurrencyUsage) (map[int]*Pair){
  pairs := make(map[int]*Pair)
  s := fmt.Sprintf("<GGINFO: %d>\n",gi.id)
  s = s + fmt.Sprintf("\tcreateFkey: %v\n",gi.createFkey)
  s = s + fmt.Sprintf("\tCoverageMap:\n")
  // sort map
  concUsageIndex  := []int{}
  for i,_ := range(gi.CoverageMap){
    concUsageIndex = append(concUsageIndex,i)
  }
  sort.Ints(concUsageIndex)
  for _,i := range(concUsageIndex){
    st,pcnt := gi.CoverageMap[i].ToString(concUsage[i])
    covCov,err := strconv.Atoi(strings.Split(pcnt,"/")[0])
    check(err)
    covReq,err := strconv.Atoi(strings.Split(pcnt,"/")[1])
    check(err)
    pairs[i] = &Pair{covCov,covReq}
    s = s + fmt.Sprintf("\t\t[%v]: %v (%v)\n",concUsage[i].String(),st,pcnt)
  }

  s = s + fmt.Sprintf("</GGINFO>\n")
  fmt.Println(s)
  return pairs
}


func (gex *GoatExperiment) CoverageGGTree(){
  pairs := make(map[int]*Pair)
  tovisit := []*GGTree{gex.GGTree}
  for ;len(tovisit)!=0;{
		cur := tovisit[0]
    //st,pcnt = cur.ToString(concUsage)
    //fmt.Println(st)
    covNodePairs := cur.Node.CovNodePairs(gex.ConcUsage.ConcUsage)
    for i,covPair := range(covNodePairs){
      if pr,ok := pairs[i];ok{
        pr.CovCov = pr.CovCov + covPair.CovCov
        pr.CovReq = pr.CovReq + covPair.CovReq
      } else{
        pairs[i] = &Pair{covPair.CovCov,covPair.CovReq}
      }
    }
    for _,child := range(cur.Children){ // iterate over local gtree childs to create global ggtree nodes based on them
      tovisit = append(tovisit,child...)
    }
    tovisit = tovisit[1:]
  }


  t := table.NewWriter()
  t.SetOutputMirror(os.Stdout)
  t.AppendHeader(table.Row{"Conc Usage","CovCov","CovReq","%"})
  totCovCov := 0
  totCovReq := 0

  for i,cu := range(gex.ConcUsage.ConcUsage){
    var row []interface{}
    cuTruncs := strings.Split(cu.String(),"/")
    cuTrunc := cuTruncs[len(cuTruncs)-1]
    row = append(row,cuTrunc)
    if pair,ok := pairs[i] ; ok{
      row = append(row,pair.CovCov)
      totCovCov = totCovCov + pair.CovCov
      row = append(row,pair.CovReq)
      totCovReq = totCovReq + pair.CovReq
      row = append(row,float64(pair.CovCov)/float64(pair.CovReq))
    } else{
      row = append(row,0)
      row = append(row,1)
      row = append(row,float64(0))
    }
    t.AppendRow(row)
  }
  var row []interface{}
  row = append(row,"Total")
  row = append(row,totCovCov)
  row = append(row,totCovReq)
  row = append(row,float64(totCovCov)/float64(totCovReq))
  t.AppendRow(row)
  t.Render()
}

type Pair struct{
  CovCov         int
  CovReq         int
}






/*func (gex *GoatExperiment) PrintCoverageTable(){

  t := table.NewWriter()
  t.SetOutputMirror(os.Stdout)
  t.AppendHeader(table.Row{"Conc Usage","Blocked","Blocking","Unblocking","No-Op"})
  // iterate over concurrency usage
  for _,cu := range(gex.ConcUsage.ConcUsage){
    switch cu.Type{
    case instrument.SEND, instrument.RECV:
    }
  }
}*/
