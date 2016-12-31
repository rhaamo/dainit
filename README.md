# dainit

dainit is a simple init system for Linux written in Go to use on my laptop.

dainit will currently do the following:

1. Set the hostname
2. Remount the root filesystem rw
3. Start udevd (TODO: Find a systemd-free replacement)
4. Start a tty and wait for a login.
5. Kill running processes, unmount filesystems, and poweroff the system once that login
   session ends.

It will not do anything else right now, and maybe never will. (It'll probably eventually
support at least enough to bring up my network device.)

## Installation/Usage
```
sudo cp $GOPATH/bin/dainit /sbin
sudo update-initramfs -u
```

then add init=/sbin/dainit to your grub configuration
