# dainit

dainit is a simple init system for Linux written in Go, mostly for me to
to use on my laptop. It's likely less full-featured and more minimalist than
your init system. (It aims to be an init system and only an init system.)

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

(The way step 3 is handled isn't very elegant and will likely fail if you have too
many slow startup processes.)

You can also create a file /etc/dainit.conf for some basic configuration. If
there are any lines of the form "Autologin: username" it will automatically log in
as that username. (If there's multiple autologin directives, it will create the
appropriate number of ttys.) If any line contains "Persist: true", then when a tty
exits, it'll respawn the tty instead of powering down the system once all the ttys
are gone.

## Installation/Usage
```
sudo cp $GOPATH/bin/dainit /sbin
sudo update-initramfs -u
```

then add `init=/sbin/dainit` to your grub configuration. (Or alternatively, make
`/sbin/init` a symlink to `dainit`.)

* This shouldn't be required, since `mount -a` should take care of it in step
  3 according to mount(8), but as far as I can tell it doesn't.
