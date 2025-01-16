package main

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"libs/kafkasarama"
	"libs/kafkasarama/consumer"
	"libs/kafkasarama/producer"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var brokers = []string{"localhost:19092"}
var topics = []string{"topic1"}

const consumerGroupId = "KAFKA_SARAMA"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	saramaCfg := sarama.NewConfig()
	client, err := kafkasarama.New(brokers, saramaCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	type tmp struct {
		Id int
		Tp string
	}

	// Write via Async Producer.
	saramaCfg.Producer.Return.Errors = true
	saramaCfg.Producer.Flush.Frequency = 20 * time.Millisecond
	asyncProducer, err := producer.NewAsyncProducer(client)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = asyncProducer.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	for i := range topics {
		t := tmp{
			Id: i,
			Tp: "async",
		}
		asyncProducer.Write(ctx, topics[i], t, i)
	}

	// Create consumer group.
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	group, err := consumer.NewConsumerGroup(client, topics, consumerGroupId)
	if err != nil {
		log.Fatal(err)
	}

	// Read messages.
	group.Read(ctx, func(session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) {
		fmt.Printf("key: %s, value: %s\n", string(message.Key), string(message.Value))
		session.MarkMessage(message, "")
	})
	defer func() {
		if err := group.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("🎉 Successfully started 🎉")

	go func() {
		for err = range asyncProducer.Errors() {
			log.Println(err)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	cancel()

	log.Print("🛑 Successfully shutdown 🛑")
}
