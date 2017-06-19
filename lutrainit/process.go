package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
	"time"
	"github.com/mitchellh/go-ps"
	"github.com/go-clog/clog"
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

		//fmt.Fprintf(os.Stderr, "[lutra] Waiting for processes to die (%d processes left)..\n", len(pids))
		clog.Info("[lutra] Waiting for processes to die (%d processes left)..\n", len(pids))

		for _, pid := range pids {
			p, err := ps.FindProcess(pid.Pid)
			if err != nil {
				return err
			}
			//fmt.Fprintf(os.Stderr, "[lutra] PID %d with name %s is still alive.\n", p.Pid(), p.Executable())
			clog.Info("[lutra] PID %d with name %s is still alive.\n", p.Pid(), p.Executable())
		}

		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("processes did not die after a minute")
}

// KillAll processes on the system.
func KillAll() {
	clog.Info("[lutra] Killing everything I can find...")
	//fmt.Fprintf(os.Stderr, "[lutra] Killing system processes..\n")
	clog.Info("[lutra] Killing system processes..")

	// Try to send a sigterm
	pids, err := getAllProcesses()
	if err != nil {
		clog.Error(2, err.Error())
		return
	}
	for _, proc := range pids {
		proc.Signal(syscall.SIGTERM)
	}

	if err := waitForDeath(); err != nil {
		clog.Error(2, err.Error())
	}
	// They didn't respond to sigterm after a minute, so be mean and send a SIGKILL
	pids, _ = getAllProcesses()
	for _, proc := range pids {
		proc.Signal(syscall.SIGKILL)
	}
	if len(pids) > 0 {
		//fmt.Fprintf(os.Stderr, "Sent kill signal to %d processes that didn't respond to term..\n", len(pids))
		clog.Info("Sent kill signal to %d processes that didn't respond to term..\n", len(pids))

		time.Sleep(2 * time.Second)
	}

	if err := waitForDeath(); err != nil {
		clog.Error(2, "%s :(", err.Error())
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
			clog.Error(2, err.Error())
			continue
		}
		rprocs = append(rprocs, proc)
	}
	return rprocs, nil
}


func doShutdown(reboot bool) {
	clog.Info("Shutdown or reboot initiated, please wait...")
	// TODO: Run shutdown scripts for services that are started instead
	// of just sending them a SIGTERM right off the bat..
	KillAll()

	// At this point we need to remove the file logger
	clog.Delete(clog.FILE)

	// This needs to be done after all the processes are dead, otherwise
	// it will fail due to being in use.
	clog.Info("[lutra] Unmounting filesystems...")
	UnmountAllExcept(append(NetFs, VirtFs...))

	// Halt the system explicitly to prevent a kernel panic.
	// Or reboot, as wanted.
	if reboot {
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	} else {
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
	}
}