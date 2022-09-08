package pulsarsdk

import (
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"log"
)

func test1Producer() {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: "pulsar://localhost:6650",
	})
	defer client.Close()

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: "my-topic",
	})
	_, err = producer.Send(context.Background(), &pulsar.ProducerMessage{
		Payload: []byte("hello"),
	})

	defer producer.Close()

	if err != nil {
		fmt.Println("Failed to publish message", err)
	} else {
		fmt.Println("Published message")
	}
}

func testSubscribe() {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: "pulsar://localhost:6650",
	})

	defer client.Close()

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            "my-topic",
		SubscriptionName: "my-sub",
		Type:             pulsar.Shared,
	})

	defer consumer.Close()

	msg, err := consumer.Receive(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Received message msgId: %#v -- content: '%s'\n",
		msg.ID(), string(msg.Payload()))
}

func test3() {
	client, err := pulsar.NewClient(pulsar.ClientOptions{URL: "pulsar://localhost:6650"})
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()
	reader, err := client.CreateReader(pulsar.ReaderOptions{
		Topic:          "topic-1",
		StartMessageID: pulsar.EarliestMessageID(),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	for reader.HasNext() {
		msg, err := reader.Next(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Received message msgId: %#v -- content: '%s'\n",
			msg.ID(), string(msg.Payload()))
	}
}
