# /etc/lutrainit/lutra.conf

    # Should lutrainit persist after last tty exit ?
    Persist: true/false
    
    # Adds as many as you want of Autologin, any line will create a TTY with specified user auto-logged
    Autologin: foo
    Autologin: root
    
# /etc/lutrainit/lutra.d/
## foo.service

    # ServiceFoo
    
    Service Foo does lot of things and provides 'foobar'
    
    Startup: start_foo.bar
    Shutdown: stop_foo.bar
    Provides: foobar
    
## baz.service

    # ServiceBaz
    
    Baz runs after Foo and then needs 'foobar'
    
    Provides: baz
    Needs: foobar
    Startup: baz --start
    Shutdown: killall baz