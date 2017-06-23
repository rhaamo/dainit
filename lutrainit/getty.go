package main

import (
	"os"
	"os/exec"
	"sync"
	"syscall"
	"github.com/go-clog/clog"
)

var (
	// 11 ttys from 0 to 12 but tty7 is excluded as it is frequently used for a X display
	ttys = [11]string{"tty1", "tty2", "tty3", "tty4", "tty5", "tty6", "tty8", "tty9", "tty10", "tty11", "tty12"}

	// GettysList of managed or unmanaged
	GettysList = make(map[int]*FollowGetty)
)

type FollowGetty struct {
	TTY			string
	PID			int
	Managed		bool
	Autologin	string
}

func ManageGettys() {
	if MainConfig.StartedReexec {

	} else {
		manageAndSpawnGettys()
	}
}

// Gettys spawn the number of ttys required for len(autologins) to login.
// If persist is true, they'll be respawned if they die.
func manageAndSpawnGettys() {
	if MainConfig.StartedReexec {}
	autologins := MainConfig.Autologins

	switch len(autologins) {
	case 0, 1:
		// If there's no autologins, we still want to spawn a tty, and if
		// there's one, there's no need to coordinate goroutines
		var user string
		if len(autologins) == 1 {
			user = autologins[0]
		}

		for {
			GettysList[0] = &FollowGetty{TTY: ttys[0], Managed: true}
			if user != "" {
				GettysList[0].Autologin = user
			}
			if err := spawnGetty(user, ttys[0], GettysList[0]); err != nil {
				GettysList[0].PID = 0
				clog.Error(2, err.Error())
			}

			// if no persistency or going shutdown, exit loop
			if !MainConfig.Persist || ShuttingDown {
				return
			}
		}
	default:
		// getty(user, tty) blocks, so spawn a goroutine for each one and wait
		// for them to finish with a waitgroup, respawning as necessary in the
		// goroutine if it happens to quit. (NB: if persist is true they will
		// never finish.)
		wg := sync.WaitGroup{}

		// Sanity check length of autologins
		if len(autologins) > 11 {
			autologins = autologins[:11]
		}

		wg.Add(len(autologins))
		for i, user := range autologins {
			go func(user, tty string, idx int) {
				defer wg.Done()
				for {
					GettysList[idx] = &FollowGetty{TTY: tty, Managed: true, Autologin: user}
					if err := spawnGetty(user, tty, GettysList[idx]); err != nil {
						GettysList[idx].PID = 0
						clog.Error(2, err.Error())
					}

					// if no persistency or going shutdown, exit loop
					if !MainConfig.Persist || ShuttingDown {
						return
					}
				}
			}(user, ttys[i], i)
		}
		// Block until all the ttys we spawned in goroutines are finished instead of
		// returning right away (and shutting down the system.)
		wg.Wait()
	}
}

// Spawn a single tty on tty, logging in with user autologin.
func spawnGetty(autologin, tty string, gettyFollow *FollowGetty) error {
	clog.Info("Spawning getty on %s with user %s", tty, autologin)

	var cmd *exec.Cmd
	if autologin != "" {
		cmd = exec.Command("getty", "--noclear", tty, "--autologin", autologin)
	} else {
		cmd = exec.Command("getty", "--noclear", tty)
	}

	// If we don't Setsid, we'll get an "inappropriate ioctl for device"
	// error upon starting the login shell.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		clog.Error(2, "[lutra] Getty %s exited with error: %s", tty, err.Error())
		return err
	}

	gettyFollow.PID = cmd.Process.Pid

	return cmd.Wait()
}
