package main

import (
	"log"
	"os"
	"strings"
)

// Remount a filesystem. Grub mounts / as ro during the boot process, and this will get it to
// be readwrite (assuming it's rw in /etc/fstab)
func Remount(dir string) {
	if err := run("mount", "-o", "remount", dir); err != nil {
		log.Println(err)
	}
}

// Mount a filesystem, creating the mount point if it doesn't exist.
func Mount(typ, device, dir, opts string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0775); err != nil {
			log.Printf("Could not create mount point %v: %v\n", dir, err)
			return
		}
	}
	if err := run("mount", "-t", typ, device, dir, "-o", opts); err != nil {
		log.Println(err)
	}
}

// MountAllExcept for ones of the type passed in the except parameter
func MountAllExcept(except []string) {
	noexcept := make([]string, len(except))
	for i, val := range except {
		noexcept[i] = "no" + val
	}
	if err := run("mount", "-a", "-t", strings.Join(noexcept, ","), "-O", "no_netdev"); err != nil {
		log.Println(err)
	}
}

// UnmountAllExcept for netdev filesystems and those passed in the except
// parameter
func UnmountAllExcept(except []string) {
	noexcept := make([]string, len(except))
	for i, val := range except {
		noexcept[i] = "no" + val
	}
	if err := run("umount", "-a", "-t", strings.Join(noexcept, ","), "-O", "no_netdev"); err != nil {
		log.Println(err)
	}
}
