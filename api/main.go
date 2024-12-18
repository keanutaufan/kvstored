package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/keanutaufan/kvstored/api/controller"
	"github.com/keanutaufan/kvstored/api/db"
	"github.com/keanutaufan/kvstored/api/realtime"
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

	defer cassandraClient.Session.Close()

	server := gin.Default()
	socketServer := realtime.NewSocketServer()

	go socketServer.Server.Serve()
	defer socketServer.Server.Close()

	keyValueController := controller.NewKeyValueController(keyValueService, socketServer)

	routes.KeyValueRoutes(server, keyValueController)

	server.GET("/socket.io/*any", gin.WrapH(socketServer.Server))
	server.POST("/socket.io/*any", gin.WrapH(socketServer.Server))

	server.Run(":8000")
}
