mock/mock_jetstream.go:
	mockgen -destination=mock/mock_jetstream.go -package=mock github.com/kumparan/ferstream JetStream

mockgen: mock/mock_jetstream.go

test: lint test-only

test-only:
	go test ./ -v --cover -timeout 60s

lint: check-cognitive-complexity
	golangci-lint run --print-issued-lines=false --exclude-use-default=false --enable=revive --enable=goimports  --enable=unconvert --fix --enable=gosec --timeout=10m

check-cognitive-complexity:
	find . -type f -name '*.go' -not -name "*.pb.go" -not -name "mock*.go" -not -name "*_test.go" \
      -exec gocognit -over 15 {} +

.PHONY: test test-only lint check-cognitive-complexity mockgen