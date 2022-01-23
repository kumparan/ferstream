package ferstream

import (
	"os"
	"testing"
	"time"

	"github.com/kumparan/tapao"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

const defaultURL = "nats://127.0.0.1:4222"

// DefaultTestOptions are default options for the unit tests.
var DefaultTestOptions = server.Options{
	Host:                  "127.0.0.1",
	Port:                  4222,
	NoLog:                 false,
	NoSigs:                true,
	MaxControlLine:        4096,
	DisableShortFirstPing: true,
	JetStream:             true,
}

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

	_, err = n.Publish("STREAM_EVENT.TEST", []byte("test"), streamConf)
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

	type msg struct {
		Data int64 `json:"data"`
	}

	streamConf := &nats.StreamConfig{
		Name:     "STREAM_NAME_ANOTHER",
		Subjects: []string{"STREAM_EVENT_ANOTHER.*"},
		MaxAge:   1 * time.Hour,
		Storage:  nats.FileStorage,
	}

	countMsg := 10
	subject := "STREAM_EVENT_ANOTHER.TEST"
	queue := "test_queue_group"

	for i := 0; i < countMsg; i++ {
		ms := &msg{
			Data: int64(1554775372665126857),
		}
		msgBytes, err := tapao.Marshal(ms)
		if err != nil {
			t.Fatal(err)
		}

		_, err = n.Publish(subject, msgBytes, streamConf)
		if err != nil {
			t.Fatal(err)
		}
	}

	receiverCh := make(chan []byte, countMsg)
	sub, err := n.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		receiverCh <- msg.Data
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < countMsg; i++ {
		b := <-receiverCh
		msg := new(msg)

		err = tapao.Unmarshal(b, msg)
		if err != nil {
			t.Fatal(err)
		}

		if msg.Data != int64(1554775372665126857) {
			t.Fatal("error")
		}
	}

	_ = sub.Unsubscribe()
}
