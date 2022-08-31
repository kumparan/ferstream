package ferstream

import (
	"time"

	"github.com/kumparan/go-utils"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type (
	// JetStream :nodoc:
	JetStream interface {
		Publish(subject string, value []byte, opts ...nats.PubOpt) (*nats.PubAck, error)
		QueueSubscribe(subj, queue string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error)
		Close()
		AddStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error)
		ConsumerInfo(streamName, consumerName string, opts ...nats.JSOpt) (*nats.ConsumerInfo, error)
	}

	// jsImpl JetStream implementation
	jsImpl struct {
		natsConn *nats.Conn
		jsCtx    nats.JetStreamContext
	}

	// JetStreamRegistrar
	JetStreamRegistrar interface {
		RegisterNATSJetStream(js JetStream)
	}

	// StreamRegistration
	StreamRegistrar interface {
		InitStream() error
	}

	// Subscriber
	Subscriber interface {
		SubscribeJetStreamEvent() error
	}

	// Message Handler :nodoc:
	MessageHandler func(payload MessageParser) (err error)
)

// NewNATSConnection :nodoc:
func NewNATSConnection(url string, natsOpts ...nats.Option) (JetStream, error) {
	nc, err := connect(url, natsOpts...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url": url,
		}).Error(err)
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	impl := &jsImpl{
		natsConn: nc,
		jsCtx:    js,
	}

	return impl, nil
}

// RegisterJetStreamClient provide jetstream instance, stream, and subscription registration
func RegisterJetStreamClient(js JetStream, clients []JetStreamRegistrar) error {
	for _, client := range clients {
		client.RegisterNATSJetStream(js)
		if streamRegistrar, ok := client.(StreamRegistrar); ok {
			err := streamRegistrar.InitStream()
			if err != nil {
				return err
			}
		}
		if subscriber, ok := client.(Subscriber); ok {
			err := subscriber.SubscribeJetStreamEvent()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NewNATSMessageHandler a wrapper to standardize how we handle NATS messages.
// Payload (arg 0) should always be empty when the method is called. The payload data will later parse data from msg.Data.
func NewNATSMessageHandler(payload MessageParser, retryAttempts int, retryInterval time.Duration, msgHandler MessageHandler, errHandler MessageHandler) nats.MsgHandler {
	return func(msg *nats.Msg) {
		logger := logrus.WithField("msg", utils.Dump(msg))
		defer func(logger *logrus.Entry) {
			err := msg.Ack()
			if err != nil {
				logger.Error(err)
			}
		}(logger)

		if msg.Data == nil {
			logger.Error("Message payload is nil")
			return
		}

		err := payload.ParseFromBytes(msg.Data)
		if err != nil {
			logger.WithField("error-detail", err).Error("Unmarshal failed")
			return
		}

		payload.AddSubject(msg.Subject)

		retryErr := utils.Retry(retryAttempts, retryInterval, func() error {
			return msgHandler(payload)
		})
		if retryErr == nil {
			return
		}

		logger.WithFields(logrus.Fields{
			"payload": utils.Dump(payload),
			"cause":   ErrGiveUpProcessingMessagePayload,
		}).Error(retryErr)

		if errHandler == nil {
			return
		}

		// hand over to error handler
		logrus.WithField("payload", utils.Dump(payload)).Warnf("handling ErrGiveUpProcessingMessagePayload")
		err = errHandler(payload)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"payload": utils.Dump(payload),
				"cause":   err.Error(),
			}).Error(err)
		}
	}
}

// connect to nats streaming
func connect(url string, options ...nats.Option) (*nats.Conn, error) {
	nc, err := nats.Connect(url, options...)
	if err != nil {
		return nil, err
	}
	return nc, nil
}

func (j *jsImpl) checkConnIsValid() (b bool) {
	return j.natsConn != nil && j.natsConn.IsConnected()
}

// Publish publish message using JetStream
func (j *jsImpl) Publish(subject string, value []byte, opts ...nats.PubOpt) (*nats.PubAck, error) {
	if !j.checkConnIsValid() {
		return nil, ErrConnectionLost
	}
	return j.jsCtx.Publish(subject, value, opts...)
}

// QueueSubscribe :nodoc:
func (j *jsImpl) QueueSubscribe(subj, queue string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	if !j.checkConnIsValid() {
		return nil, ErrConnectionLost
	}
	return j.jsCtx.QueueSubscribe(subj, queue, cb, opts...)
}

// Close close NATS connection
func (j *jsImpl) Close() {
	if j.checkConnIsValid() {
		j.natsConn.Close()
	}
}

// AddStream add stream
func (j *jsImpl) AddStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error) {
	if !j.checkConnIsValid() {
		return nil, ErrConnectionLost
	}

	streamInfo, _ := j.jsCtx.StreamInfo(cfg.Name)

	if streamInfo == nil {
		return j.jsCtx.AddStream(cfg, opts...)
	}

	return j.jsCtx.UpdateStream(cfg)

}

// ConsumerInfo
func (j *jsImpl) ConsumerInfo(streamName, consumerName string, opts ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	if !j.checkConnIsValid() {
		return nil, ErrConnectionLost
	}

	return j.jsCtx.ConsumerInfo(streamName, consumerName, opts...)
}
