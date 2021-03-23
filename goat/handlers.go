package goat

import (
	"runtime"
	"time"
	"runtime/trace"
	"os"
	"fmt"
	"strconv"
	"log"
)

func Sched_Handler(){
	randBound_yield()
  // other handlers can come down here
}

func Start(maxprocs int) chan interface{}{
  fmt.Println("GOAT start...")
  runtime.GOMAXPROCS(maxprocs)
  trace.Start(os.Stderr)
  ch := make(chan interface{})
  return ch
}

func Watch(ch chan interface{}){
	fmt.Println("GOAT stop...")
	to, err := strconv.Atoi(os.Getenv("GOATTO"))
	if err != nil{
		panic("GOATTO not set")
	}
  select {
  case <- ch:
		fmt.Println("GOAT finished (normal)")
		ch <- 0

  case <- time.After(time.Second * time.Duration(to)):
    trace.Stop()
    fmt.Println("GOAT stopped (timeout)")
    os.Exit(0)
  }
}


func Stop(ch chan interface{}){
	if err := recover() ; err != nil{
		// an error occured
		//time.Sleep(time.Millisecond)
		trace.Stop()
		log.Println(err)
	}
	ch <- true
	<-ch
	time.Sleep(time.Millisecond)
	trace.Stop()
}
