package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
	"github.com/rhaamo/lutrainit/shared/ipc"
	"github.com/go-clog/clog"
)

type runState uint8

// Types of valid runState
const (
	notStarted = runState(iota)
	starting
	started
	stopped
	errored
)

func (rs runState) String() string {
	switch rs {
	case notStarted:
		return "not started"
	case starting:
		return "being started"
	case started:
		return "already started"
	case stopped:
		return "stopped"
	case errored:
		return "errored"
	default:
		return "in an invalid state"
	}
}

// Service represents a struct with usefull infos used for management of services
type Service struct {
	Name     	ServiceName
	Description	string
	Startup  	Command
	Shutdown 	Command
	Provides 	[]ServiceType
	Needs    	[]ServiceType

	Type		string	// forking or simple
	PIDFile		string

	state 		runState
}

// StartServices starts all declared services
func StartServices() {
	wg := sync.WaitGroup{}

	var startedMu = &sync.RWMutex{}
	startedTypes := make(map[ServiceType]bool)
	for _, services := range StartupServices {
		wg.Add(len(services))
		for _, s := range services {
			go func(s *Service) {
				// TODO: This should ensure that Needs are satisfiable instead of getting into an
				// infinite loop when they're not.
				// (TODO(2): Prove N=NP in order to do the above efficiently.)
				for satisfied, tries := false, 0; satisfied == false && tries < 60; tries++ {
					satisfied = s.NeedsSatisfied(startedTypes, startedMu)
					time.Sleep(2 * time.Second)

				}
				if s.state == notStarted {
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
			}(s)
		}
	}
	wg.Wait()
}

// Start the Service s. if type is oneshot or forking
func (s *Service) Start() error {
	if s.state != notStarted {
		return fmt.Errorf("Service %v is %v", s.Name, s.state.String())
	}
	s.state = starting
	LoadedServices[ipc.ServiceName(s.Name)].State = ipc.Starting
	LoadedServices[ipc.ServiceName(s.Name)].LastActionAt = time.Now().UTC().Unix()
	LoadedServices[ipc.ServiceName(s.Name)].LastAction = ipc.Start

	cmd := exec.Command("sh", "-c", s.Startup.String())
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		s.state = errored
		LoadedServices[ipc.ServiceName(s.Name)].State = ipc.Errored
		LoadedServices[ipc.ServiceName(s.Name)].LastActionAt = time.Now().UTC().Unix()
		LoadedServices[ipc.ServiceName(s.Name)].LastAction = ipc.Start

		clog.Error(2,"[lutra] Error starting service %s: %s", s.Name, err.Error())

		return err
	}
	s.state = started
	LoadedServices[ipc.ServiceName(s.Name)].State = ipc.Started
	LoadedServices[ipc.ServiceName(s.Name)].LastActionAt = time.Now().UTC().Unix()
	LoadedServices[ipc.ServiceName(s.Name)].LastAction = ipc.Start

	clog.Info("[lutra] Started service %s", s.Name)

	return nil
}

// StartSimple and track the PID and process state (for simple service without auto-fork function)
func(s *Service) StartSimple() {
	s.state = starting
	LoadedServices[ipc.ServiceName(s.Name)].State = ipc.Starting
	LoadedServices[ipc.ServiceName(s.Name)].LastActionAt = time.Now().UTC().Unix()
	LoadedServices[ipc.ServiceName(s.Name)].LastAction = ipc.Start
	LoadedServices[ipc.ServiceName(s.Name)].LastKnownPID = 0

	cmd := exec.Command("sh", "-c", s.Startup.String())
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		clog.Error(2,"[lutra] Service %s exited with error: %s", s.Name, err.Error())
		s.state = errored
		LoadedServices[ipc.ServiceName(s.Name)].State = ipc.Errored
		LoadedServices[ipc.ServiceName(s.Name)].LastMessage = err.Error()
		LoadedServices[ipc.ServiceName(s.Name)].LastKnownPID = 0
		return
	}
	// Waiting for the command to finish
	s.state = started
	LoadedServices[ipc.ServiceName(s.Name)].State = ipc.Started
	LoadedServices[ipc.ServiceName(s.Name)].LastKnownPID = cmd.Process.Pid
	clog.Info("[lutra] Started service %s", s.Name)

	err := cmd.Wait()
	if err != nil {
		clog.Error(2, "[lutra] Service %s finished with error: %s", s.Name, err.Error())
		s.state = stopped
		LoadedServices[ipc.ServiceName(s.Name)].State = ipc.Stopped
		LoadedServices[ipc.ServiceName(s.Name)].LastMessage = err.Error()
		LoadedServices[ipc.ServiceName(s.Name)].LastKnownPID = 0
	} else {
		s.state = stopped
		clog.Info("[lutra] Service stopped:	 %s", s.Name)
		LoadedServices[ipc.ServiceName(s.Name)].State = ipc.Stopped
		LoadedServices[ipc.ServiceName(s.Name)].LastKnownPID = 0
		LoadedServices[ipc.ServiceName(s.Name)].LastActionAt = time.Now().UTC().Unix()
		LoadedServices[ipc.ServiceName(s.Name)].LastAction = ipc.Stop
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
