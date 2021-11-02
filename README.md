# lelastic

### what is lelastic
lelastic is a elastic IP client designed to make the use of elastic IPs as easy as possible
it does not need any configuration dependencies or advanced bgp skills

### how it works
a few assumptions are made by this tool:
- your elastic IP is configured on the system (for ipv4 a /32 and for ipv6 a /64 or /56 is used)
- for ipv6 you have to use the entire block in a single allocation (as of right now). this can and will be improved

### install:
```
version=v0.0.1
curl -LO https://github.com/linode/lelastic/releases/download/$version/lelastic.gz
gunzip lelastic.gz
chmod 755 lelastic
mv lelastic /usr/local/bin/
```

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
