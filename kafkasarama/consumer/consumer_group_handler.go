package consumer

import (
	"errors"
	"github.com/IBM/sarama"
)

type handler struct {
	fn    func(sarama.ConsumerGroupSession, *sarama.ConsumerMessage)
	ready chan bool
}

func (h *handler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *handler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return errors.New("message channel was closed")
			}

			h.fn(session, message)
		case <-session.Context().Done():
			return nil
		}
	}
}
