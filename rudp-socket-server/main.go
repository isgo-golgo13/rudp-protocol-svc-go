package main

import (
	"fmt"
	"net"
	rudp_protocol_kit "rudp-protocol-svc/rudp-protocol-kit"
	storage_kit "rudp-protocol-svc/storage-kit"
	"sort"
	"time"
)

type PacketBuffer struct {
	buffer              map[uint32]rudp_protocol_kit.RudpPacket // Packet buffer
	expectedSequenceNum uint32                                  // The next expected sequence number
}

func NewPacketBuffer() *PacketBuffer {
	return &PacketBuffer{
		buffer:              make(map[uint32]rudp_protocol_kit.RudpPacket),
		expectedSequenceNum: 0,
	}
}

// Buffer out-of-order packets
func (pb *PacketBuffer) bufferPacket(packet rudp_protocol_kit.RudpPacket) {
	pb.buffer[packet.SequenceNum] = packet
}

// Process the current packet if it matches the expected sequence number
func (pb *PacketBuffer) processPacket(packet rudp_protocol_kit.RudpPacket, repo *storage_kit.PacketRepository) bool {
	if packet.SequenceNum == pb.expectedSequenceNum {
		pb.storePacket(packet, repo)
		pb.expectedSequenceNum++
		return true
	}
	return false
}

// Store the packet in the ScyllaDB database
func (pb *PacketBuffer) storePacket(packet rudp_protocol_kit.RudpPacket, repo *storage_kit.PacketRepository) {
	timestampReceived := time.Now().UnixMilli()
	fmt.Printf("Processing packet with sequence number %d: %s\n", packet.SequenceNum, string(packet.Data))
	err := repo.SavePacket(int(packet.SequenceNum), packet.TimestampSent, int(packet.RetryCount), packet.RetryCount > 0)
	if err != nil {
		fmt.Printf("Failed to store packet: %v\n", err)
	}
}

// Process any buffered packets that are now in order
func (pb *PacketBuffer) processBufferedPackets(repo *storage_kit.PacketRepository) {
	// Extract sequence numbers
	keys := make([]uint32, 0, len(pb.buffer))
	for k := range pb.buffer {
		keys = append(keys, k)
	}

	// Sort the sequence numbers
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	// Process packets in order
	for _, sequenceNum := range keys {
		if sequenceNum == pb.expectedSequenceNum {
			packet := pb.buffer[sequenceNum]
			pb.storePacket(packet, repo)
			pb.expectedSequenceNum++
			delete(pb.buffer, sequenceNum) // Remove processed packet
		}
	}
}

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Initialize the storage and packet buffer
	storage, err := storage_kit.NewScyllaStorage("scylla-db")
	if err != nil {
		panic(err)
	}
	defer storage.Close()

	repo := storage_kit.NewPacketRepository(storage)
	packetBuffer := NewPacketBuffer()

	for {
		// Receive a packet
		packet, _, err := rudp_protocol_kit.RudpRecv(conn)
		if err != nil {
			fmt.Println("Failed to receive packet:", err)
			continue
		}

		fmt.Printf("Received packet with sequence number %d\n", packet.SequenceNum)

		// Check if the packet is the expected one
		if packetBuffer.processPacket(packet, repo) {
			// Process any buffered packets that can now be processed in order
			packetBuffer.processBufferedPackets(repo)
		} else if packet.SequenceNum > packetBuffer.expectedSequenceNum {
			// Packet is out of order, buffer it
			fmt.Printf("Buffering out-of-order packet with sequence number %d\n", packet.SequenceNum)
			packetBuffer.bufferPacket(packet)
		} else {
			// Duplicate or old packet, ignore it
			fmt.Printf("Ignoring duplicate or old packet with sequence number %d\n", packet.SequenceNum)
		}

		// Send ACK
		ackPacket := rudp_protocol_kit.RudpPacket{
			SequenceNum: packet.SequenceNum,
			Ack:         true,
			Data:        []byte{},
			RetryCount:  0,
		}
		err = rudp_protocol_kit.RudpSend(conn, ackPacket, serverAddr)
		if err != nil {
			fmt.Println("Failed to send ACK:", err)
		} else {
			fmt.Printf("ACK sent for sequence number %d\n", packet.SequenceNum)
		}
	}
}
