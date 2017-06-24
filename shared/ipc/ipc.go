package ipc

import (
	"regexp"
)

// WARNING: Huges parts of this file should be synced with content of lutrainit/lutrainit/service.go

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

// AnswerReload is a reload answer
type AnswerReload struct {
	Err			bool
	ErrStr		string
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

// Command defines a command string used by Startup or Shutdown
type Command string

func (c Command) String() string {
	return string(c)
}

// ServiceName defines the service name
type ServiceName string

// ServiceType provides or needs
type ServiceType string

// RunState is running state
type RunState uint8

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


// ServiceAction for start/stop/restart
type ServiceAction struct {
	Name		string
	Action		LastAction
}

// ServiceActionAnswer is a service action answer
type ServiceActionAnswer struct {
	Name		string
	Action		LastAction
	Err			bool
	ErrStr		string
}

// Service represents a struct with usefull infos used for management of services
type Service struct {
	Name		ServiceName
	AutoStart	bool

	Provides 	[]ServiceType
	Needs    	[]ServiceType

	Description		string		// Currently not used
	State			RunState

	LastAction		LastAction
	LastActionAt	int64		// Timestamp of the last action (UTC)
	LastMessage		string
	LastKnownPID	int

	Type			string // forking or simple
	PIDFile			string

	Startup  	Command
	Shutdown 	Command
	CheckAlive  Command

	Deleted			bool
}

// IsCustASCII is a custom regexp checker for sanity
var IsCustASCII = regexp.MustCompile(`^[a-zA-Z0-9_\-.]+$`).MatchString

// IsCustASCIISpace is a custom regexp checker for sanity with a space !!!
var IsCustASCIISpace = regexp.MustCompile(`^[a-zA-Z0-9_\-. ]+$`).MatchString
