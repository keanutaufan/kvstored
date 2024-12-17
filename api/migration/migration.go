package main

import (
	"fmt"
	"log"

	"github.com/gocql/gocql"
)

type CassandraMigration struct {
	session *gocql.Session
}

func NewCassandraMigration(hosts []string) (*CassandraMigration, error) {
	// Create a cluster configuration
	cluster := gocql.NewCluster(hosts...)
	cluster.Consistency = gocql.Quorum

	// Create a session to the default keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("error creating Cassandra session: %v", err)
	}

	return &CassandraMigration{session: session}, nil
}

func (m *CassandraMigration) RunMigrations() error {
	// Create Keyspace
	err := m.createKeyspace()
	if err != nil {
		return fmt.Errorf("error creating keyspace: %v", err)
	}

	// Create Tables
	err = m.createTables()
	if err != nil {
		return fmt.Errorf("error creating tables: %v", err)
	}

	return nil
}

func (m *CassandraMigration) createKeyspace() error {
	query := `
		CREATE KEYSPACE IF NOT EXISTS kv_store_app 
		WITH replication = {
			'class': 'NetworkTopologyStrategy', 
			'replication_factor': 3
		}
	`
	return m.session.Query(query).Exec()
}

func (m *CassandraMigration) createTables() error {
	queries := []string{
		`
		CREATE TABLE IF NOT EXISTS kv_store_app.key_values (
			app_id text,
			key text,
			value text,
			created_at timestamp,
			PRIMARY KEY ((app_id), key)
		)
		`,
	}

	for _, query := range queries {
		if err := m.session.Query(query).Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (m *CassandraMigration) Close() {
	if m.session != nil {
		m.session.Close()
	}
}

func main() {
	migration, err := NewCassandraMigration([]string{"localhost:9042", "localhost:9043", "localhost:9044"})
	if err != nil {
		log.Fatalf("Failed to create migration: %v", err)
	}
	defer migration.Close()

	if err := migration.RunMigrations(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}
