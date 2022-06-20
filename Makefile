build:
	go build cmd/discord-tts/discord-tts.go
lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.39.0 golangci-lint run -v
