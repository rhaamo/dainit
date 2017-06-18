# /etc/lutrainit/lutra.conf

    # Should lutrainit persist after last tty exit ?
    Persist: true/false
    
    # Adds as many as you want of Autologin, any line will create a TTY with specified user auto-logged
    Autologin: foo
    Autologin: root
    
# /etc/lutrainit/lutra.d/
## foo.service

    Name: ServiceFoo
    Description: Service Foo does lot of things and provides 'foobar'
    
    Startup: start_foo.bar
    Shutdown: stop_foo.bar
    Provides: foobar
    
## baz.service

    Name: ServiceBaz
    Description: Baz runs after Foo and then needs 'foobar'
    
    Provides: baz
    Needs: foobar
    Startup: baz --start
    Shutdown: killall baz
    
# Explanations
## Service
- Name: name of the service, a-Z0-9 without spaces, - and _ allowed
- Description: One line description of the service
- Relations
  - Provides: foo,bar
  - Needs: bar,other
  
  Separate multiples keywords with `,`. Only a-Z0-9 - and _ allowed
- Startup: One line command to start service
- Shutdown: One line command to stop service

Provides are mandatory, you can just put here the Name of the service.

Needs are used for relationship, like udev can only be started when loopback have been brought up.