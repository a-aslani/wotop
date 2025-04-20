package pubsub

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"time"
)

type EventData struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Payload Payload `json:"payload"`
}

type Payload interface{}

type Event struct {
	conn     *Connection
	producer *producer
	consumer *Consumer
	appName  string
}

func newConnection(appName, username, password, host, vhost string) (*Connection, error) {
	return NewConnection(fmt.Sprintf("%s-connection", appName), ConnectionOptions{
		URI: fmt.Sprintf("amqp://%s:%s@%s/%s", username, password, host, vhost),
		Config: amqp.Config{
			Dial:      amqp.DefaultDial(time.Second),
			Heartbeat: time.Second,
		},
		LazyConnection: mo.Some(true),
	})
}

func NewEvent(appName, username, password, host, vhost string) (*Event, error) {

	event := &Event{}

	event.appName = appName

	conn, err := newConnection(appName, username, password, host, vhost)
	if err != nil {
		return nil, err
	}

	event.conn = conn

	producer := NewProducer(event.conn, fmt.Sprintf("%s-producer", appName), ProducerOptions{
		Exchange: ProducerOptionsExchange{
			Name: mo.Some(fmt.Sprintf("%s.event", appName)),
			Kind: mo.Some(ExchangeKindTopic),
		},
	})

	event.producer = producer

	return event, nil
}

func (e *Event) SetConsumer(queueName string, bindings []ConsumerOptionsBinding) {
	e.consumer = NewConsumer(e.conn, fmt.Sprintf("%s-consumer", e.appName), ConsumerOptions{
		Queue: ConsumerOptionsQueue{
			Name: queueName,
		},
		Bindings: bindings,
		Message: ConsumerOptionsMessage{
			PrefetchCount: mo.Some(1000),
		},
		EnableDeadLetter: mo.Some(true),
	})
}

func (e *Event) Publish(eventName string, payload Payload) error {

	body, _ := json.Marshal(EventData{
		ID:      uuid.NewString(),
		Name:    eventName,
		Payload: payload,
	})

	return e.producer.Publish(eventName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (e *Event) Consume(msg func(int64, *amqp.Delivery)) {

	channel := e.consumer.Consume()

	var i int64 = 0
	for m := range channel {
		lo.Try0(func() { // handle exceptions
			msg(i, m)
		})
		i++
	}
}
