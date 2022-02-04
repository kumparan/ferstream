package ferstream

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

const defaultURL = "nats://127.0.0.1:4222"

func RunBasicJetStreamServer() *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.JetStream = true
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
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		n.Close()
	}()

	streamConf := &nats.StreamConfig{
		Name:     "STREAM_NAME",
		Subjects: []string{"STREAM_EVENT.*"},
		MaxAge:   1 * time.Hour,
		Storage:  nats.FileStorage,
	}

	_, err = n.AddStream(streamConf)
	if err != nil {
		t.Fatal(err)
	}

	_, err = n.Publish("STREAM_EVENT.TEST", []byte("test"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestQueueSubscribe(t *testing.T) {
	n, err := NewNATSConnection(defaultURL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		n.Close()
	}()

	streamConf := &nats.StreamConfig{
		Name:     "STREAM_NAME_ANOTHER",
		Subjects: []string{"STREAM_EVENT_ANOTHER.*"},
		MaxAge:   1 * time.Hour,
		Storage:  nats.FileStorage,
	}

	_, err = n.AddStream(streamConf)
	if err != nil {
		t.Fatal(err)
	}

	countMsg := 10
	subject := "STREAM_EVENT_ANOTHER.TEST"
	queue := "test_queue_group"

	msgBytes, err := NewNatsEventMessage().WithEvent(&NatsEvent{
		ID:     int64(1232),
		UserID: int64(21),
	}).Build()

	assert.NoError(t, err)

	for i := 0; i < countMsg; i++ {
		_, err = n.Publish(subject, msgBytes)
		if err != nil {
			t.Fatal(err)
		}
	}

	receiverCh := make(chan *nats.Msg)
	sub, err := n.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		receiverCh <- msg
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < countMsg; i++ {
		b := <-receiverCh

		assert.Equal(t, msgBytes, b.Data)
		assert.Equal(t, subject, b.Subject, "test subject")
	}

	_ = sub.Unsubscribe()
}

func TestAddStream(t *testing.T) {
	n, err := NewNATSConnection(defaultURL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		n.Close()
	}()

	streamConf := &nats.StreamConfig{
		Name:     "STREAM_NAMEXX",
		Subjects: []string{"STREAM_EVENTXX.*"},
		MaxAge:   2 * time.Hour,
		Storage:  nats.FileStorage,
	}

	_, err = n.AddStream(streamConf)

	if err != nil {
		log.Fatal(err)
	}

	// test update config
	updateConf := &nats.StreamConfig{
		Name:     "STREAM_NAMEXX",
		Subjects: []string{"STREAM_EVENTXX.*"},
		MaxAge:   2 * time.Hour,
		Storage:  nats.FileStorage,
	}

	_, err = n.AddStream(updateConf)

	if err != nil {
		t.Fatal(err)
	}
}
