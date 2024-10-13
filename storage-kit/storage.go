package storage_kit

import (
	"time"

	"github.com/gocql/gocql"
)

// ScyllaDB struct to manage the session
type ScyllaStorage struct {
	Session *gocql.Session
}

// Initialize ScyllaDB connection
func NewScyllaStorage(contactPoint string) (*ScyllaStorage, error) {
	cluster := gocql.NewCluster(contactPoint)
	cluster.Keyspace = "rudp_keyspace"
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return &ScyllaStorage{Session: session}, nil
}

// Save packet data to ScyllaDB
func (s *ScyllaStorage) SavePacket(packetID int, timestampSent, timestampReceived time.Time, retried bool, retryCount int) error {
	query := `INSERT INTO packet_logs (primary_key_id, packet_id, timestamp_sent, timestamp_received, retried, retried_count)
	          VALUES (uuid(), ?, ?, ?, ?, ?)`
	return s.Session.Query(query, packetID, timestampSent, timestampReceived, retried, retryCount).Exec()
}

// Close ScyllaDB connection
func (s *ScyllaStorage) Close() {
	s.Session.Close()
}
