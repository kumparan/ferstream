package ferstream

import (
	"github.com/nats-io/nats.go"
)

type (
	// JetStream :nodoc:
	JetStream interface {
		Publish(subject string, value []byte) error
		QueueSubscribe(subj, queue string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error)
		Close() error
	}
)