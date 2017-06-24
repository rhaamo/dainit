# /etc/lutrainit/lutra.conf
INI style configuration.

    [global]
    ; Should lutrainit persist after last tty exit ?
    persist=true
    
    ; If you want auto login set the following line with any user
    ; For multiple user (one per tty) or non-autologin tty separate them with a ,
    ; One tty, no autologin
    ; autologin=
    ; Two tty, first autologin, second non-autologin
    ; autologin=kiosk,
    ; six non-autologin ttys
    autologin=,,,,,
    
    [logging]
    filename=/var/log/lutrainit.log
    ; This enables automated log rotate (switch of following options)
    rotate = true
    ; Segment log daily
    rotate_daily = true
    ; Max size shift of single file, default is 28 means 1 << 28, 256MB
    max_size_shift = 28
    ; Max line number of single file
    max_lines = 1000000
    ; Expired days of log file (delete after max days)
    max_days = 7
    
## Default values
- persist: true
- autologin: ,
  - two non-autologin ttys (tty1 and tty2)
- logging
  - filename: /var/log/lutrainit.log
  - rotate: true
  - rotate_daily: true
  - max_size_shift: 28
  - max_lines: 1000000
  - max_days: 7
