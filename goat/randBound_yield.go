package goat

import (
	"strconv"
	"math/rand"
	"runtime"
	"time"
	"os"
	"sync"
)

type SharedCounter struct {
	cnt    int
	sync.Mutex
}

var bound SharedCounter

func randBound_yield(){
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(2) == 1 {
		bound.Lock()
		defer bound.Unlock()
		if thr, err := strconv.Atoi(os.Getenv("GOATRSBOUND")); err == nil && bound.cnt < thr {
			bound.cnt++
      bound.Unlock()
      runtime.Gosched()
		}
	}
}
