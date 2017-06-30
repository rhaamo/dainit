# /etc/lutrainit/lutra.d/
## foo.service

    Name: ServiceFoo
    Description: Service Foo does lot of things and provides 'foobar'
    Type: forking
    PIDFile: /var/run/foo.pid
    
    Startup: start_foo.bar
    Shutdown: stop_foo.bar
    Provides: foobar
    
## baz.service

    Name: ServiceBaz
    Description: Baz runs after Foo and then needs 'foobar'
    Type: simple
    
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
- CheckAlive: One line command to check if service is alive, however it use PIDFile
- PIDFile: File the daemon store his PID. Should be mandatory for Type: forking
- Autostart: true/false, will the service started on boot ?
- Type:
  - forking: service is expected to fork by himself, PIDFile: would be great
  - oneshot: expected to fork by himself, no stop/status possible, it's a one-shot thing
  - simple: daemon doesn't fork by himself
  - virtual: used only for dependencies ordering
  
Provides are mandatory, you can just put here the Name of the service.

Needs are used for relationship, like udev can only be started when loopback have been brought up.

## Default values
- Autostart: true
- Type: forking
