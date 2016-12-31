package main

import (
	"fmt"
	"os"
)

// udevd is part of systemd, but unfortunately it's the easiest way to get hw peripherals setup
// in Linux. I'd rather avoid it, but I don't know of a better alternative right now.
func systemdHacks() {
	if err := runquiet("udevd", "--daemon"); err != nil {
		fmt.Fprintf(os.Stderr, "Could not start udevd: %v\n", err)
	}
	if err := runquiet("udevadm", "trigger", "--action=add", "--type=subsystems"); err != nil {
		fmt.Fprintf(os.Stderr, "Could not add udevd subsystems: %v\n", err)
	}
	if err := runquiet("udevadm", "trigger", "--action=add", "--type=devices"); err != nil {
		fmt.Fprintf(os.Stderr, "Could not add udevd devices: %v\n", err)
	}
	if err := runquiet("udevadm", "settle"); err != nil {
		fmt.Fprintf(os.Stderr, "Could not wait for udev devices to be processed?\n")
	}

}
