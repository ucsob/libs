package kafkasarama

import (
	"fmt"
	"github.com/IBM/sarama"
)

type Client interface {
	sarama.Client
	IsHealthy() bool
}

type client struct {
	sarama.Client
}

func New(brokers []string, cfg *sarama.Config) (Client, error) {
	c, err := sarama.NewClient(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("sarama.NewClient: %w", err)
	}

	return &client{c}, nil
}

func (c *client) IsHealthy() bool {
	return len(c.Brokers()) > 0
}
