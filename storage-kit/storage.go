package storage_kit

import (
	"fmt"
	"os"
	"time"

	"strings"

	"github.com/gocql/gocql"
)

// ScyllaDB struct to manage the session
type ScyllaStorage struct {
	Session *gocql.Session
}

// Initialize ScyllaDB connection and load schema
func NewScyllaStorage(contactPoint string) (*ScyllaStorage, error) {
	cluster := gocql.NewCluster(contactPoint)
	cluster.Keyspace = "rudp_keyspace"
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	storage := &ScyllaStorage{Session: session}

	// Load and execute the schema from the CQL file
	err = storage.loadSchema("schema/schema.cql")
	if err != nil {
		return nil, fmt.Errorf("failed to load schema: %v", err)
	}

	return storage, nil
}

// Function to load and execute the schema from a .cql file
func (s *ScyllaStorage) loadSchema(filePath string) error {
	// Read the .cql file
	schemaData, err := os.ReadFile(filePath) // Replaces ioutil.ReadFile
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Split the schema into individual CQL statements and execute them
	queries := strings.Split(string(schemaData), ";")
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query != "" {
			if err := s.Session.Query(query).Exec(); err != nil {
				return fmt.Errorf("failed to execute schema query: %w", err)
			}
		}
	}

	fmt.Println("Successfully loaded and executed CQL schema from:", filePath)
	return nil
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
