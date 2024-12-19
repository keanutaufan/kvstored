package db

import (
	"time"

	"github.com/gocql/gocql"
)

type CassandraClient struct {
	Session *gocql.Session
}

func NewCassandraClient(hosts []string) (*CassandraClient, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = "kv_store_app"
	cluster.Consistency = gocql.Quorum
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
	cluster.ReconnectInterval = 1 * time.Second
	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return &CassandraClient{Session: session}, nil
}
