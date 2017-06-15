package main

import (
	"syscall"
)

// Continually reaps children in a background goroutine.
func reapChildren() {
	for {
		// If there are multiple zombies, Wait4 will reap one, and then block until the next child changes state.
		// We call it with NOHANG a few times to clear up any backlog, and then make a blocking call until our
		// next child dies.
		var status syscall.WaitStatus

		// If there are more than 10 zombies, we likely have other problems.
		for i := 0; i < 10; i++ {
			// We don't really care what the pid was that got reaped, or if there's nothing to wait for
			syscall.Wait4(-1, &status, syscall.WNOHANG, nil)
		}

		// This blocks, so that we're not spending all of our time reaping processes..
		syscall.Wait4(-1, &status, 0, nil)
	}
}
