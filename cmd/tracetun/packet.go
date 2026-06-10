package main

import (
	"encoding/binary"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	ipv4HeaderLen = 20
	replyTTL      = 64
	protoICMP     = 1
	maxQuote      = 548
)

func icmpTimeExceeded(src, dst, in []byte) []byte {
	icmpBytes, err := (&icmp.Message{
		Type: ipv4.ICMPTypeTimeExceeded,
		Code: 0,
		Body: &icmp.TimeExceeded{Data: in[:min(len(in), maxQuote)]},
	}).Marshal(nil) // ICMP checksum computed and filled here
	if err != nil {
		return nil
	}
	return ipv4Packet(src, dst, protoICMP, icmpBytes)
}

func ipv4Packet(src, dst []byte, proto byte, payload []byte) []byte {
	pkt := make([]byte, ipv4HeaderLen+len(payload))
	pkt[0] = 0x45 // version 4, IHL 5 words
	binary.BigEndian.PutUint16(pkt[2:4], uint16(len(pkt)))
	pkt[8] = replyTTL
	pkt[9] = proto
	copy(pkt[12:16], src)
	copy(pkt[16:20], dst)
	binary.BigEndian.PutUint16(pkt[10:12], ipChecksum(pkt[:ipv4HeaderLen]))
	copy(pkt[ipv4HeaderLen:], payload)
	return pkt
}

func ipChecksum(b []byte) uint16 {
	var sum uint32
	for i := 0; i+1 < len(b); i += 2 {
		sum += uint32(b[i])<<8 | uint32(b[i+1])
	}
	if len(b)%2 == 1 {
		sum += uint32(b[len(b)-1]) << 8
	}
	for sum>>16 != 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^uint16(sum)
}
