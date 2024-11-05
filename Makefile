all: buildserver buildagent

buildserver:
	GOOS=windows go build -o bin/server.exe cmd/server/main.go

buildagent:
	GOOS=windows go build -o bin/agent.exe cmd/agent/main.go

test:
	go test ./...

cover:
	go test ./... -coverprofile cover && go tool cover -func cover

coverv:
	go tool cover -html cover

mock:
	mockgen -destination ./internal/api/mock/mock_store.go github.com/xoxloviwan/go-monitor/internal/api ReaderWriter

lint:
	go build -o bin/multichecker.exe cmd/staticlint/main.go