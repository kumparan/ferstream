mock/mock_jetstream.go:
	mockgen -destination=mock/mock_jetstream.go -package=mock github.com/kumparan/ferstream JetStream

mockgen: mock/mock_jetstream.go