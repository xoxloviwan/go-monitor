all: buildserver buildagent

buildserver:
	GOOS=windows go build -o bin/server.exe cmd/server/main.go

buildagent:
	GOOS=windows go build -o bin/agent.exe cmd/agent/main.go

test:
	go test ./...

mock:
	mockgen -destination ./internal/api/mock/mock_store.go github.com/xoxloviwan/go-monitor/internal/api ReaderWriter

lint:
	go build -o bin/multichecker.exe cmd/staticlint/main.go

GOBIN ?= $$(go env GOPATH)/bin

.PHONY: install-go-test-coverage
install-go-test-coverage:
	go install github.com/vladopajic/go-test-coverage/v2@v2.10.1

.PHONY: check-coverage
check-coverage: install-go-test-coverage
	go test ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...
	go tool cover -func ./cover.out
	${GOBIN}/go-test-coverage --config=./.testcoverage.yml

coverv:
	go tool cover -html cover.out