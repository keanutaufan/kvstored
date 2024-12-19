package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	topicName := "kvstore"
	bootstrapServer := "kafka1:9092,kafka2:9092,kafka3:9092"
	replicationFactor := "3"
	partitions := "3"
	containerName := "kafka1"

	cmd := exec.Command("docker", "exec", containerName, "/opt/kafka/bin/kafka-topics.sh", "--create",
		"--topic", topicName,
		"--bootstrap-server", bootstrapServer,
		"--replication-factor", replicationFactor,
		"--partitions", partitions,
		"--config", "min.insync.replicas=2")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed to create Kafka topic:", err)
		return
	}

	fmt.Printf("Kafka topic %s created successfully with:\n", topicName)
	fmt.Printf("- Replication Factor: %s\n", replicationFactor)
	fmt.Printf("- Partitions: %s\n", partitions)
	fmt.Printf("- Min In-Sync Replicas: 2\n")
}
