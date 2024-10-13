package storage_kit

import (
	"time"
)

type PacketRepository struct {
	Storage *ScyllaStorage
}

// Initialize repository
func NewPacketRepository(storage *ScyllaStorage) *PacketRepository {
	return &PacketRepository{Storage: storage}
}

// Save a packet with sent and received timestamps
func (r *PacketRepository) SavePacket(packetID int, timestampSent time.Time, retryCount int, retried bool) error {
	timestampReceived := time.Now()
	return r.Storage.SavePacket(packetID, timestampSent, timestampReceived, retried, retryCount)
}
