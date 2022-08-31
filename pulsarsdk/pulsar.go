package pulsarsdk

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
)

var (
	cli Client
)

type Client struct {
	pulsar.Client
}

func GetClient() Client {
	return cli
}

type Subscriber func(message pulsar.Message, err error)

func Init(option pulsar.ClientOptions) (err error) {
	client, err := pulsar.NewClient(option)
	if err != nil {
		return
	}
	cli = Client{
		client,
	}
	return
}

// 生产消息
func ProductMsg(ctx context.Context, msg []byte, topic string) (msgID pulsar.MessageID, err error) {
	prod, err := cli.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})
	if err != nil {
		return
	}
	msgID, err = prod.Send(ctx, &pulsar.ProducerMessage{
		Payload: msg,
	})
	return
}

func SubscribeMsg(ctx context.Context, topic, SubscriptionName string, callBack Subscriber) {
	consumer, err := cli.Subscribe(pulsar.ConsumerOptions{
		Topic:            "my-topic",
		SubscriptionName: "my-sub",
		Type:             pulsar.Shared,
	})
	defer consumer.Close()
	msg, err := consumer.Receive(ctx)
	if err != nil {
		callBack(nil, err)
		return
	}
	callBack(msg, nil)
}

func Closed() {
	cli.Close()
}
