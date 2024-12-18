package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	topicName := "kvstore"
	bootstrapServer := "localhost:9092"
	replicationFactor := "1"
	partitions := "1"
	containerName := "kvstored-kafka-1"

	cmd := exec.Command("docker", "exec", containerName, "/opt/kafka/bin/kafka-topics.sh", "--create",
		"--topic", topicName,
		"--bootstrap-server", bootstrapServer,
		"--replication-factor", replicationFactor,
		"--partitions", partitions)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed to create Kafka topic:", err)
		return
	}

	fmt.Println("Kafka topic", topicName, "created successfully.")
}
