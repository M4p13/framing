package framing

import (
	"context"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Middleware func(func(msgBody []byte)) func(msgBody []byte)

type JetstreamConsumer struct {
	Domain     string
	Stream     string
	Durable    string
	Handler    []func(msgBody []byte)
	Middleware []Middleware
}

type JetstreamManager struct {
	streamRegistry map[string]*jetstream.JetStream
	consumers      []*JetstreamConsumer
	nats           *nats.Conn
	middleware     []Middleware
}

func NewJetstreamManager(nc *nats.Conn) *JetstreamManager {
	streamRegistry := map[string]*jetstream.JetStream{}
	return &JetstreamManager{
		streamRegistry: streamRegistry,
		nats:           nc,
		middleware:     []Middleware{},
	}
}

func (jsm *JetstreamManager) Use(middleware Middleware) {
	jsm.middleware = append(jsm.middleware, middleware)
}

func (jsm *JetstreamManager) applyMiddleware(handler func(msgBody []byte), consumerMiddlewares []Middleware) func(msgBody []byte) {
	h := handler

	for i := len(consumerMiddlewares) - 1; i >= 0; i-- {
		h = consumerMiddlewares[i](h)
	}

	for i := len(jsm.middleware) - 1; i >= 0; i-- {
		h = jsm.middleware[i](h)
	}

	return h
}

func (jsm *JetstreamManager) GetDomain(domain string) (jetstream.JetStream, error) {
	stream, ok := jsm.streamRegistry[domain]
	if ok {
		return *stream, nil
	} else {
		js, err := jetstream.NewWithDomain(jsm.nats, domain)
		if err != nil {
			return nil, err
		}
		jsm.streamRegistry[domain] = &js
		return js, nil
	}
}

func (jsm *JetstreamManager) RegisterHandle(jc *JetstreamConsumer) error {
	jsm.consumers = append(jsm.consumers, jc)
	js, err := jsm.GetDomain(jc.Domain)
	if err != nil {
		return err
	}
	js.DeleteConsumer(context.Background(), jc.Stream, jc.Durable)
	c, err := js.CreateOrUpdateConsumer(context.Background(), jc.Stream, jetstream.ConsumerConfig{
		Durable: jc.Durable,
	})
	if err != nil {
		return err
	}
	c.Consume(func(msg jetstream.Msg) {
		data := msg.Data()
		for _, h := range jc.Handler {
			wrappedHandler := jsm.applyMiddleware(h, jc.Middleware)
			wrappedHandler(data)
		}
		msg.Ack()
	})
	return nil
}

func ErrorHandlingMiddleware(handler func(msgBody []byte)) func(msgBody []byte) {
	return func(msgBody []byte) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in message handler: %v", r)
			}
		}()
		handler(msgBody)
	}
}
