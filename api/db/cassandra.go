package db

import "github.com/gocql/gocql"

type CassandraClient struct {
	Session *gocql.Session
}

func NewCassandraClient(hosts []string) (*CassandraClient, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = "kv_store_app"
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return &CassandraClient{Session: session}, nil
}
