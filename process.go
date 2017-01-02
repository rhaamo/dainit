package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"syscall"
	"time"
)

// Waits up to a minute for all processes to die.
func waitForDeath() error {
	for i := 0; i < 30; i++ {
		pids, err := getAllProcesses()
		if err != nil {
			return err
		}
		if len(pids) == 0 {
			return nil
		}

		fmt.Fprintf(os.Stderr, "Waiting for processes to die (%d processes left)..\n", len(pids))
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("Processes did not die after a minute.")
}

// Kills all processes on the system.
func KillAll() {
	fmt.Fprintf(os.Stderr, "Killing system processes..\n")
	// Try to send a sigterm
	pids, err := getAllProcesses()
	if err != nil {
		log.Println(err)
	}
	for _, proc := range pids {
		proc.Signal(syscall.SIGTERM)
	}

	if err := waitForDeath(); err != nil {
		log.Println(err)
	}
	// They didn't respond to sigterm after a minute, so be mean and send a SIGKILL
	pids, _ = getAllProcesses()
	for _, proc := range pids {
		proc.Signal(syscall.SIGKILL)
	}
	if len(pids) > 0 {
		fmt.Fprintf(os.Stderr, "Sent kill signal to %d processes that didn't respond to term..\n", len(pids))
		time.Sleep(2 * time.Second)
	}

	if err := waitForDeath(); err != nil {
		log.Println(err, " :(")
	}
}

// Get a list of all processes on the system by checking /proc/*/cmdline files
func getAllProcesses() ([]*os.Process, error) {
	procs, err := ioutil.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	rprocs := make([]*os.Process, 0)
	for _, f := range procs {
		if !f.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(f.Name())
		if err != nil {
			// The directory wasn't an integer, so it wasn't a pid.
			continue
		}
		if pid < 2 {
			// don't include the init system in the procs that get killed.
			continue
		}
		cmdline := fmt.Sprintf("/proc/%d/cmdline", pid)
		if _, err := os.Stat(cmdline); os.IsNotExist(err) {
			// There was no command line, it's not a process to kill
			continue
		}
		contents, err := ioutil.ReadFile(cmdline)
		if len(contents) == 0 {
			// the cmdline file was empty, it's not a real command
			continue
		}
		proc, err := os.FindProcess(pid)
		if err != nil {
			log.Println(err)
		}
		rprocs = append(rprocs, proc)
	}
	return rprocs, nil
}
