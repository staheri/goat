package goat

import (
	"runtime"
	"time"
	"runtime/trace"
	"os"
	"fmt"
	"strconv"
)

func Sched_Handler(){
	randBound_yield()
  // other handlers can come down here
}

func Start() chan interface{}{
  fmt.Println("GOAT start...")
  runtime.GOMAXPROCS(1)
  trace.Start(os.Stderr)
  ch := make(chan interface{})
  return ch
}

func Stop(ch chan interface{}){
	fmt.Println("GOAT stop...")
	to, err := strconv.Atoi(os.Getenv("GOATTO"))
	if err != nil{
		panic("GOATTO not set")
	}
  select {
  case <- ch:
    trace.Stop()
		fmt.Println("GOAT finished (normal)")
		ch <- 0

  case <- time.After(time.Second * time.Duration(to)):
    trace.Stop()
    fmt.Println("GOAT stopped (timeout)")
    os.Exit(0)
  }
}
