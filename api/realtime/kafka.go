package realtime

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/keanutaufan/kvstored/api/entity"
	"github.com/segmentio/kafka-go"
)

type KeyChangeMessage struct {
	Type  string           `json:"type"` // "set", "update", "delete"
	AppID string           `json:"app_id"`
	Key   string           `json:"key"`
	Value *entity.KeyValue `json:"value,omitempty"`
}

type KafkaService struct {
	writer *kafka.Writer
	reader *kafka.Reader
}

func NewKafkaService(brokers []string, groupID string) *KafkaService {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   "kvstore",
	})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "kvstore",
		GroupID: groupID,
	})

	return &KafkaService{
		writer: writer,
		reader: reader,
	}
}

func (k *KafkaService) PublishKeyChange(msgType string, appID, key string, value *entity.KeyValue) error {
	msg := KeyChangeMessage{
		Type:  msgType,
		AppID: appID,
		Key:   key,
		Value: value,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return k.writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: payload,
			Time:  time.Now(),
		},
	)
}

func (k *KafkaService) StartConsumer(socketServer *SocketServer) {
	for {
		msg, err := k.reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading Kafka message: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var keyChange KeyChangeMessage
		if err := json.Unmarshal(msg.Value, &keyChange); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		switch keyChange.Type {
		case "set":
			if keyChange.Value != nil {
				socketServer.NotifyKeySet(*keyChange.Value)
			}
		case "update":
			if keyChange.Value != nil {
				socketServer.NotifyKeyUpdated(*keyChange.Value)
			}
		case "delete":
			socketServer.NotifyKeyDeleted(keyChange.AppID, keyChange.Key)
		}
	}
}

func (k *KafkaService) AsyncPublishKeyChange(msgType string, appID, key string, value *entity.KeyValue) {
	go func() {
		if err := k.PublishKeyChange(msgType, appID, key, value); err != nil {
			log.Printf("Error publishing Kafka message: %v", err)
		}
	}()
}

func (k *KafkaService) Close() error {
	if err := k.writer.Close(); err != nil {
		return err
	}
	return k.reader.Close()
}
