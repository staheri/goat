package main

import (
	"time"
)

func sleep(duration time.Duration){
  time.Sleep(duration*time.Millisecond)
}
