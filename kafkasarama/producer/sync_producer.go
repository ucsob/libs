package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
)

type SyncProducer interface {
	sarama.SyncProducer
	Write(ctx context.Context, topic string, value any, key ...any) (int32, int64, error)
}

type syncProducer struct {
	sarama.SyncProducer
}

func NewSyncProducer(client sarama.Client) (SyncProducer, error) {
	client.Config().Producer.Return.Successes = true

	pr, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, fmt.Errorf("sarama.NewSyncProducerFromClient: %w", err)
	}

	return &syncProducer{pr}, nil
}

func (p *syncProducer) Write(ctx context.Context, topic string, value any, key ...any) (int32, int64, error) {
	select {
	case <-ctx.Done():
		return 0, 0, nil
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return 0, 0, err
		}

		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(data),
		}

		if len(key) > 0 && key[0] != "" {
			data, _ = json.Marshal(key[0])
			msg.Key = sarama.StringEncoder(data)
		}

		return p.SendMessage(msg)
	}
}
