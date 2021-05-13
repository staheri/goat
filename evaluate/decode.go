package evaluate



const (
	catGRTN         = iota  // Goroutine events
	catCHNL                 // Channel events
	catWGCV                 // WaitingGroup & Conditional Variable events
	catMUTX                 // Mutex events
	catPROC                 // Process events
	catGCMM                 // Garbage collection/memory events
	catSYSC                 // Syscall events
	catMISC                 // Other events
	catBLCK                 // Blocking events
	catSCHD                 // test-sched events
	catCNT                  // cat Count
)
