# lelastic

### what is lelastic
lelastic is a elastic IP client designed to make the use of elastic IPs as easy as possible
it does not need any configuration dependencies or advanced bgp skills

### how it works
a few assumptions are made by this tool:
- your elastic IP needs to be configured on "lo" interface
- for ipv4 you need to configure the IP as a /32 tied to the loopback interface
- for ipv6 you can use any IP out of your /64 or /56 and of any subnet size as a loopback IP. if you want to announce your subnet as a /56 you need to toggle the flag `-send56`

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
  -send56
        Advertise ipv6 as /56 subnet
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

- to annnounce the IPv6 on your loopback as a /56 subnet
```
./lelastic -dcid 10 -primary -send56
```
