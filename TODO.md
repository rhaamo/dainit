# TO DO

## Tests
Add more tests

## Documentation

To be improved, code itself or external.

## Userland tool controller

ATM it seems that the init only starts services and that's all.

Introducing lutrinactl, using some socket like /var/otterland would permit to do the following:

- Services status
- Service start,stop,restart,status
- Init itself statistics (Go GC stats, memory, etc.)

## Process reload

An init like SySVInit or SystemD can reload itself using 'telinit u' or other command.

What would we need to do in Go to achieve the same result ?

Init shouldn't be killed, and would reuse PID1 then.

## Logging

Would be great to add better logging, verbose or not, file after / rw,remount.

## Getty improvements

Use some "special" config file for that.

Maybe treat them like systemd with a getty.whatever file.

## FS Mounts/umount

To be improved, more checks or support.

## Services files improvements

 What | Working on | Done
------|------------|------
PreExec | | |
PostExec | | |
Retry | | |
Timeout ? | | |

## Socket Activation

Basic socket activation, is it possible using Go ?