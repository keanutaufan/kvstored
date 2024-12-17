package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/keanutaufan/kvstored/api/controller"
	"github.com/keanutaufan/kvstored/api/db"
	"github.com/keanutaufan/kvstored/api/repository"
	"github.com/keanutaufan/kvstored/api/routes"
	"github.com/keanutaufan/kvstored/api/service"
)

func main() {
	cassandraClient, err := db.NewCassandraClient([]string{"127.0.0.1:9042", "127.0.0.1:9043", "127.0.0.1:9044"})
	if err != nil {
		log.Fatalf("Failed to create Cassandra client: %v", err)
	}

	keyValueRepository := repository.NewKeyValueRepository(cassandraClient)
	keyValueService := service.NewKeyValueService(keyValueRepository)
	keyValueController := controller.NewKeyValueController(keyValueService)

	defer cassandraClient.Session.Close()

	server := gin.Default()

	routes.KeyValueRoutes(server, keyValueController)

	server.Run(":8000")
}
