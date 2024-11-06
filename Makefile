DATEBIN = "/c/Program Files/Git/usr/bin/date.exe"
GITBIN = "/c/Program Files/Git/cmd/git.exe"
VERSION := v1.0.1
COMMIT := $$($(GITBIN) rev-parse --short HEAD)
DATE := $$($(DATEBIN) -I)
LDFLAGS := -X main.buildVersion=$(VERSION) -X main.buildDate=$(DATE) -X main.buildCommit=$(COMMIT)

all: buildserver buildagent

buildserver:
	GOOS=windows go build -o bin/server.exe -ldflags "$(LDFLAGS)" cmd/server/main.go

buildagent:
	GOOS=windows go build -o bin/agent.exe -ldflags "$(LDFLAGS)" cmd/agent/main.go

test:
	go test ./...

lint:
	go build -o bin/multichecker.exe cmd/staticlint/main.go

GOBIN ?= $$(go env GOPATH)/bin

.PHONY: install-go-test-coverage
install-go-test-coverage:
	go install github.com/vladopajic/go-test-coverage/v2@v2.10.1

.PHONY: cover
cover: install-go-test-coverage
	go test ./... -coverprofile=./cover.out -coverpkg=./...
	go tool cover -func ./cover.out
	${GOBIN}/go-test-coverage --config=./.testcoverage.yml

coverv:
	go tool cover -html cover.out