package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
	"github.com/rhaamo/lutrainit/shared/ipc"
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

type Service struct {
	Name     ServiceName
	Startup  Command
	Shutdown Command
	Provides []ServiceType
	Needs    []ServiceType

	state runState
}

func StartServices() {
	wg := sync.WaitGroup{}

	var startedMu *sync.RWMutex = &sync.RWMutex{}
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
					if err := s.Start(); err != nil {
						log.Println(err)
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

// Starts the Service s.
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

		return err
	}
	s.state = started
	LoadedServices[ipc.ServiceName(s.Name)].State = ipc.Started
	LoadedServices[ipc.ServiceName(s.Name)].LastActionAt = time.Now().UTC().Unix()
	LoadedServices[ipc.ServiceName(s.Name)].LastAction = ipc.Start

	return nil
}

// Checks if all of s's needs are satified by the passed list of provided types
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
