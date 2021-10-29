# lelastic

### what is my-dhcpd
lelastic is a elastic IP client designed to make the use of elastic IPs as easy as possible
it does not need any configuration dependencies or advanced bgp skills


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
