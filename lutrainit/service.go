package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
	"github.com/go-clog/clog"
	"io/ioutil"
	"strconv"
	"bytes"
	"github.com/mitchellh/go-ps"
	"syscall"
)

const (
	errWaitNoChild   = "wait: no child processes"
	errWaitIDNoChild = "waitid: no child processes"
)

// ServiceName defines the service name
type ServiceName string

// ServiceType provides or needs
type ServiceType string

// RunState define running state of service
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

// LastAction represent the latest action done to the service
type LastAction uint8

// Last actions constants
const (
	Unknown = LastAction(iota)
	PreStart
	Start
	PostStart
	PreStop
	Stop
	PostStop
	Reload
	Restart
	Forcekill
)

func (la LastAction) String() string {
	switch la {
	case Unknown:
		return "unknown"
	case PreStart:
		return "pre start"
	case Start:
		return "start"
	case PostStart:
		return "post start"
	case PreStop:
		return "pre stop"
	case Stop:
		return "stop"
	case PostStop:
		return "post stop"
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

// StartupService is a lightweight service used only for startup loop
type StartupService struct {
	Name		ServiceName
	AutoStart	bool

	Provides 	[]ServiceType
	Needs    	[]ServiceType
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

	ExecPreStart	Command
	ExecPreStop		Command
	ExecPostStart	Command
	ExecPostStop	Command

	Deleted			bool
	Filename		string
}

// StartServices starts all declared services at start
func StartServices() {
	wg := sync.WaitGroup{}

	var startedMu = &sync.RWMutex{}
	startedTypes := make(map[ServiceType]bool)
	for _, services := range StartupServices {
		wg.Add(len(services))
		for _, s := range services {
			lS := LoadedServices[s.Name]
			go func(s *Service) {
				// TODO: This should ensure that Needs are satisfiable instead of getting into an
				// infinite loop when they're not.
				// (TODO(2): Prove N=NP in order to do the above efficiently.)
				for satisfied, tries := false, 0; satisfied == false && tries < 60; tries++ {
					satisfied = s.NeedsSatisfied(startedTypes, startedMu)
					time.Sleep(2 * time.Second)

				}
				if s.State == NotStarted && s.AutoStart {
					// Start the service
					if s.Type == "oneshot" || s.Type == "forking" {
						if err := s.Start(); err != nil {
							clog.Error(2, err.Error())
						}
					} else if s.Type == "simple" {
						go s.StartSimple()
					} else {
						// What are you doing here ?
						clog.Warn("I don't know why but I'm asked to start %s with type %s\n", s.Name, s.Type)
					}
				}

				startedMu.Lock()
				for _, t := range s.Provides {
					startedTypes[t] = true
				}
				startedMu.Unlock()
				wg.Done()
			}(lS)
		}
	}
	wg.Wait()
}

// Start the Service s. if type is oneshot or forking
func(s Service) Start() error {
	LoadedServicesMu.Lock()
	defer LoadedServicesMu.Unlock()

	if s.State != NotStarted {
		return fmt.Errorf("Service %v is %v", s.Name, s.State.String())
	}
	s.State = Starting
	LoadedServices[s.Name].State = Starting
	LoadedServices[s.Name].LastActionAt = time.Now().UTC().Unix()
	LoadedServices[s.Name].LastAction = Start

	if s.ExecPreStart != "" {
		LoadedServices[s.Name].LastActionAt = time.Now().UTC().Unix()
		LoadedServices[s.Name].LastAction = PreStart

		err := justExecACommand(s.ExecPreStart.String())
		if err != nil {
			clog.Error(2, "error in %s ExecPreStart: %s", s.Name, err.Error())
			return err
		}
	}

	cmd := exec.Command("sh", "-c", s.Startup.String())
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Run(); err != nil {
		if err.Error() == errWaitNoChild || err.Error() == errWaitIDNoChild {
			// Process exited cleanly
			s.State = Started
			LoadedServices[s.Name].State = Started
			LoadedServices[s.Name].LastActionAt = time.Now().UTC().Unix()
			LoadedServices[s.Name].LastAction = Start

			clog.Info("[lutra] Started service %s", s.Name)

			return nil
		}

		s.State = Errored
		LoadedServices[s.Name].State = Errored
		LoadedServices[s.Name].LastActionAt = time.Now().UTC().Unix()
		LoadedServices[s.Name].LastAction = Start

		clog.Error(2,"[lutra] Error starting service %s: %s", s.Name, err.Error())

		return err
	}

	if s.ExecPostStart != "" {
		LoadedServices[s.Name].LastActionAt = time.Now().UTC().Unix()
		LoadedServices[s.Name].LastAction = PostStart

		err := justExecACommand(s.ExecPostStart.String())
		if err != nil {
			clog.Error(2, "error in %s ExecPostStart: %s", s.Name, err.Error())
			return err
		}
	}

	s.State = Started
	LoadedServices[s.Name].State = Started
	LoadedServices[s.Name].LastActionAt = time.Now().UTC().Unix()
	LoadedServices[s.Name].LastAction = Start

	clog.Info("[lutra] Started service %s", s.Name)

	return nil
}

// StartSimple and track the PID and process state (for simple service without auto-fork function)
// Please remember that this function locks in the middle (cmd.Wait()) for any mutex operation
func(s Service) StartSimple() {
	LoadedServicesMu.Lock()
	s.State = Starting
	LoadedServices[s.Name].State = Starting
	LoadedServices[s.Name].LastActionAt = time.Now().UTC().Unix()
	LoadedServices[s.Name].LastAction = Start
	LoadedServices[s.Name].LastKnownPID = 0
	LoadedServicesMu.Unlock()

	if s.ExecPreStart != "" {
		LoadedServicesMu.Lock()
		LoadedServices[s.Name].LastActionAt = time.Now().UTC().Unix()
		LoadedServices[s.Name].LastAction = PreStart
		LoadedServicesMu.Unlock()

		err := justExecACommand(s.ExecPreStart.String())
		if err != nil {
			clog.Error(2, "error in %s ExecPreStart: %s", s.Name, err.Error())
			return
		}
	}

	cmd := exec.Command("sh", "-c", s.Startup.String())
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		clog.Error(2,"[lutra] Service %s exited with error: %s", s.Name, err.Error())
		LoadedServicesMu.Lock()
		s.State = Errored
		LoadedServices[s.Name].State = Errored
		LoadedServices[s.Name].LastMessage = err.Error()
		LoadedServices[s.Name].LastKnownPID = 0
		LoadedServicesMu.Unlock()
		return
	}
	// Waiting for the command to finish
	LoadedServicesMu.Lock()
	s.State = Started
	LoadedServices[s.Name].State = Started
	LoadedServices[s.Name].LastKnownPID = cmd.Process.Pid
	LoadedServicesMu.Unlock()
	clog.Info("[lutra] Started service %s", s.Name)

	err := cmd.Wait()
	if err != nil {
		clog.Error(2, "[lutra] Service %s finished with error: %s", s.Name, err.Error())
		LoadedServicesMu.Lock()
		s.State = Stopped
		LoadedServices[s.Name].State = Stopped
		LoadedServices[s.Name].LastMessage = err.Error()
		LoadedServices[s.Name].LastKnownPID = 0
		LoadedServicesMu.Unlock()
	} else {
		LoadedServicesMu.Lock()
		s.State = Stopped
		clog.Info("[lutra] Service stopped:	 %s", s.Name)
		LoadedServices[s.Name].State = Stopped
		LoadedServices[s.Name].LastKnownPID = 0
		LoadedServices[s.Name].LastActionAt = time.Now().UTC().Unix()
		LoadedServices[s.Name].LastAction = Stop
		LoadedServicesMu.Unlock()
	}

	if s.ExecPostStart != "" {
		LoadedServicesMu.Lock()
		LoadedServices[s.Name].LastActionAt = time.Now().UTC().Unix()
		LoadedServices[s.Name].LastAction = PostStart
		LoadedServicesMu.Unlock()

		err := justExecACommand(s.ExecPostStart.String())
		if err != nil {
			clog.Error(2, "error in %s ExecPostStart: %s", s.Name, err.Error())
			return
		}
	}

}

// NeedsSatisfied if all of s's needs are satified by the passed list of provided types
func (s Service) NeedsSatisfied(started map[ServiceType]bool, mu *sync.RWMutex) bool {
	mu.RLock()
	defer mu.RUnlock()
	for _, st := range s.Needs {
		if !started[st] {
			return false
		}
	}
	return true
}

func getProcessPid(s *Service) (pid int, err error) {
	d, err := ioutil.ReadFile(s.PIDFile)
	if err != nil {
		return 0, err
	}

	pid, err = strconv.Atoi(string(bytes.TrimSpace(d)))
	if err != nil {
		return 0, fmt.Errorf("error parsing pid from: %s", s.PIDFile)
	}

	return pid, nil
}

func processAliveByPid(pid int) (alive bool, err error) {
	if pid == 0 {
		return false, fmt.Errorf("why are you asking me if PID 0 is alive ?")
	}

	_, err = ps.FindProcess(pid)
	if err != nil {
		return false, err
	}

	return true, nil
}

// returns true if command successfull, else always false
func processAliveByCmd(command string) (alive bool, err error) {
	cmd := exec.Command("sh", "-c", command)

	if err = cmd.Run(); err != nil {
		return false, err
	}
	// did the command fail because of an unsuccessful exit code
	if _, ok := err.(*exec.ExitError); ok {
		return false, nil
	}

	return true, nil
}

// justExecACommand of the specified service
func justExecACommand(command string) (err error) {
	cmd := exec.Command("sh", "-c", command)

	if err = cmd.Run(); err != nil {
		return err
	}
	// did the command fail because of an unsuccessful exit code
	if _, ok := err.(*exec.ExitError); ok {
		return nil
	}

	return nil
}

func checkIfProcessAlive(s *Service) (alive bool, pid int, err error) {
	// Check using PID
	if s.PIDFile != "" {
		if _, err := os.Stat(s.PIDFile); os.IsNotExist(err) {
			return false, 0, nil // NO PID
		}

		pid, err := getProcessPid(s)
		if err != nil {
			return false, 0, err
		}
		running, err := processAliveByPid(pid)
		if err != nil {
			return false, 0, err
		}
		return running, pid, nil
	}

	// TODO: some sort of 'pgrep blah' fork forking types

	// Else if it's a simple, check status from list
	if s.Type == "simple" {
		return s.State == Started, 0, nil
	}

	// Cannot determine process state
	return true, 0, fmt.Errorf("cannot determine process state")
}

// CheckAndStartService will check if process alive and start
func CheckAndStartService(s *Service) (err error) {
	if s.Type != "oneshot" {
		alive, pid, err := checkIfProcessAlive(s)
		if err != nil {
			return err
		}

		if alive && pid != 0 {
			return fmt.Errorf("process %s already running as PID %d", s.Name, pid)
		} else if alive {
			return fmt.Errorf("process %s already running", s.Name)
		}
	}

	// start service
	if s.Type == "simple" {
		go s.StartSimple()
	} else {
		s.Start()
	}

	return nil
}

// We manage the "cmd" by CheckAndStopService, no simple/forking logic here
func shutdownProcess(s *Service, cmd string) (err error) {
	if s.ExecPreStop != "" {
		LoadedServicesMu.Lock()
		LoadedServices[s.Name].LastActionAt = time.Now().UTC().Unix()
		LoadedServices[s.Name].LastAction = PreStop
		LoadedServicesMu.Unlock()

		err = justExecACommand(s.ExecPreStop.String())
		if err != nil {
			clog.Error(2, "error in %s ExecPreStop: %s", s.Name, err.Error())
			return err
		}
	}

	err = justExecACommand(cmd)
	if err != nil {
		return err
	}

	if s.ExecPostStop != "" {
		LoadedServicesMu.Lock()
		LoadedServices[s.Name].LastActionAt = time.Now().UTC().Unix()
		LoadedServices[s.Name].LastAction = PostStop
		LoadedServicesMu.Unlock()

		err = justExecACommand(s.ExecPostStop.String())
		if err != nil {
			clog.Error(2, "error in %s ExecPostStop: %s", s.Name, err.Error())
			return err
		}
	}
	return nil
}

// CheckAndStopService will check if process running and stop
func CheckAndStopService(s *Service) (err error) {
	// Well, we don't really care if process is running, yeah ?
	LoadedServicesMu.Lock()
	defer LoadedServicesMu.Unlock()

	LoadedServices[s.Name].LastActionAt = time.Now().UTC().Unix()
	LoadedServices[s.Name].LastAction = Stop

	// If simple check struct status
	if s.Type == "simple" {
		if s.State == Starting || s.State == Started {
			// kill process according to cmd Shutdown
			if s.Shutdown != "" {
				err = shutdownProcess(s, s.Shutdown.String())
			} else {
				err = shutdownProcess(s, fmt.Sprintf("pkill %d", s.LastKnownPID))
			}
			if err != nil {
				LoadedServices[s.Name].State = Errored
				clog.Info("Service %s errored", s.Name)
				return err
			}
			LoadedServices[s.Name].State = Stopped
			clog.Info("Service %s stopped", s.Name)
			return err
		}
		LoadedServices[s.Name].State = Stopped
		clog.Info("Service %s isn't alive", s.Name)
		return fmt.Errorf("process %s doesn't seems to be alive ?", s.Name)
	}

	if s.Shutdown != "" {
		err = shutdownProcess(s, s.Shutdown.String())
	} else {
		err = fmt.Errorf("no Shutdown: command defined for %s, I don't know how to kill it", s.Name)
	}
	if err != nil {
		LoadedServices[s.Name].State = Errored
		clog.Info("Service %s errored", s.Name)
		return err
	}
	LoadedServices[s.Name].State = Stopped
	clog.Info("Service %s stopped", s.Name)
	return err
}