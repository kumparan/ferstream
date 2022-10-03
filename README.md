# ferstream
Ferstream is [Go](https://go.dev/) client for [nats jetstream](https://docs.nats.io/nats-concepts/jetstream) created based on [nats.io go client](https://github.com/nats-io/nats.go).
Ferstream is wrapping the functionality of nats.io go client to simplify the usage and also standarize the structure of the message so the publisher and consumer are aware of the message form and can easily be handled.

## Usage
- **Create Nats Connection**
```go
config := []nats.Option{
    nats.RetryOnFailedReconnect(true),
    nats.ReconnectWait(time.Second * 5),
}
conn, _ := ferstream.NewNatsConnection("nats://localhost:4222", conf...)
```
You can specify the config depending of your requirements. The available config can be seen on [nats.io go client](https://github.com/nats-io/nats.go).

- **Create The Client Implementation**  
As a publisher and subscriber you need to create the implementation of [`JetStramRegistrar`](https://github.com/kumparan/ferstream/blob/306533c24b2d3b05297d260872af429651aea21d/jetstream.go#L28) and [`StreamRegistrar`](https://github.com/kumparan/ferstream/blob/306533c24b2d3b05297d260872af429651aea21d/jetstream.go#L33) interface. For subscriber, you also need to implement [`Subscriber`](https://github.com/kumparan/ferstream/blob/306533c24b2d3b05297d260872af429651aea21d/jetstream.go#L38) interface.
- Publisher Implementation Example
```go
type JSPublisher struct {
	js ferstream.JetStream
}

func (p *JSPublisher) RegisterNATSJetStream(js ferstream.JetStream) {
	p.js = js
}

func (p *JSPublisher) InitStream() error {
	// you can do anything you need here before adding stream
	_, err := p.js.AddStream(&nats.StreamConfig{
	    Name: "YOUR_STREAM_NAME",
		Subject: []string{"SUBJECT"},
		StreamMaxAge: time.Hour * 24,
		Storage: nats.FileStorage,
	})
	if err != nil {
	    log.Fatal(err)	
	}
}
```
- Subscriber Implementation Example
```go
type JSSubscriber struct {
	js ferstream.JetStream
}

func (s *JSSubscriber) RegisterNATSJetStream(js ferstream.JetStream) {
	p.js = js
}

func (s *JSSubscriber) InitStream() error {
	// you can do anything you need here before adding stream
	_, err := p.js.AddStream(&nats.StreamConfig{
	    Name: "YOUR_STREAM_NAME",
		Subject: []string{"SUBJECT"},
		StreamMaxAge: time.Hour * 24,
		Storage: nats.FileStorage,
	})
	if err != nil {
	    log.Fatal(err)	
	}
}

func (s *JSSubscriber) SubscribeJetStreamEvent() error {
	var (
		retryAttempts = 3
		retryInterval = time.Second * 5
	)
	msgHandler := func (payload ferstream.MessageParser) error {
            msg, _ := payload.(*ferstream.NatsEventMessage)
			// do something with the message
            return nil
	}
	errHandler := func (payload ferstream.MessageParser) error {
		// something you want to do if the msgHandler returns an error
	}
	natsSubOpts := []nats.SubOpt{nats.ManualAck(), nats.Durable("your-durable-id")}
	_, err := s.js.QueueSubscribe("SUBJECT", "your-queue-group", 
	    ferstream.NewNATSMessageHandler(new(ferstream.NatsEventMessage), retryAttempts, retryInterval, msgHandler, errHandler, natsSubOpts)
            natsSubOpts...
        )
	if err != nil {
		log.Fatal(err)
	}   
}
```
- **Register Clients**
```go
publisher := &JSPublisher{}
subscriber := &JSSubscriber{}
jsRegistrar := []ferstream.JetStreamRegistrar{publisher, subscriber}
err := ferstream.RegisterJetStreamClient(conn, jsRegistrar)
```
- **Publish Event**
```go
func (p *JSPublisher) PublishEvent() {
	// mock payload
	type Job struct {
		Id int64
		Name string
	}    
	j := &Job{Id:1, Name: "job"}
	userID := int64(1)
	eventMsg := ferstream.NewNatsEventMessage().WithEvent(
		&ferstream.NatsEvent{
			ID:     j.Id,
			UserID: 1,
		})
	eventMsg = eventMsg.WithBody(j)
	msgByte, _ := eventMsg.Build()
	_ := p.js.Publish("EVENT-SUBJECT", msgByte)
}
```