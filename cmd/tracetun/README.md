# Setup

    modprobe tun
    ip tuntap add dev tun0 mode tun
    ip link set tun0 up
    ip addr add 10.100.0.1/24 dev tun0

# Teardown

    ip link del tun0
