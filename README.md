# Go Traceroute

Traceroute in Go.

I want something that doesn't use raw sockets or otherwise need root.

The goal here is to make a library as well as a prometheus exporter.

## Design notes

I think in Go we can do this all with golang.org/x/sys/unix


UDP Socket: `unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP)`
Set the TTL: `unix.SetsockoptInt(fd, unix.IPPROTO_IP, unix.IP_TTL, ttl)`
Request errors: `unix.SetsockoptInt(fd, unix.IPPROTO_IP, unix.IP_RECVERR, 1)`

We can use `unix.Recvmsg` and then `unix.ParseSocketControlMessage` to get
the responses requested with `IP_RECVERR`.

The trickiest bit seems like it's actually knowing when to call unix.Recvmsg.
I think we might be able to use syscall.RawConn.Read, but need to try it out.
