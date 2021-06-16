package goatSafeInc

import (
	"sync"
	"time"
  "testing"
)

// SafeCounter is safe to use concurrently.
type SafeCounter struct {
	mu sync.Mutex
	v  map[string]int
}

// Inc increments the counter for the given key.
func (c *SafeCounter) Inc(key string) {
	c.mu.Lock()
	c.v[key]++
	c.mu.Unlock()
}

// Value returns the current value of the counter for the given key.
func (c *SafeCounter) Value(key string) int {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.mu.Unlock()
	return c.v[key]
}

func TestGoatSafeInc(t *testing.T) {
	c := SafeCounter{v: make(map[string]int)}
	for i := 0; i < 10; i++ {
		go c.Inc("somekey")
	}


  c.Value("somekey")
  time.Sleep(time.Second)
}
