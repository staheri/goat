// detectors
package evaluate

import(
  "strings"
  _"fmt"
)

// Detector function for BuiltinDL
func builtinDL_detector(out []byte) (bool,string) {
  if strings.Contains(string(out), "asleep"){
    return true,"GDL"
  }
  return false,""
}

// Detector function for goleak
func goleak_detector(out []byte) (bool,string) {
  if strings.Contains(string(out), "found unexpected goroutines"){
    return true,"PDL"
  } else if strings.Contains(string(out), "asleep"){
    return true, "GDL - from builtin"
  } else if strings.Contains(string(out), "timed out"){
    return true,"TO/GDL"
  }
  return false,""
}

// Detector function for LockDL
func lockDL_detector(out []byte) (bool,string) {
  if strings.Contains(string(out), "POTENTIAL DEADLOCK:"){
    return true,"DL"
  } else if strings.Contains(string(out), "timed out"){
    return true,"TO/GDL"
  }
  return false,""
}


// Detector function for Race
func race_detector(out []byte) (bool,string) {

  //fmt.Printf("************ OUT *************\n%v\n******************************\n",string(out))
  msg := ""
  if strings.Contains(string(out), "panic:"){
    msg = msg + "PANIC("
    if strings.Contains(string(out), "send on closed channel"){
      msg = msg + "S.O.C)"
    } else if strings.Contains(string(out), "runtime error: invalid memory" ){
      msg = msg + "RT-Nil.Mem)"
    } else if strings.Contains(string(out), "sync: negative " ){
      msg = msg + "Neg.Wg.Cnt)"
    }else{
      msg = msg + "X)"
    }
  } else if strings.Contains(string(out), "send on closed channel"){
    msg = "PANIC(S.O.C)"
  }
  if strings.Contains(string(out), "WARNING: DATA RACE"){
    if msg != ""{
      return true,msg+"/RACE"
    }
    return true,"RACE"
  } else{
    if msg != ""{
      return true,msg
    }
  }
  return false,""
}
