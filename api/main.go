package main

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

// KeyValue represents the structure of our key-value store
type KeyValue struct {
	AppID     string    `json:"app_id" binding:"required"`
	Key       string    `json:"key" binding:"required"`
	Value     string    `json:"value" binding:"required"`
	Timestamp time.Time `json:"timestamp"`
}

// CassandraClient manages Cassandra connection and operations
type CassandraClient struct {
	session *gocql.Session
}

// NewCassandraClient creates a new Cassandra client
func NewCassandraClient(hosts []string) (*CassandraClient, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = "kv_store_app"
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return &CassandraClient{session: session}, nil
}

// Set stores a key-value pair scoped to an app_id
func (c *CassandraClient) Set(appID, key, value string) error {
	// Validate input
	if strings.TrimSpace(appID) == "" || strings.TrimSpace(key) == "" {
		return errors.New("app_id and key cannot be empty")
	}

	return c.session.Query(`
		INSERT INTO kv_store_app.key_values (app_id, key, value, created_at) 
		VALUES (?, ?, ?, ?)
	`, appID, key, value, time.Now()).Exec()
}

// Get retrieves a value by app_id and key
func (c *CassandraClient) Get(appID, key string) (string, error) {
	var value string
	err := c.session.Query(`
		SELECT value FROM kv_store_app.key_values 
		WHERE app_id = ? AND key = ?
	`, appID, key).Scan(&value)

	if err == gocql.ErrNotFound {
		return "", errors.New("key not found for the given app")
	}
	return value, err
}

func (c *CassandraClient) Update(appID, key, value string) error {
	// Validate input
	if strings.TrimSpace(appID) == "" || strings.TrimSpace(key) == "" {
		return errors.New("app_id and key cannot be empty")
	}

	// Check if the key exists before updating
	var existing string
	err := c.session.Query(`
		SELECT value FROM kv_store_app.key_values 
		WHERE app_id = ? AND key = ?
	`, appID, key).Scan(&existing)

	if err == gocql.ErrNotFound {
		return errors.New("key not found for the given app")
	} else if err != nil {
		return err
	}

	// Perform the update
	return c.session.Query(`
		UPDATE kv_store_app.key_values 
		SET value = ?, created_at = ?
		WHERE app_id = ? AND key = ?
	`, value, time.Now(), appID, key).Exec()
}

// Delete removes a key-value pair
func (c *CassandraClient) Delete(appID, key string) error {
	return c.session.Query(`
		DELETE FROM kv_store_app.key_values 
		WHERE app_id = ? AND key = ?
	`, appID, key).Exec()
}

func main() {
	// Initialize Cassandra client
	cassandraClient, err := NewCassandraClient([]string{"localhost"})
	if err != nil {
		log.Fatalf("Failed to create Cassandra client: %v", err)
	}
	defer cassandraClient.session.Close()

	// Setup Gin router
	r := gin.Default()

	// Key-Value endpoints
	r.POST("/kv", func(c *gin.Context) {
		var kv KeyValue
		if err := c.BindJSON(&kv); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if err := cassandraClient.Set(kv.AppID, kv.Key, kv.Value); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "success"})
	})

	r.GET("/kv/:app_id/:key", func(c *gin.Context) {
		appID := c.Param("app_id")
		key := c.Param("key")

		value, err := cassandraClient.Get(appID, key)
		if err != nil {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"app_id": appID,
			"key":    key,
			"value":  value,
		})
	})

	r.PUT("/kv", func(c *gin.Context) {
		var kv KeyValue
		if err := c.BindJSON(&kv); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if err := cassandraClient.Update(kv.AppID, kv.Key, kv.Value); err != nil {
			if strings.Contains(err.Error(), "key not found") {
				c.JSON(404, gin.H{"error": err.Error()})
			} else {
				c.JSON(500, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(200, gin.H{"status": "updated"})
	})

	r.DELETE("/kv/:app_id/:key", func(c *gin.Context) {
		appID := c.Param("app_id")
		key := c.Param("key")

		if err := cassandraClient.Delete(appID, key); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "deleted"})
	})

	// Start server
	r.Run(":8000")
}
