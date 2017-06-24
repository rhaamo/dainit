# lutrainit

lutrainit is an init system for Linux written in Go.

lutrainit will do the following:

1. Set the hostname
2. Remount the root filesystem[1]
3. Mount all other non-network filesystems and activate swap partitions
3. Start processes with config files in /etc/lutrainit/lutra.d/ after their dependencies
   ("Needs") are started. See `conf/` for a samples config.
4. Start some TTY or anything other user-specified.
5. Kill running processes, unmount filesystems, and poweroff the system once that last
   login session ends.

(The way step 4 is handled isn't very elegant and will likely fail if you have too
many slow startup processes.)

You can also create a file `/etc/lutrainit/lutra.conf` for some basic configuration.

If there are any lines of the form `autologin: username` it will automatically log in
as that username. (If there's multiple autologin directives, it will create the
appropriate number of ttys.)
If any line contains `persist: true`, then when a tty exits, it'll respawn the tty instead of powering down the system once all the ttys are gone.

See the `conf/` folder for what you can put in `/etc/lutrainit/` including services files in `/etc/lutrainit/lutra.d/`.

## lutractl

A tool exists and communicate with the init daemon using RPC on socket `/run/ottersock`, it can then show init version, statistics about goroutines, memory, etc.

## Installation/Usage
```
sudo cp lutrainit/lutrainit /sbin
sudo cp lutractl/lutractl /sbin
```

then add `init=/sbin/lutrainit` to your grub configuration. (Or alternatively, make
`/sbin/init` a symlink to `lutrainit`.)

[1] This shouldn't be required, since `mount -a` should take care of it in step
  3 according to mount(8), but as far as I can tell it doesn't.
