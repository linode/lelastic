# lelastic

### what is lelastic
lelastic is a elastic IP client designed to make the use of elastic IPs as easy as possible
it does not need any configuration dependencies or advanced bgp skills

### caveats
- this tool is completely optional and any bgp client/daemon of your choice will do the job.
- the sole purpose of this tool is to get you up and running as fast and easy as possible
- it does not support any fancy or advanced bgp trickery

### notes
- lelastic is basically a wrapper for gobgp (https://github.com/osrg/gobgp)
- lelastic when started up exposes the gobgp API, just like gobgp does on port 50051
	- since version 0.0.6 this is bound to localhost only. prior to this it was actually exposed

### how it works
a few assumptions are made by this tool:
- your elastic IP needs to be configured on "lo" interface
- for ipv4 you need to configure the IP as a /32 tied to the loopback interface
- for ipv6 you can use any IP out of your /64 or /56 and of any subnet size as a loopback IP. if you want to announce your subnet as a /56 you need to toggle the flag `-send56` otherwise it will default to announcing a /64


### install:
```
version=v0.0.6
curl -LO https://github.com/linode/lelastic/releases/download/$version/lelastic.gz
gunzip lelastic.gz
chmod 755 lelastic
mv lelastic /usr/local/bin/
```

#### containerizing:

You can containzerize lelastic using Docker and push it to a registry. Use this approach if you'd like to use lelastic in a kubernetes cluster.

* First, be on a Linux machine or a Linode.
* Build the container: `cd lelastic && docker build -t lelastic .`
* Tag the container: `docker tag lelastic your-name-here/lelastic:latest`
* Optionally run on your machine to test it first: `docker run -dp 3000:3000 lelastic` 
* Push it to a registry for your kubernetes cluster to pull from later: `docker push your-name-here/lelastic:latest`


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
        Advertise ipv6 as /56 subnet (defaults to /64)
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
