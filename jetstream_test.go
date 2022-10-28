package ferstream

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultURL = "nats://127.0.0.1:4222"

func RunBasicJetStreamServer() *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.JetStream = true
	opts.StoreDir = ".jetstream_log"
	return natsserver.RunServer(&opts)
}

func TestMain(t *testing.M) {
	srv := RunBasicJetStreamServer()
	defer srv.Shutdown()
	m := t.Run()
	os.Exit(m)
}

func TestPublish(t *testing.T) {
	natsOpts := []nats.Option{
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, e error) {
			t.Fatalf("Error euy %q", e)
		}),
	}

	n, err := NewNATSConnection(defaultURL, natsOpts...)
	require.NoError(t, err)
	defer SafeClose(n)

	streamConf := &nats.StreamConfig{
		Name:      "STREAM_NAME_PUBLISH",
		Subjects:  []string{"STREAM_NAME_PUBLISH.*"},
		Storage:   nats.FileStorage,
		Retention: nats.WorkQueuePolicy,
	}

	_, err = n.AddStream(streamConf)
	require.NoError(t, err)

	_, err = n.Publish("STREAM_NAME_PUBLISH.TEST", []byte("test"))
	require.NoError(t, err)
}

func TestQueueSubscribe(t *testing.T) {
	t.Run("queue subscribe NatsEventMessage", func(t *testing.T) {
		n, err := NewNATSConnection(defaultURL)
		require.NoError(t, err)

		defer SafeClose(n)

		streamConf := &nats.StreamConfig{
			Name:      "STREAM_NAME_QUEUE_SUBSCRIBE",
			Subjects:  []string{"STREAM_NAME_QUEUE_SUBSCRIBE.*"},
			Storage:   nats.FileStorage,
			Retention: nats.WorkQueuePolicy,
		}

		_, err = n.AddStream(streamConf)
		require.NoError(t, err)

		countMsg := 10
		subject := "STREAM_NAME_QUEUE_SUBSCRIBE.TEST"
		queue := "test_queue_group"

		msgBytes, err := NewNatsEventMessage().WithEvent(&NatsEvent{
			ID:     int64(1232),
			UserID: int64(21),
		}).Build()

		require.NoError(t, err)

		for i := 0; i < countMsg; i++ {
			_, err = n.Publish(subject, msgBytes)
			require.NoError(t, err)
		}

		receiverCh := make(chan *nats.Msg)
		sub, err := n.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
			receiverCh <- msg
		})
		require.NoError(t, err)

		for i := 0; i < countMsg; i++ {
			b := <-receiverCh

			assert.Equal(t, msgBytes, b.Data)
			assert.Equal(t, subject, b.Subject, "test subject")
		}

		_ = sub.Unsubscribe()
	})

	t.Run("queue subscribe NatsEventAuditLogMessage", func(t *testing.T) {
		n, err := NewNATSConnection(defaultURL)
		require.NoError(t, err)
		defer SafeClose(n)

		streamConf := &nats.StreamConfig{
			Name:      "STREAM_NAME_AUDIT",
			Subjects:  []string{"STREAM_NAME_AUDIT.*"},
			Storage:   nats.FileStorage,
			Retention: nats.WorkQueuePolicy,
		}

		_, err = n.AddStream(streamConf)
		require.NoError(t, err)

		countMsg := 10
		subject := "STREAM_NAME_AUDIT.TEST_NATS_EVENT_AUDIT_LOG_MESSAGE"
		queue := "test_queue_group"

		type User struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		}

		oldData := User{
			ID:   int64(123),
			Name: "test name",
		}

		newData := User{
			ID:   int64(123),
			Name: "new test name",
		}

		byteOldData, err := json.Marshal(oldData)
		require.NoError(t, err)
		byteNewData, err := json.Marshal(newData)
		require.NoError(t, err)

		createdAt, err := time.Parse("2006-01-02", "2020-01-29")
		require.NoError(t, err)

		msg := &NatsEventAuditLogMessage{
			ServiceName:    "test-audit",
			UserID:         123,
			AuditableType:  "user",
			AuditableID:    "123",
			Action:         "update",
			AuditedChanges: string(byteNewData),
			OldData:        string(byteOldData),
			NewData:        string(byteNewData),
			CreatedAt:      createdAt,
			Error:          nil,
		}
		msgBytes, err := msg.Build()
		require.NoError(t, err)

		for i := 0; i < countMsg; i++ {
			_, err = n.Publish(subject, msgBytes)
			require.NoError(t, err)

		}

		receiverCh := make(chan *nats.Msg)
		sub, err := n.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
			receiverCh <- msg
		})
		require.NoError(t, err)

		for i := 0; i < countMsg; i++ {
			b := <-receiverCh

			assert.Equal(t, msgBytes, b.Data)
			assert.Equal(t, subject, b.Subject, "test subject")
		}

		_ = sub.Unsubscribe()
	})
}

func TestSubscribe(t *testing.T) {
	t.Run("subscribe NatsEventMessage", func(t *testing.T) {
		n, err := NewNATSConnection(defaultURL)
		require.NoError(t, err)

		n2, err := NewNATSConnection(defaultURL)
		require.NoError(t, err)
		defer SafeClose(n)
		defer SafeClose(n2)

		streamConf := &nats.StreamConfig{
			Name:     "STREAM_NAME_SUBSCRIBE",
			Subjects: []string{"STREAM_NAME_SUBSCRIBE.*"},
			Storage:  nats.FileStorage,
		}

		_, err = n.AddStream(streamConf)
		require.NoError(t, err)

		countMsg := 10
		subject := "STREAM_NAME_SUBSCRIBE.TEST"

		msgBytes, err := NewNatsEventMessage().WithEvent(&NatsEvent{
			ID:     int64(1232),
			UserID: int64(21),
		}).Build()

		require.NoError(t, err)

		for i := 0; i < countMsg; i++ {
			_, err = n.Publish(subject, msgBytes)
			require.NoError(t, err)
		}

		receiverCh := make(chan *nats.Msg)
		sub, err := n.Subscribe(subject, func(msg *nats.Msg) {
			receiverCh <- msg
		})
		require.NoError(t, err)

		sub2, err := n2.Subscribe(subject, func(msg *nats.Msg) {
			receiverCh <- msg
		})
		require.NoError(t, err)

		// receive double of the message count beacuse of 2 subscriber
		for i := 0; i < countMsg*2; i++ {
			b := <-receiverCh

			assert.Equal(t, msgBytes, b.Data)
			assert.Equal(t, subject, b.Subject, "test subject")
		}

		_ = sub.Unsubscribe()
		_ = sub2.Unsubscribe()
	})

	t.Run("subscribe NatsEventAuditLogMessage", func(t *testing.T) {
		n, err := NewNATSConnection(defaultURL)
		require.NoError(t, err)

		n2, err := NewNATSConnection(defaultURL)
		require.NoError(t, err)

		defer SafeClose(n)
		defer SafeClose(n2)

		streamConf := &nats.StreamConfig{
			Name:     "STREAM_NAME_AUDIT",
			Subjects: []string{"STREAM_NAME_AUDIT.*"},
			Storage:  nats.FileStorage,
		}

		_, err = n.AddStream(streamConf)
		require.NoError(t, err)

		countMsg := 10
		subject := "STREAM_NAME_AUDIT.TEST_NATS_EVENT_AUDIT_LOG_MESSAGE"

		type User struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		}

		oldData := User{
			ID:   int64(123),
			Name: "test name",
		}

		newData := User{
			ID:   int64(123),
			Name: "new test name",
		}

		byteOldData, err := json.Marshal(oldData)
		require.NoError(t, err)
		byteNewData, err := json.Marshal(newData)
		require.NoError(t, err)

		createdAt, err := time.Parse("2006-01-02", "2020-01-29")
		require.NoError(t, err)

		msg := &NatsEventAuditLogMessage{
			ServiceName:    "test-audit",
			UserID:         123,
			AuditableType:  "user",
			AuditableID:    "123",
			Action:         "update",
			AuditedChanges: string(byteNewData),
			OldData:        string(byteOldData),
			NewData:        string(byteNewData),
			CreatedAt:      createdAt,
			Error:          nil,
		}
		msgBytes, err := msg.Build()
		require.NoError(t, err)

		for i := 0; i < countMsg; i++ {
			_, err = n.Publish(subject, msgBytes)
			require.NoError(t, err)

		}

		receiverCh := make(chan *nats.Msg)
		sub, err := n.Subscribe(subject, func(msg *nats.Msg) {
			receiverCh <- msg
		})
		require.NoError(t, err)

		sub2, err := n2.Subscribe(subject, func(msg *nats.Msg) {
			receiverCh <- msg
		})
		require.NoError(t, err)

		for i := 0; i < countMsg*2; i++ {
			b := <-receiverCh

			assert.Equal(t, msgBytes, b.Data)
			assert.Equal(t, subject, b.Subject, "test subject")
		}

		_ = sub.Unsubscribe()
		_ = sub2.Unsubscribe()
	})
}

func TestAddStream(t *testing.T) {
	n, err := NewNATSConnection(defaultURL)
	require.NoError(t, err)

	defer SafeClose(n)

	streamConf := &nats.StreamConfig{
		Name:     "STREAM_NAMEXX",
		Subjects: []string{"STREAM_EVENTXX.*"},
		MaxAge:   2 * time.Hour,
		Storage:  nats.FileStorage,
	}

	_, err = n.AddStream(streamConf)
	require.NoError(t, err)

	// test update config
	updateConf := &nats.StreamConfig{
		Name:     "STREAM_NAMEXX",
		Subjects: []string{"STREAM_EVENTXX.*"},
		MaxAge:   1 * time.Hour,
		Storage:  nats.FileStorage,
	}

	_, err = n.AddStream(updateConf)
	require.NoError(t, err)
}

func TestConsumerInfo(t *testing.T) {
	n, err := NewNATSConnection(defaultURL)
	require.NoError(t, err)
	defer SafeClose(n)

	updateConf := &nats.StreamConfig{
		Name:     "STREAM_NAME",
		Subjects: []string{"STREAM_EVENT.*"},
		MaxAge:   2 * time.Hour,
		Storage:  nats.FileStorage,
	}

	_, err = n.AddStream(updateConf)
	require.NoError(t, err)

	_, err = n.QueueSubscribe("STREAM_EVENT.A", "QUEUEUEUE", func(msg *nats.Msg) {}, nats.Durable("DURABLE_NAME"))
	require.NoError(t, err)

	consumerInfo, err := n.ConsumerInfo("STREAM_NAME", "DURABLE_NAME")
	assert.NotNil(t, consumerInfo)
	require.NoError(t, err)

	consumerInfo2, err := n.ConsumerInfo("STREAM_NAME", "DURABLE_NAME2")
	require.Error(t, err)
	assert.Nil(t, consumerInfo2)
	assert.Equal(t, nats.ErrConsumerNotFound, err)
}
