package ferstream

import (
	"fmt"
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
		Subscribe(subj string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error)
		AddStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error)
		ConsumerInfo(streamName, consumerName string, opts ...nats.JSOpt) (*nats.ConsumerInfo, error)
		GetNATSConnection() *nats.Conn
	}

	// jsImpl JetStream implementation
	jsImpl struct {
		natsConn *nats.Conn
		jsCtx    nats.JetStreamContext
	}

	// JetStreamRegistrar :nodoc:
	JetStreamRegistrar interface {
		RegisterNATSJetStream(js JetStream)
	}

	// StreamRegistrar :nodoc:
	StreamRegistrar interface {
		InitStream() error
	}

	// Subscriber :nodoc:
	Subscriber interface {
		SubscribeJetStreamEvent() error
	}

	// MessageHandler :nodoc:
	MessageHandler func(payload MessageParser) (err error)
)

// GetNATSConnection :nodoc:
func (j *jsImpl) GetNATSConnection() *nats.Conn {
	if j == nil {
		return nil
	}
	return j.natsConn
}

// Publish publish message using JetStream
func (j *jsImpl) Publish(subject string, value []byte, opts ...nats.PubOpt) (*nats.PubAck, error) {
	if !j.isValidConn() {
		return nil, ErrConnectionLost
	}
	return j.jsCtx.Publish(subject, value, opts...)
}

// QueueSubscribe :nodoc:
func (j *jsImpl) QueueSubscribe(subj, queue string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	if !j.isValidConn() {
		return nil, ErrConnectionLost
	}
	return j.jsCtx.QueueSubscribe(subj, queue, cb, opts...)
}

// Subscribe :nodoc:
func (j *jsImpl) Subscribe(subj string, cb nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	if !j.isValidConn() {
		return nil, ErrConnectionLost
	}
	return j.jsCtx.Subscribe(subj, cb, opts...)
}

// AddStream add stream
func (j *jsImpl) AddStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) (*nats.StreamInfo, error) {
	if !j.isValidConn() {
		return nil, ErrConnectionLost
	}

	streamInfo, _ := j.jsCtx.StreamInfo(cfg.Name)

	if streamInfo == nil {
		return j.jsCtx.AddStream(cfg, opts...)
	}

	return j.jsCtx.UpdateStream(cfg)

}

// ConsumerInfo :nodoc:
func (j *jsImpl) ConsumerInfo(streamName, consumerName string, opts ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	if !j.isValidConn() {
		return nil, ErrConnectionLost
	}

	return j.jsCtx.ConsumerInfo(streamName, consumerName, opts...)
}

func (j *jsImpl) isValidConn() (b bool) {
	return j.natsConn != nil && j.natsConn.IsConnected()
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
			logger.Error("message payload is nil")
			return
		}

		err := payload.ParseFromBytes(msg.Data)
		if err != nil {
			logger.WithField("error-detail", err).Error("unmarshal failed")
			return
		}

		payload.AddSubject(msg.Subject)
		defer logger.WithField("payload", utils.Dump(payload)).Warn("message payload")

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

// SafeClose :nodoc:
func SafeClose(js JetStream) {
	if js == nil {
		return
	}
	natsConn := js.GetNATSConnection()
	if natsConn == nil {
		return
	}
	if !natsConn.IsConnected() {
		return
	}
	err := natsConn.Drain()
	if err != nil {
		logrus.Errorf("draining connection error. reason: %q\n", err)
		logrus.Info("force closing...")
		natsConn.Close()
	}
}

// NewNATSConnection :nodoc:
func NewNATSConnection(NATSJSHost string, clients []JetStreamRegistrar, natsOpts ...nats.Option) (JetStream, error) {
	opts := []nats.Option{
		nats.UseOldRequestStyle(),
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			logrus.Errorf("NATS got error! reason: %q\n", err)
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			logrus.Errorf("NATS got disconnected! reason: %q\n", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			_, err := initJetStreamClients(nc, clients)
			if err != nil {
				logrus.Errorf("NATS failed to reconnect. reason: %q\n", err)
				return
			}

			fmt.Printf("NATS got reconnected to %q\n", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logrus.Errorf("NATS connection closed. reason: %q\n", nc.LastError())
		}),
	}

	natsOpts = append(natsOpts, opts...)

	nc, err := nats.Connect(NATSJSHost, natsOpts...)
	if err != nil {
		logrus.Errorf("NATS failed to connect. reason: %q\n", err)
		return nil, err
	}

	return initJetStreamClients(nc, clients)
}

// registerJetStreamClient provide jetstream instance, stream, and subscription registration
func registerJetStreamClient(js JetStream, clients []JetStreamRegistrar) error {
	for _, client := range clients {
		client.RegisterNATSJetStream(js)
	}

	for _, client := range clients {
		if streamRegistrar, ok := client.(StreamRegistrar); ok {
			err := streamRegistrar.InitStream()
			if err != nil {
				logrus.WithField("client", fmt.Sprintf("%T", client)).Error(err)
				return err
			}
		}

		if subscriber, ok := client.(Subscriber); ok {
			err := subscriber.SubscribeJetStreamEvent()
			if err != nil {
				logrus.WithField("client", fmt.Sprintf("%T", client)).Error(err)
				return err
			}
		}
	}

	return nil
}

func initJetStreamClients(nc *nats.Conn, clients []JetStreamRegistrar) (JetStream, error) {
	jsCtx, err := nc.JetStream()
	if err != nil {
		logrus.Errorf("failed to get jetstream context. reason: %q\n", err)
		return nil, err
	}

	js := &jsImpl{
		natsConn: nc,
		jsCtx:    jsCtx,
	}

	err = registerJetStreamClient(js, clients)
	if err != nil {
		logrus.Errorf("failed to register jetstream client. reason %q\n", err)
		return nil, err
	}

	return js, nil
}
