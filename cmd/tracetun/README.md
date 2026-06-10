This is a utility which replies with ICMP Time Exceeded packets
Useful for testing.

# Setup

    modprobe tun
    ip tuntap add dev tun0 mode tun
    ip link set tun0 up
    ip addr add 10.100.0.1/24 dev tun0

# Run

    go run ./cmd/tracetun tun0 cmd/tracetun/bad.horse.json

# Teardown

    ip link del tun0
