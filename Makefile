.PHONY: mock
mock:
	@mockgen -source=internal/ratelimit/types.go -package=limitmocks -destination=internal/ratelimit/mocks/ratelimit.mock.go
	@go mod tidy