package main

import (
  "time"
  "sync"
  "fmt"
)

// https://github.com/kubernetes/kubernetes/pull/16223
// Buggy version: e755988d5922df4d0e111a0167d9859359113463
// https://github.com/kubernetes/kubernetes/pull/10182
// buggy version: 4b990d128a17eea9058d28a3b3688ab8abafbd94

// Buggy scenario
// G1                G2               G3
// -----------------------------------------------
// blockRecv
//                 lock
//   (unblocks <-) sends
//                 unlock
//                                  lock
//                                  send // block
// lock //block

func main() {
  ch1 := make(chan int)
  stop := make(chan int)
  var m sync.Mutex

  // goroutine 1
  go func(){
    for {
      select{
      case <- stop: // terminate
        <- stop
        return
      case stop <- 0:
        return 
      default:
      }
      <- ch1
      m.Lock()
      m.Unlock()
      <- ch1
    }
  }()

  // goroutine 2
  go func() {
    m.Lock()
    ch1 <- 1
    m.Unlock()
    stop <- 1
  }()

  // goroutine 3
  go func() {
    m.Lock()
    ch1 <- 1
    m.Unlock()
    stop <- 1
  }()

  time.Sleep(time.Second)
  fmt.Println("End of main!")
}
