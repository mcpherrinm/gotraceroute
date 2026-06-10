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
