package evaluate
// Event types in the trace.
// Verbatim copy from src/runtime/trace.go with the "trace" prefix removed.
const (
	EvNone              = 0  // unused
	EvBatch             = 1  // start of per-P batch of events [pid, timestamp]
	EvFrequency         = 2  // contains tracer timer frequency [frequency (ticks per second)]
	EvStack             = 3  // stack [stack id, number of PCs, array of {PC, func string ID, file string ID, line}]
	EvGomaxprocs        = 4  // current value of GOMAXPROCS [timestamp, GOMAXPROCS, stack id]
	EvProcStart         = 5  // start of P [timestamp, thread id]
	EvProcStop          = 6  // stop of P [timestamp]
	EvGCStart           = 7  // GC start [timestamp, seq, stack id]
	EvGCDone            = 8  // GC done [timestamp]
	EvGCSTWStart        = 9  // GC mark termination start [timestamp, kind]
	EvGCSTWDone         = 10 // GC mark termination done [timestamp]
	EvGCSweepStart      = 11 // GC sweep start [timestamp, stack id]
	EvGCSweepDone       = 12 // GC sweep done [timestamp, swept, reclaimed]
	EvGoCreate          = 13 // goroutine creation [timestamp, new goroutine id, new stack id, stack id]
	EvGoStart           = 14 // goroutine starts running [timestamp, goroutine id, seq]
	EvGoEnd             = 15 // goroutine ends [timestamp]
	EvGoStop            = 16 // goroutine stops (like in select{}) [timestamp, stack]
	EvGoSched           = 17 // goroutine calls Gosched [timestamp, stack]
	EvGoPreempt         = 18 // goroutine is preempted [timestamp, stack]
	EvGoSleep           = 19 // goroutine calls Sleep [timestamp, stack]
	EvGoBlock           = 20 // goroutine blocks [timestamp, stack]
	EvGoUnblock         = 21 // goroutine is unblocked [timestamp, goroutine id, seq, stack]
	EvGoBlockSend       = 22 // goroutine blocks on chan send [timestamp, stack]
	EvGoBlockRecv       = 23 // goroutine blocks on chan recv [timestamp, stack]
	EvGoBlockSelect     = 24 // goroutine blocks on select [timestamp, stack]
	EvGoBlockSync       = 25 // goroutine blocks on Mutex/RWMutex [timestamp, stack]
	EvGoBlockCond       = 26 // goroutine blocks on Cond [timestamp, stack]
	EvGoBlockNet        = 27 // goroutine blocks on network [timestamp, stack]
	EvGoSysCall         = 28 // syscall enter [timestamp, stack]
	EvGoSysExit         = 29 // syscall exit [timestamp, goroutine id, seq, real timestamp]
	EvGoSysBlock        = 30 // syscall blocks [timestamp]
	EvGoWaiting         = 31 // denotes that goroutine is blocked when tracing starts [timestamp, goroutine id]
	EvGoInSyscall       = 32 // denotes that goroutine is in syscall when tracing starts [timestamp, goroutine id]
	EvHeapAlloc         = 33 // memstats.heap_live change [timestamp, heap_alloc]
	EvNextGC            = 34 // memstats.next_gc change [timestamp, next_gc]
	EvTimerGoroutine    = 35 // denotes timer goroutine [timer goroutine id]
	EvFutileWakeup      = 36 // denotes that the previous wakeup of this goroutine was futile [timestamp]
	EvString            = 37 // string dictionary entry [ID, length, string]
	EvGoStartLocal      = 38 // goroutine starts running on the same P as the last event [timestamp, goroutine id]
	EvGoUnblockLocal    = 39 // goroutine is unblocked on the same P as the last event [timestamp, goroutine id, stack]
	EvGoSysExitLocal    = 40 // syscall exit on the same P as the last event [timestamp, goroutine id, real timestamp]
	EvGoStartLabel      = 41 // goroutine starts running with label [timestamp, goroutine id, seq, label string id]
	EvGoBlockGC         = 42 // goroutine blocks on GC assist [timestamp, stack]
	EvGCMarkAssistStart = 43 // GC mark assist start [timestamp, stack]
	EvGCMarkAssistDone  = 44 // GC mark assist done [timestamp]
	EvUserTaskCreate    = 45 // trace.NewContext [timestamp, internal task id, internal parent id, stack, name string]
	EvUserTaskEnd       = 46 // end of task [timestamp, internal task id, stack]
	EvUserRegion        = 47 // trace.WithRegion [timestamp, internal task id, mode(0:start, 1:end), stack, name string]
	EvUserLog           = 48 // trace.Log [timestamp, internal id, key string id, stack, value string]
	EvChSend            = 49 // GOAT: chan send [timestamp, stack, channel id, ch_event id, value, pos]
	EvChRecv            = 50 // GOAT: chan recv [timestamp, stack, channel id, ch_event id, value, pos]
	EvChMake            = 51 // GOAT: chan make [timestamp, stack, channel id]
	EvChClose           = 52 // GOAT: chan close [timestamp, stack, channel id]
	EvWgAdd             = 53 // GOAT: wg add (and inited) [timestamp, stack, wg id, value]
	EvWgWait            = 54 // GOAT: wg wait [timestamp, stack, wg id, pos]
	EvMuLock            = 55 // GOAT: mu lock [timestamp, stack, mu id, pos]
	EvMuUnlock          = 56 // GOAT: mu unlock [timestamp, stack, mu id]
	EvSelect            = 57 // GOAT: select [timestamp, stack, pos]
	EvSched             = 58 // GOAT: sched [timestamp, stack, pos, curg, aux]
	EvCvWait            = 59 // GOAT: cond var wait [timestamp, stack, cv id]
  EvCvSig             = 60 // GOAT: cond var signal [timestamp, stack, cv id, {1: signal, 2: broadcast}]
	EvCount             = 61
)

var EventDescriptions = [EvCount]struct {
	Name       string
	minVersion int
	Stack      bool
	Args       []string
	SArgs      []string // string arguments
}{
	EvNone:              {"None", 1005, false, []string{}, nil},
	EvBatch:             {"Batch", 1005, false, []string{"p", "ticks"}, nil}, // in 1.5 format it was {"p", "seq", "ticks"}
	EvFrequency:         {"Frequency", 1005, false, []string{"freq"}, nil},   // in 1.5 format it was {"freq", "unused"}
	EvStack:             {"Stack", 1005, false, []string{"id", "siz"}, nil},
	EvGomaxprocs:        {"Gomaxprocs", 1005, true, []string{"procs"}, nil},
	EvProcStart:         {"ProcStart", 1005, false, []string{"thread"}, nil},
	EvProcStop:          {"ProcStop", 1005, false, []string{}, nil},
	EvGCStart:           {"GCStart", 1005, true, []string{"seq"}, nil}, // in 1.5 format it was {}
	EvGCDone:            {"GCDone", 1005, false, []string{}, nil},
	EvGCSTWStart:        {"GCSTWStart", 1005, false, []string{"kindid"}, []string{"kind"}}, // <= 1.9, args was {} (implicitly {0})
	EvGCSTWDone:         {"GCSTWDone", 1005, false, []string{}, nil},
	EvGCSweepStart:      {"GCSweepStart", 1005, true, []string{}, nil},
	EvGCSweepDone:       {"GCSweepDone", 1005, false, []string{"swept", "reclaimed"}, nil}, // before 1.9, format was {}
	EvGoCreate:          {"GoCreate", 1005, true, []string{"g", "stack"}, nil},
	EvGoStart:           {"GoStart", 1005, false, []string{"g", "seq"}, nil}, // in 1.5 format it was {"g"}
	EvGoEnd:             {"GoEnd", 1005, false, []string{}, nil},
	EvGoStop:            {"GoStop", 1005, true, []string{}, nil},
	EvGoSched:           {"GoSched", 1005, true, []string{}, nil},
	EvGoPreempt:         {"GoPreempt", 1005, true, []string{}, nil},
	EvGoSleep:           {"GoSleep", 1005, true, []string{}, nil},
	EvGoBlock:           {"GoBlock", 1005, true, []string{}, nil},
	EvGoUnblock:         {"GoUnblock", 1005, true, []string{"g", "seq"}, nil}, // in 1.5 format it was {"g"}
	EvGoBlockSend:       {"GoBlockSend", 1005, true, []string{}, nil},
	EvGoBlockRecv:       {"GoBlockRecv", 1005, true, []string{}, nil},
	EvGoBlockSelect:     {"GoBlockSelect", 1005, true, []string{}, nil},
	EvGoBlockSync:       {"GoBlockSync", 1005, true, []string{}, nil},
	EvGoBlockCond:       {"GoBlockCond", 1005, true, []string{}, nil},
	EvGoBlockNet:        {"GoBlockNet", 1005, true, []string{}, nil},
	EvGoSysCall:         {"GoSysCall", 1005, true, []string{}, nil},
	EvGoSysExit:         {"GoSysExit", 1005, false, []string{"g", "seq", "ts"}, nil},
	EvGoSysBlock:        {"GoSysBlock", 1005, false, []string{}, nil},
	EvGoWaiting:         {"GoWaiting", 1005, false, []string{"g"}, nil},
	EvGoInSyscall:       {"GoInSyscall", 1005, false, []string{"g"}, nil},
	EvHeapAlloc:         {"HeapAlloc", 1005, false, []string{"mem"}, nil},
	EvNextGC:            {"NextGC", 1005, false, []string{"mem"}, nil},
	EvTimerGoroutine:    {"TimerGoroutine", 1005, false, []string{"g"}, nil}, // in 1.5 format it was {"g", "unused"}
	EvFutileWakeup:      {"FutileWakeup", 1005, false, []string{}, nil},
	EvString:            {"String", 1007, false, []string{}, nil},
	EvGoStartLocal:      {"GoStartLocal", 1007, false, []string{"g"}, nil},
	EvGoUnblockLocal:    {"GoUnblockLocal", 1007, true, []string{"g"}, nil},
	EvGoSysExitLocal:    {"GoSysExitLocal", 1007, false, []string{"g", "ts"}, nil},
	EvGoStartLabel:      {"GoStartLabel", 1008, false, []string{"g", "seq", "labelid"}, []string{"label"}},
	EvGoBlockGC:         {"GoBlockGC", 1008, true, []string{}, nil},
	EvGCMarkAssistStart: {"GCMarkAssistStart", 1009, true, []string{}, nil},
	EvGCMarkAssistDone:  {"GCMarkAssistDone", 1009, false, []string{}, nil},
	EvUserTaskCreate:    {"UserTaskCreate", 1011, true, []string{"taskid", "pid", "typeid"}, []string{"name"}},
	EvUserTaskEnd:       {"UserTaskEnd", 1011, true, []string{"taskid"}, nil},
	EvUserRegion:        {"UserRegion", 1011, true, []string{"taskid", "mode", "typeid"}, []string{"name"}},
	EvUserLog:           {"UserLog", 1011, true, []string{"id", "keyid"}, []string{"category", "message"}},
	EvChSend:            {"ChSend", 1011, true, []string{"cid","chid","val","pos"},nil}, // GOAT: chan send [timestamp, stack, channel id, ch_event id, value, pos]
	EvChRecv:            {"ChRecv", 1011, true, []string{"cid","chid","val","pos"},nil}, // GOAT: chan send [timestamp, stack, channel id, ch_event id, value, pos]
	EvChMake:            {"ChMake", 1011, true, []string{"cid"},nil},// GOAT: chan make [timestamp, stack, channel id]
	EvChClose:           {"ChClose", 1011, true, []string{"cid"},nil},// GOAT: chan close [timestamp, stack, channel id]
	EvWgAdd:             {"WgAdd", 1011, true, []string{"wid","val"},nil}, // GOAT: wg add (and inited) [timestamp, stack, wg id, value]
	EvWgWait:            {"WgWait", 1011, true, []string{"wid","pos"},nil}, // GOAT: wg wait [timestamp, stack, wg id]
	EvMuLock:            {"MuLock", 1011, true, []string{"muid","pos"},nil},// GOAT: mu lock [timestamp, stack, mu id]
	EvMuUnlock:          {"MuUnlock", 1011, true, []string{"muid"},nil},// GOAT: mu unlock [timestamp, stack, mu id]
	EvSelect:            {"Select", 1011, true, []string{"pos"},nil},// GOAT: select [timestamp, stack, pos]
	EvSched:             {"Sched", 1011, true, []string{"pos","curg","aux"},nil}, // GOAT: sched [timestamp, stack, pos, curg, aux]
	EvCvWait:            {"CvWait",1011, true, []string{"cvid"},nil}, // GOAT: cond var wait [timestamp, stack, cv id]
  EvCvSig:             {"CvSig",1011, true, []string{"cvid","pos"},nil}, // GOAT: cond var signal [timestamp, stack, cv id, {1: signal, 2: broadcast}]
}

var ctgDescriptions = [10]struct {
	Category      string
	Members      []string
}{
	catGRTN :  {"GRTN",[]string{"EvGoCreate","EvGoStart","EvGoEnd","EvGoStop","EvGoSched","EvGoPreempt","EvGoSleep","EvGoBlock","EvGoUnblock","EvGoBlockSend","EvGoBlockRecv","EvGoBlockSelect","EvGoBlockSync","EvGoBlockCond","EvGoBlockNet","EvGoWaiting","EvGoInSyscall","EvGoStartLocal","EvGoUnblockLocal","EvGoSysExitLocal","EvGoStartLabel","EvGoBlockGC"}},
	catBLCK :  {"BLCK",[]string{"EvGoStart","EvGoEnd","EvGoStop","EvGoSched","EvGoPreempt","EvGoSleep","EvGoBlock","EvGoUnblock","EvGoBlockSend","EvGoBlockRecv","EvGoBlockSelect","EvGoBlockSync","EvGoBlockCond","EvGoBlockNet","EvGoUnblockLocal","EvGoBlockGC"}},
  catCHNL :  {"CHNL",[]string{"EvChSend","EvChRecv","EvChMake","EvChClose"}},
	catWGCV :  {"WGCV",[]string{"EvWgAdd","EvWgWait","EvCvWait","EvCvSig"}},
	catMUTX :  {"MUTX",[]string{"EvMuLock","EvMuUnlock"}},
  catPROC :  {"PROC",[]string{"EvNone","EvBatch","EvFrequency","EvStack","EvGomaxprocs","EvProcStart","EvProcStop"}},
  catGCMM :  {"GCMM",[]string{"EvGCStart","EvGCDone","EvGCSTWStart","EvGCSTWDone","EvGCSweepStart","EvGCSweepDone","EvHeapAlloc","EvNextGC","EvGCMarkAssistStart","EvGCMarkAssistDone"}},
  catSYSC :  {"SYSC",[]string{"EvGoSysCall","EvGoSysExit","EvGoSysBlock"}},
  catMISC :  {"MISC",[]string{"EvUserTaskCreate","EvUserTaskEnd","EvUserRegion","EvUserLog","EvTimerGoroutine","EvFutileWakeup","EvString"}},
	catSCHD :  {"SCHD",[]string{"EvSelect","EvWgAdd","EvChSend","EvWgWait","EvChClose","EvChRecv","EvMuLock","EvMuUnlock","EvCvWait","EvCvSig"}},
}


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
