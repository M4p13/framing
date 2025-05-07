package framing

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type (
	Middleware     func(func(msgBody []byte)) func(msgBody []byte)
	NatsMiddleware func(func(msg *nats.Msg)) func(msg *nats.Msg)
)

type NatsManager struct {
	nc               *nats.Conn
	JetstreamManager *JetstreamManager
	SubManager       *SubManager
	AppIdentity      AppIdentity
}

func (n *NatsManager) GetAppId() string {
	return fmt.Sprintf("%s.%s", n.AppIdentity.Name, n.AppIdentity.UUID.String())
}

func (n *NatsManager) GetAppTopic() string {
	return fmt.Sprintf("%s.system.application.%s", n.AppIdentity.AppDomain, n.GetAppId())
}

func (n *NatsManager) Write(p []byte) (int, error) {
	topic := n.GetAppTopic()
	n.nc.Publish(topic+".log", p)
	return len(p), nil
}

func NewNatsManager(nc *nats.Conn, ai AppIdentity) *NatsManager {
	return &NatsManager{
		nc:               nc,
		JetstreamManager: NewJetstreamManager(nc),
		SubManager:       NewSubManager(nc),
		AppIdentity:      ai,
	}
}

func (n *NatsManager) StartAlive() {
	topic := n.GetAppTopic()
	go func() {
		for {
			time.Sleep(time.Second * 3)
			n.nc.Publish(topic+".alive", []byte{1})
		}
	}()
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
