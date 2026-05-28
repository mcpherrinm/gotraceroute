# Go Traceroute

Traceroute in Go.

I want something that doesn't use raw sockets or otherwise need root.

The goal here is to make a cli, library, and a service.

The service will expose traceroutes via Prometheus, useful for continuous monitoring.

## Design notes

We can do this mostly with golang.org/x/sys/unix, but it took me a bit to figure out how to actually do it.

The key ingredients are:

* UDP Socket: `unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP)`
* Set the TTL: `unix.SetsockoptInt(fd, unix.IPPROTO_IP, unix.IP_TTL, ttl)`
* Request errors: `unix.SetsockoptInt(fd, unix.IPPROTO_IP, unix.IP_RECVERR, 1)`
* Recieve errors: `unix.Recvmsg(fd, buf, oob, unix.MSG_ERRQUEUE)`
