package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
)

// Gettys spawn the number of ttys required for len(autologins) to login.
// If persist is true, they'll be respawned if they die.
func Gettys(autologins []string, persist bool) {
	switch len(autologins) {
	case 0, 1:
		// If there's no autologins, we still want to spawn a tty, and if
		// there's one, there's no need to coordinate goroutines
		var user string
		if len(autologins) == 1 {
			user = autologins[0]
		}
		tty := "tty1"

		for {
			if err := getty(user, tty); err != nil {
				log.Println(err)
			}
			if !persist {
				return
			}
		}
	default:
		// getty(user, tty) blocks, so spawn a goroutine for each one and wait
		// for them to finish with a waitgroup, respawning as necessary in the
		// goroutine if it happens to quit. (NB: if persist is true they will
		// never finish.)
		wg := sync.WaitGroup{}
		wg.Add(len(autologins))
		for i, user := range autologins {
			go func(user, tty string) {
				defer wg.Done()
				for {
					if err := getty(user, tty); err != nil {
						log.Println(err)
					}
					if !persist {
						return
					}
				}
			}(user, "tty"+strconv.Itoa(i+1))
		}
		// Block until all the ttys we spawned in goroutines are finished instead of
		// returning right away (and shutting down the system.)
		wg.Wait()
	}
}

// Spawn a single tty on tty, logging in with user autologin.
func getty(autologin, tty string) error {
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

	return cmd.Run()
}
