# Trace Tun

This is a utility which replies with ICMP Time Exceeded packets.

It listens on a tun device and sends replies based on the incoming TTL.

Useful for testing traceroute.

## Setup

    modprobe tun
    ip tuntap add dev tun0 mode tun
    ip link set tun0 up
    ip addr add 10.100.0.1/24 dev tun0

## Run

    go run ./cmd/tracetun tun0 cmd/tracetun/bad.horse.json

## Teardown

    ip link del tun0

## Docker

To be able to create the tun device, you need

    docker run --cap-add=NET_ADMIN --device=/dev/net/tun:/dev/net/tun ...

There's a Containerfile here which builds tracetun and includes `ip` for setup.

    docker build . -t tracetun -f cmd/tracetun/Containerfile
