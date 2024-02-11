.PHONY:	bench
bench:
	@go test -bench=. -benchmem  ./...

.PHONY:	ut
ut:
	@go test -tags=goexperiment.arenas -race ./...

.PHONY:	setup
setup:
	@sh ./.script/setup.sh

.PHONY:	fmt
fmt:
	@sh ./.script/goimports.sh

.PHONY:	lint
lint:
	@golangci-lint run -c .golangci.yml

.PHONY: tidy
tidy:
	@go mod tidy -v

.PHONY: check
check:
	@$(MAKE) fmt
	@$(MAKE) tidy

# e2e 测试
.PHONY: e2e
e2e:
	sh ./.script/integrate_test.sh

.PHONY: e2e_up
e2e_up:
	docker compose -f .script/integration_test_compose.yml up -d

.PHONY: e2e_down
e2e_down:
	docker compose -f .script/integration_test_compose.yml down
mock:
	mockgen -copyright_file=.license_header -package=mocks -destination=internal/mocks/pipeline.mock.go github.com/redis/go-redis/v9 Pipeliner
	mockgen -copyright_file=.license_header -source=session/types.go -package=session -destination=session/provider.mock_test.go Provider