package rudp_protocol_kit

import (
	"encoding/json"
	"net"
	"time"
)

const MAX_DATA_SIZE = 512

// RUDP packet structure
type RudpPacket struct {
	SequenceNum   uint32 `json:"sequence_num"`
	Ack           bool   `json:"ack"`
	Data          []byte `json:"data"`
	RetryCount    uint32 `json:"retry_count"`
	TimestampSent int64  `json:"timestamp_sent"`
}

// Create a new RUDP packet
func NewRudpPacket(sequenceNum uint32, data []byte, retryCount uint32) *RudpPacket {
	return &RudpPacket{
		SequenceNum:   sequenceNum,
		Ack:           false,
		Data:          data,
		RetryCount:    retryCount,
		TimestampSent: time.Now().UnixMilli(),
	}
}

// Encode the RUDP packet to bytes
func (p *RudpPacket) ToBytes() ([]byte, error) {
	return json.Marshal(p)
}

// Decode the RUDP packet from bytes
func FromBytes(data []byte) (*RudpPacket, error) {
	var packet RudpPacket
	err := json.Unmarshal(data, &packet)
	if err != nil {
		return nil, err
	}
	return &packet, nil
}

// Send a RUDP packet over UDP
func RudpSend(conn *net.UDPConn, packet *RudpPacket, addr *net.UDPAddr) error {
	data, err := packet.ToBytes()
	if err != nil {
		return err
	}
	_, err = conn.WriteToUDP(data, addr)
	return err
}

// Receive a RUDP packet over UDP
func RudpRecv(conn *net.UDPConn) (*RudpPacket, *net.UDPAddr, error) {
	buffer := make([]byte, MAX_DATA_SIZE)
	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, nil, err
	}
	packet, err := FromBytes(buffer[:n])
	return packet, addr, err
}
