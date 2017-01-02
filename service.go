package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
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

func StartServices(services map[ServiceType][]*Service) {
	wg := sync.WaitGroup{}

	startedTypes := make(map[ServiceType]bool)
	for _, services := range services {
		wg.Add(len(services))
		for _, s := range services {
			go func(s *Service) {
				// TODO: This should ensure that Needs are satisfiable instead of getting into an
				// infinite loop when they're not.
				// (TODO(2): Prove N=NP in order to do the above efficiently.)
				for satisfied, tries := false, 0; satisfied == false && tries < 60; tries++ {
					satisfied = s.NeedsSatisfied(startedTypes)
					time.Sleep(2 * time.Second)

				}
				if s.state == notStarted {
					if err := s.Start(); err != nil {
						log.Println(err)
					}

				}

				// We don't use a mutex to handle startedTypes, because things only ever get set from false
				// to true, so if there's a race it's a race to set the variable to the same thing.
				for _, t := range s.Provides {
					startedTypes[t] = true
				}
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
	cmd := exec.Command("sh", "-c", s.Startup.String())
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		s.state = errored
		return err
	}
	s.state = started
	return nil
}

// Checks if all of s's needs are satified by the passed list of provided types
func (s Service) NeedsSatisfied(started map[ServiceType]bool) bool {
	for _, st := range s.Needs {
		if !started[st] {
			return false
		}
	}
	return true
}
