package main

import (
  "sync"
  "fmt"
)

func main() {
  ch := make(chan int)
  m := sync.Mutex{}
	cv := sync.NewCond(&m)
  var m1 sync.Mutex

  // goroutine 1
  go func() {
		//time.Sleep(5*time.Millisecond) // simulates computation
    sleep(5)
    m1.Lock()
    cv.L.Lock()
    cv.Signal()
    cv.L.Unlock()
    m1.Unlock()
  }()

  // goroutine 2
  go func() {
    cv.L.Lock()
    cv.Wait()
    cv.L.Unlock()
    close(ch)
  }()

  // goroutine 3
  go func(){
		//time.Sleep(5*time.Millisecond) // simulates computation
    sleep(5)
    m1.Lock()
    <-ch
    x := <-ch
    m1.Unlock()
    ch <- x
  }()

  //go new()
  fmt.Println("End of main!")
}

//
// func new(){
//   time.Sleep(5*time.Millisecond) // simulates computation
//   select {
// 	case <-eventch:
//     <- eventch
//     eventch <- 2
// 	case eventch <- 2:
// 		cv.L.Lock()
// 		cv.x.c.x.c.x.c.Wait()
// 		cv.L.Unlock()
// 		close(done)
// 	}
//
//   select {
// 	case <-eventch:
//     <- eventch
//     eventch <- 2
// 	default:
//     KIR
// 	}
//
//   for elem := range queue {
//     fmt.Println(elem)
//   }
//
//   for _,s := range(l){
//     fmt.Println(elem)
//   }
//
//   for i := 0 ; i < 10 ; i++{
//     fmt.Println(i)
//   }
//
// }
