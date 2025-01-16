package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
)

type AsyncProducer interface {
	sarama.AsyncProducer
	Write(ctx context.Context, topic string, value any, key ...any)
}

type asyncProducer struct {
	sarama.AsyncProducer
}

func NewAsyncProducer(client sarama.Client) (AsyncProducer, error) {
	pr, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		return nil, fmt.Errorf("sarama.NewAsyncProducerFromClient: %w", err)
	}

	return &asyncProducer{pr}, nil
}

func (p *asyncProducer) Write(ctx context.Context, topic string, value any, key ...any) {
	select {
	case <-ctx.Done():
		p.AsyncProducer.AsyncClose()
		return
	default:
		data, _ := json.Marshal(value)

		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(data),
		}

		if len(key) > 0 && key[0] != "" {
			data, _ = json.Marshal(key[0])
			msg.Key = sarama.StringEncoder(data)
		}

		p.Input() <- msg
	}
}
