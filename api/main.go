package main

import (
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/keanutaufan/kvstored/api/controller"
	"github.com/keanutaufan/kvstored/api/db"
	"github.com/keanutaufan/kvstored/api/realtime"
	"github.com/keanutaufan/kvstored/api/repository"
	"github.com/keanutaufan/kvstored/api/routes"
	"github.com/keanutaufan/kvstored/api/service"
	"github.com/keanutaufan/kvstored/api/utils"
)

func main() {
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			panic(err)
		}
	}

	cqlHosts := utils.LoadEnv("CASSANDRA_HOSTS", "localhost")
	cassandraClient, err := db.NewCassandraClient(strings.Split(cqlHosts, ","))
	if err != nil {
		log.Fatalf("Failed to create Cassandra client: %v", err)
	}

	keyValueRepository := repository.NewKeyValueRepository(cassandraClient)
	keyValueService := service.NewKeyValueService(keyValueRepository)

	defer cassandraClient.Session.Close()

	nodeId := utils.LoadEnv("NODE_ID", "kvstored1")
	kafkaHosts := utils.LoadEnv("KAFKA_HOSTS", "localhost")
	kafkaService := realtime.NewKafkaService(strings.Split(kafkaHosts, ","), "kvstored-group-"+nodeId)
	defer kafkaService.Close()

	socketServer := realtime.NewSocketServer()
	go socketServer.Server.Serve()
	defer socketServer.Server.Close()

	go kafkaService.StartConsumer(socketServer)

	keyValueController := controller.NewKeyValueController(keyValueService, kafkaService)

	server := gin.Default()
	routes.KeyValueRoutes(server, keyValueController)

	server.GET("/socket.io/*any", gin.WrapH(socketServer.Server))
	server.POST("/socket.io/*any", gin.WrapH(socketServer.Server))

	port := utils.LoadEnv("PORT", "8000")
	server.Run(":" + port)
}
