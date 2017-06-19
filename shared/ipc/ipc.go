package ipc

import "regexp"

// Version used with Version command of lutractl
// Returns server version to the client
type Version struct {
	ServerVersion string
	ServerBuildTime string
	ServerBuildHash string
}

// SysStatus used with Stats command of lutractl
type SysStatus struct {
	Uptime       string
	NumGoroutine int

	// General statistics.
	MemAllocated string // bytes allocated and still in use
	MemTotal     string // bytes allocated (even if freed)
	MemSys       string // bytes obtained from system (sum of XxxSys below)
	Lookups      uint64 // number of pointer lookups
	MemMallocs   uint64 // number of mallocs
	MemFrees     uint64 // number of frees

	// Main allocation heap statistics.
	HeapAlloc    string // bytes allocated and still in use
	HeapSys      string // bytes obtained from system
	HeapIdle     string // bytes in idle spans
	HeapInuse    string // bytes in non-idle span
	HeapReleased string // bytes released to the OS
	HeapObjects  uint64 // total number of allocated objects

	// Low-level fixed-size structure allocator statistics.
	//  Inuse is bytes used now.
	//  Sys is bytes obtained from system.
	StackInuse  string // bootstrap stacks
	StackSys    string
	MSpanInuse  string // mspan structures
	MSpanSys    string
	MCacheInuse string // mcache structures
	MCacheSys   string
	BuckHashSys string // profiling bucket hash table
	GCSys       string // GC metadata
	OtherSys    string // other system allocations

	// Garbage collector statistics.
	NextGC       string // next run in HeapAlloc time (bytes)
	LastGC       string // last run in absolute time (ns)
	PauseTotalNs string
	PauseNs      string // circular buffer of recent GC pause times, most recent at [(NumGC+255)%256]
	NumGC        uint32
}

// AskStatus struct with limited service name or asking for all
type AskStatus struct {
	Name		string
	All			bool
}

// ServiceName specify service name
type ServiceName string
// RunState represents one of the states the service can be
type RunState uint8

// LoadedService is a lightweight Service
type LoadedService struct {
	Name			ServiceName
	Description		string		// Currently not used
	State			RunState

	LastAction		LastAction
	LastActionAt	int64		// Timestamp of the last action (UTC)
	LastMessage		string
	LastKnownPID	int

	Type			string // forking or simple
	PIDFile			string
}

// Types of valid runState
const (
	NotStarted = RunState(iota)
	Starting
	Started
	Stopped
	Errored
)

func (rs RunState) String() string {
	switch rs {
	case NotStarted:
		return "not started"
	case Starting:
		return "being started"
	case Started:
		return "already started"
	case Stopped:
		return "stopped"
	case Errored:
		return "errored"
	default:
		return "in an invalid state"
	}
}

// LastAction represent the latest action done to the service
type LastAction uint8

// Last actions constants
const (
	Unknown = LastAction(iota)
	Start
	Stop
	Reload
	Restart
	Forcekill
)

func (la LastAction) String() string {
	switch la {
	case Unknown:
		return "unknown"
	case Start:
		return "start"
	case Stop:
		return "stop"
	case Reload:
		return "reload"
	case Restart:
		return "restart"
	case Forcekill:
		return "force kill"
	default:
		return "really unknown"
	}
}

// IsCustASCII is a custom regexp checker for sanity
var IsCustASCII = regexp.MustCompile(`^[a-zA-Z0-9_\-.]+$`).MatchString

// IsCustASCIISpace is a custom regexp checker for sanity with a space !!!
var IsCustASCIISpace = regexp.MustCompile(`^[a-zA-Z0-9_\-. ]+$`).MatchString
