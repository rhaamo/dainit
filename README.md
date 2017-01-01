# dainit

dainit is a simple init system for Linux written in Go, mostly for me to
to use on my laptop. It's likely less full-featured and more minimalist than
your init system.

dainit will do the following:

1. Set the hostname
2. Remount the root filesystem*
3. Mount all other non-network filesystems and activate swap partitions
3. Start processes with config files in /etc/dainit after their dependencies
   ("Needs") are started. See `conf/` for a sample udevd and wpa_supplicant (which
   depends on udevd to finish before being started) config.
4. Start a tty and wait for a login.
5. Kill running processes, unmount filesystems, and poweroff the system once that
login session ends.

(The way step 3 is handled isn't very elegant and will likely fail if you have
slow startup processes.)

## Installation/Usage
```
sudo cp $GOPATH/bin/dainit /sbin
sudo update-initramfs -u
```

then add `init=/sbin/dainit` to your grub configuration (or add it to
`SUPPORTED_INITS` at the top of `/etc/grub.d/10_linux` and run `grub-update`**)

* This shouldn't be required, since `mount -a` should take care of it in step
  3 according to mount(8), but as far as I can tell it doesn't.
** At least, in the way that grub was configured by Debian on my system, that
   worked.
