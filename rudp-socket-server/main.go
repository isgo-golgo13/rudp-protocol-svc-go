package main

import (
	"fmt"
	"net"
	"time"

	rudp_protocol_kit "rudp-protocol-svc/rudp-protocol-kit"
	storage_kit "rudp-protocol-svc/storage-kit"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Initialize ScyllaDB storage
	storage, err := storage_kit.NewScyllaStorage("scylla-db")
	if err != nil {
		panic(err)
	}
	defer storage.Close()

	repo := storage_kit.NewPacketRepository(storage)

	for {
		// Receive a packet
		packet, clientAddr, err := rudp_protocol_kit.RudpRecv(conn)
		if err != nil {
			fmt.Println("Failed to receive packet:", err)
			continue
		}

		fmt.Printf("Received packet with sequence number %d: %s\n", packet.SequenceNum, packet.Data)

		// Save the packet data to ScyllaDB
		timestampSent := time.UnixMilli(packet.TimestampSent)
		if err := repo.SavePacket(int(packet.SequenceNum), timestampSent, int(packet.RetryCount), packet.RetryCount > 0); err != nil {
			fmt.Println("Failed to save packet:", err)
		}

		// Send an ACK back
		ackPacket := rudp_protocol_kit.NewRudpPacket(packet.SequenceNum, []byte{}, 0)
		ackPacket.Ack = true
		err = rudp_protocol_kit.RudpSend(conn, ackPacket, clientAddr)
		if err != nil {
			fmt.Println("Failed to send ACK:", err)
		} else {
			fmt.Printf("ACK sent for sequence number %d\n", packet.SequenceNum)
		}
	}
}
