package goatDefSel

import "testing"

func TestGoatDefSel(t *testing.T) {
	c := make(chan int)
	go func() {
		c <- 0
		c <- 0
	}()
	it:for{
		select {
		default:
		case <- c:
			break it
		}
	}
	<- c
}
