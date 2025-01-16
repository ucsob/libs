package consumer

import (
	"context"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"log"
)

type Group interface {
	sarama.ConsumerGroup
	Read(ctx context.Context, fn func(sarama.ConsumerGroupSession, *sarama.ConsumerMessage))
}

type group struct {
	sarama.ConsumerGroup
	topics []string
}

func NewConsumerGroup(client sarama.Client, topics []string, groupId string) (Group, error) {
	g, err := sarama.NewConsumerGroupFromClient(groupId, client)
	if err != nil {
		return nil, fmt.Errorf("sarama.NewConsumerGroupFromClient: %w", err)
	}

	return &group{
		ConsumerGroup: g,
		topics:        topics,
	}, nil
}

func (g *group) Read(ctx context.Context, fn func(sarama.ConsumerGroupSession, *sarama.ConsumerMessage)) {
	h := handler{
		fn:    fn,
		ready: make(chan bool),
	}

	go func() {
		for {
			if err := g.ConsumerGroup.Consume(ctx, g.topics, &h); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Panicf("Error from consumer: %v", err)
			}

			if ctx.Err() != nil {
				return
			}
			h.ready = make(chan bool)
		}
	}()

	<-h.ready
}
