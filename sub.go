package framing

import (
	"github.com/nats-io/nats.go"
)

type Subscription struct {
	Topic      string
	Handler    []func(msg *nats.Msg)
	Middleware []NatsMiddleware
}

func (s *Subscription) GetTopic() string {
	return s.Topic
}

func (s *Subscription) GetMiddleware() []NatsMiddleware {
	return s.Middleware
}

func (s *Subscription) GetHandlers() []func(msg *nats.Msg) {
	return s.Handler
}

type Responder struct {
	Topic      string
	Handler    []func(msg *nats.Msg)
	Middleware []NatsMiddleware
}

func (r *Responder) GetTopic() string {
	return r.Topic
}

func (r *Responder) GetMiddleware() []NatsMiddleware {
	return r.Middleware
}

func (r *Responder) GetHandlers() []func(msg *nats.Msg) {
	return r.Handler
}

type Subscriber interface {
	GetTopic() string
	GetMiddleware() []NatsMiddleware
	GetHandlers() []func(msg *nats.Msg)
}

type SubManager struct {
	Subscriber []*Subscriber
	middleware []NatsMiddleware
	nc         *nats.Conn
}

func NewSubManager(nc *nats.Conn) *SubManager {
	return &SubManager{
		Subscriber: []*Subscriber{},
		middleware: []NatsMiddleware{},
		nc:         nc,
	}
}

func (sm *SubManager) Use(middleware NatsMiddleware) {
	sm.middleware = append(sm.middleware, middleware)
}

func (sm *SubManager) applyMiddleware(handler func(msg *nats.Msg), consumerMiddlewares []NatsMiddleware) func(msg *nats.Msg) {
	h := handler

	for i := len(consumerMiddlewares) - 1; i >= 0; i-- {
		h = consumerMiddlewares[i](h)
	}

	for i := len(sm.middleware) - 1; i >= 0; i-- {
		h = sm.middleware[i](h)
	}

	return h
}

func (sm *SubManager) RegisterHandle(sub *Subscriber) error {
	s := *sub
	sm.Subscriber = append(sm.Subscriber, sub)
	topic := s.GetTopic()
	mw := s.GetMiddleware()
	for _, h := range s.GetHandlers() {
		wrappedHandler := sm.applyMiddleware(h, mw)
		sm.nc.Subscribe(topic, wrappedHandler)
	}
	return nil
}
