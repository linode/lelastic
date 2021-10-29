# lelastic

### what is my-dhcpd
lelastic is a elastic IP client designed to make the use of elastic IPs as easy as possible
it does not need any configuration dependencies or advanced bgp skills


### usage:
```

Usage of ./lelastic:
  -dcid int
        dcid for your DC
  -logjson
        set log format to json
  -loglevel string
        set log level: trace, debug, info or warn (default "info")
  -primary
        advertise as primary
  -secondary
        advertise as secondary
```


### example:


- to annnounce the linode as primary simply run
```
./lelastic -dcid 10 -primary
```

- to annnounce the linode as secondary simply run
```
./lelastic -dcid 10 -secondary
```
