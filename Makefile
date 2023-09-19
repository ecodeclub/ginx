.PHONY: mock
mock:
	@mockgen -source=ratelimit/types.go -package=limitmocks -destination=ratelimit/mocks/ratelimit.mock.go
	@go mod tidy