package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	sequenceNum := uint32(0)

	for {
		// Create a new packet with the current timestamp
		timestampSent := time.Now().UnixMilli()
		payload := fmt.Sprintf("Timestamp: %d", timestampSent)
		packet := rudp_protocol_kit.NewRudpPacket(sequenceNum, []byte(payload), 0)

		// Send the packet
		err := rudp_protocol_kit.RudpSend(conn, packet, serverAddr)
		if err != nil {
			fmt.Println("Failed to send packet:", err)
		} else {
			fmt.Printf("Packet with sequence number %d sent.\n", sequenceNum)
		}

		sequenceNum++
		time.Sleep(5 * time.Second) // Send every 5 seconds
	}
}
